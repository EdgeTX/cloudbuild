package artifactory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/edgetx/cloudbuild/buildlogs"
	"github.com/edgetx/cloudbuild/firmware"
	"github.com/edgetx/cloudbuild/source"
	"github.com/edgetx/cloudbuild/storage"
	"github.com/pkg/errors"
)

type Artifactory struct {
	BuildJobsRepository BuildJobsRepository
	ArtifactStorage     storage.Handler
	SourceRepository    string
	BuildContainerImage string
}

func New(
	buildJobsRepository BuildJobsRepository,
	artifactStorage storage.Handler,
	buildContainerImage string,
	sourceRepository string,
) *Artifactory {
	return &Artifactory{
		BuildJobsRepository: buildJobsRepository,
		ArtifactStorage:     artifactStorage,
		BuildContainerImage: buildContainerImage,
		SourceRepository:    sourceRepository,
	}
}

func (artifactory *Artifactory) GetBuild(commitHash string, flags []firmware.BuildFlag) (*BuildJobDto, error) {
	buildJob, err := artifactory.BuildJobsRepository.Get(commitHash, flags)
	if err != nil {
		return nil, err
	}
	if err == nil && buildJob == nil {
		return nil, errors.New("not found")
	}

	return BuildJobDtoFromModel(buildJob)
}

func (artifactory *Artifactory) CreateBuildJob(
	requesterIP string,
	commitHash string,
	flags []firmware.BuildFlag,
) (*BuildJobDto, error) {
	artifactModel, err := artifactory.BuildJobsRepository.Get(commitHash, flags)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check for existing build")
	}

	if err == nil && artifactModel != nil {
		// forcibly restart completely failed build
		if artifactModel.Status == BuildError {
			artifactModel.AuditLogs = append(artifactModel.AuditLogs, AuditLogModel{
				RequestIP: requesterIP,
				From:      BuildError,
				To:        WaitingForBuild,
			})
			err := artifactory.BuildJobsRepository.Save(&BuildJobModel{
				ID:        artifactModel.ID,
				Status:    WaitingForBuild,
				AuditLogs: artifactModel.AuditLogs,
			})
			if err != nil {
				return nil, err
			}
			artifactModel, err = artifactory.BuildJobsRepository.Get(commitHash, flags)
			if err != nil {
				return nil, err
			}
			return BuildJobDtoFromModel(artifactModel)
		}
		return BuildJobDtoFromModel(artifactModel)
	}

	buildFlagsJSON, err := json.Marshal(flags)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal build flags")
	}

	newArtifactModel, err := artifactory.BuildJobsRepository.Create(BuildJobModel{
		Status:         WaitingForBuild,
		CommitHash:     commitHash,
		ContainerImage: artifactory.BuildContainerImage,
		BuildFlags:     buildFlagsJSON,
		BuildFlagsHash: HashBuildFlags(flags),
		AuditLogs: []AuditLogModel{
			{
				RequestIP: requesterIP,
				From:      VoidStatus,
				To:        WaitingForBuild,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return BuildJobDtoFromModel(newArtifactModel)
}

func (artifactory *Artifactory) Build(
	ctx context.Context,
	build *BuildJobModel,
	recorder *buildlogs.Recorder,
	sources source.Downloader,
	builder firmware.Builder,
) (*BuildJobModel, error) {
	now := time.Now()
	build.BuildAttempts += 1
	build.BuildStartedAt = now
	build.AuditLogs = append(build.AuditLogs, AuditLogModel{
		From:      WaitingForBuild,
		To:        BuildInProgress,
		CreatedAt: now,
	})
	onBuildFailure := func(err error, build *BuildJobModel) (*BuildJobModel, error) {
		build.BuildEndedAt = time.Now()
		newStatus := WaitingForBuild
		if build.BuildAttempts >= MaxBuildAttempts {
			newStatus = BuildError
		}
		build.Status = newStatus

		build.AuditLogs = append(build.AuditLogs, AuditLogModel{
			From:      BuildInProgress,
			To:        newStatus,
			CreatedAt: time.Now(),
			StdOut:    recorder.Logs(),
		})

		revertErr := artifactory.BuildJobsRepository.Save(build)
		if revertErr != nil {
			return build, errors.Errorf("failed to process build: %s and failed to update job: %s", err, revertErr)
		}
		return build, err
	}

	err := sources.Download(ctx, artifactory.SourceRepository, build.CommitHash)
	if err != nil {
		return onBuildFailure(err, build)
	}

	var flags []firmware.BuildFlag
	err = json.Unmarshal([]byte(build.BuildFlags.String()), &flags)
	if err != nil {
		return onBuildFailure(err, build)
	}

	firmwareBin, err := builder.Build(ctx, build.ContainerImage, flags)
	if err != nil {
		return onBuildFailure(err, build)
	}

	fileName := fmt.Sprintf("%s-%s.bin", build.CommitHash, build.BuildFlagsHash)
	firmwareURL, err := artifactory.ArtifactStorage.Upload(ctx, firmwareBin, fileName)
	if err != nil {
		return onBuildFailure(err, build)
	}

	build.Status = BuildSuccess
	build.Artifacts = append(build.Artifacts, ArtifactModel{
		Slug:        "firmware",
		DownloadURL: firmwareURL.String(),
	})
	build.AuditLogs = append(build.AuditLogs, AuditLogModel{
		From:      BuildInProgress,
		To:        BuildSuccess,
		CreatedAt: time.Now(),
		StdOut:    recorder.Logs(),
	})

	err = artifactory.BuildJobsRepository.Save(build)
	if err != nil {
		return onBuildFailure(err, build)
	}

	return build, nil
}

func (artifactory *Artifactory) ProcessNextBuildJob(
	ctx context.Context,
	recorder *buildlogs.Recorder,
	sources source.Downloader,
	builder firmware.Builder,
) (*BuildJobModel, error) {
	err := artifactory.BuildJobsRepository.TimeoutBuilds(time.Minute * 15)
	if err != nil {
		return nil, err
	}

	build, err := artifactory.BuildJobsRepository.ReservePendingBuild()
	if err != nil {
		return nil, err
	}

	if err == nil && build == nil {
		return nil, nil
	}

	return artifactory.Build(ctx, build, recorder, sources, builder)
}
