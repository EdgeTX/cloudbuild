package artifactory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/edgetx/cloudbuild/buildlogs"
	"github.com/edgetx/cloudbuild/config"
	"github.com/edgetx/cloudbuild/firmware"
	"github.com/edgetx/cloudbuild/source"
	"github.com/edgetx/cloudbuild/storage"
)

var (
	ErrNoArtifactStorage = errors.New("missing artifact storage")
	ErrBuildNotFound     = errors.New("build not found")
)

type Artifactory struct {
	BuildJobsRepository BuildJobsRepository
	ArtifactStorage     storage.Handler
	SourceRepository    string
	BuildContainerImage string
	PrefixURL           *url.URL
}

func New(
	buildJobsRepository BuildJobsRepository,
	artifactStorage storage.Handler,
	buildContainerImage string,
	sourceRepository string,
	prefixURL *url.URL,
) *Artifactory {
	return &Artifactory{
		BuildJobsRepository: buildJobsRepository,
		ArtifactStorage:     artifactStorage,
		BuildContainerImage: buildContainerImage,
		SourceRepository:    sourceRepository,
		PrefixURL:           prefixURL,
	}
}

func NewFromConfig(ctx context.Context, c *config.CloudbuildOpts) (*Artifactory, error) {
	buildJobsRepository, err := NewBuildJobsDBRepositoryFromConfig(c)
	if err != nil {
		return nil, err
	}
	artifactStorage := storage.NewFromConfig(ctx, c)
	if artifactStorage == nil {
		return nil, ErrNoArtifactStorage
	}
	prefixURL, err := url.Parse(c.DownloadURL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse PREFIX_URL: %w", err)
	}
	return New(
		buildJobsRepository,
		artifactStorage,
		c.BuildImage,
		c.SourceRepository,
		prefixURL,
	), nil
}

func (artifactory *Artifactory) ListJobs(status string) (*[]BuildJobDto, error) {
	jobs, err := artifactory.BuildJobsRepository.List(status)
	if err != nil {
		return nil, err
	}

	resJobs := make([]BuildJobDto, len(*jobs))
	for i := range *jobs {
		j, err := BuildJobDtoFromModel(&(*jobs)[i], artifactory.PrefixURL)
		if err != nil {
			return nil, err
		}
		resJobs[i] = *j
	}

	return &resJobs, nil
}

func (artifactory *Artifactory) GetBuild(commitHash string, flags []firmware.BuildFlag) (*BuildJobDto, error) {
	buildJob, err := artifactory.BuildJobsRepository.Get(commitHash, flags)
	if err != nil {
		return nil, err
	}
	if err == nil && buildJob == nil {
		return nil, ErrBuildNotFound
	}

	return BuildJobDtoFromModel(buildJob, artifactory.PrefixURL)
}

func (artifactory *Artifactory) CreateBuildJob(
	requesterIP string,
	commitHash string,
	flags []firmware.BuildFlag,
) (*BuildJobDto, error) {
	artifactModel, err := artifactory.BuildJobsRepository.Get(commitHash, flags)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing build: %w", err)
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
			return BuildJobDtoFromModel(artifactModel, artifactory.PrefixURL)
		}
		return BuildJobDtoFromModel(artifactModel, artifactory.PrefixURL)
	}

	buildFlagsJSON, err := json.Marshal(flags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal build flags: %w", err)
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

	return BuildJobDtoFromModel(newArtifactModel, artifactory.PrefixURL)
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
			return build, fmt.Errorf(
				"failed to process build: %w and failed to update job: %w",
				err, revertErr)
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
	err = artifactory.ArtifactStorage.Upload(ctx, firmwareBin, fileName)
	if err != nil {
		return onBuildFailure(err, build)
	}

	build.Status = BuildSuccess
	build.Artifacts = append(build.Artifacts, ArtifactModel{
		Slug:     "firmware",
		Filename: fileName,
	})
	build.AuditLogs = append(build.AuditLogs, AuditLogModel{
		From:      BuildInProgress,
		To:        BuildSuccess,
		CreatedAt: time.Now(),
		StdOut:    recorder.Logs(),
	})
	build.BuildEndedAt = time.Now()

	err = artifactory.BuildJobsRepository.Save(build)
	if err != nil {
		return onBuildFailure(err, build)
	}

	return build, nil
}

func (artifactory *Artifactory) ReservePendingBuild() (*BuildJobModel, error) {
	return artifactory.BuildJobsRepository.ReservePendingBuild()
}

func (artifactory *Artifactory) RunGarbageCollector() {
	jobsRepo := artifactory.BuildJobsRepository
	for {
		jobsRepo.TimeoutBuilds(MaxBuildDuration)
		time.Sleep(time.Second * 1)
	}
}
