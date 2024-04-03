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
	"github.com/edgetx/cloudbuild/database"
	"github.com/edgetx/cloudbuild/firmware"
	"github.com/edgetx/cloudbuild/source"
	"github.com/edgetx/cloudbuild/storage"
	"github.com/prometheus/client_golang/prometheus"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
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

func (artifactory *Artifactory) ListJobs(query *JobQuery) (*database.Pagination, error) {
	res, err := artifactory.BuildJobsRepository.List(query)
	if err != nil {
		return nil, err
	}

	res.Rows, err = BuildJobsDtoFromInterface(res.Rows, artifactory.PrefixURL)
	return res, err
}

func (artifactory *Artifactory) DeleteJob(id string) error {
	uid, err := uuid.FromString(id)
	if err != nil {
		return err
	}
	return artifactory.BuildJobsRepository.Delete(uid)
}

func (artifactory *Artifactory) GetBuild(request *BuildRequest) (*BuildJobDto, error) {
	buildJob, err := artifactory.BuildJobsRepository.Get(request)
	if err != nil {
		return nil, err
	}
	if buildJob == nil {
		return nil, ErrBuildNotFound
	}

	return BuildJobDtoFromModel(buildJob, artifactory.PrefixURL)
}

func (artifactory *Artifactory) GetLogs(jobID string) (*[]AuditLogDto, error) {
	uid, err := uuid.FromString(jobID)
	if err != nil {
		return nil, err
	}

	logs, err := artifactory.BuildJobsRepository.GetLogs(uid)
	if err != nil {
		return nil, err
	}

	auditLogs := make([]AuditLogDto, 0)
	for i := range *logs {
		auditLogs = append(auditLogs, AuditLogDtoFromModel(&(*logs)[i]))
	}

	return &auditLogs, nil
}

func (artifactory *Artifactory) restartFailedJob(
	requesterIP string, job *BuildJobModel,
) (*BuildJobDto, error) {
	job.AuditLogs = append(job.AuditLogs, AuditLogModel{
		RequestIP: requesterIP,
		From:      BuildError,
		To:        WaitingForBuild,
	})
	job.Status = WaitingForBuild
	job.BuildAttempts = 0

	err := artifactory.BuildJobsRepository.Save(job)
	if err != nil {
		return nil, err
	}

	return BuildJobDtoFromModel(job, artifactory.PrefixURL)
}

func (artifactory *Artifactory) CreateBuildJob(
	requesterIP string, request *BuildRequest,
) (*BuildJobDto, error) {
	job, err := artifactory.BuildJobsRepository.Get(request)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing build: %w", err)
	}

	if job != nil {
		// forcibly restart completely failed build
		if job.Status == BuildError {
			return artifactory.restartFailedJob(requesterIP, job)
		}
		return BuildJobDtoFromModel(job, artifactory.PrefixURL)
	}

	buildFlags, err := request.GetBuildFlags()
	if err != nil {
		return nil, fmt.Errorf("failed to get build flags: %w", err)
	}
	optFlagsJSON, err := json.Marshal(request.Flags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal option flags: %w", err)
	}
	buildFlagsJSON, err := json.Marshal(buildFlags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal build flags: %w", err)
	}

	job, err = artifactory.BuildJobsRepository.Create(BuildJobModel{
		Status:         WaitingForBuild,
		CommitRef:      request.Release,
		CommitHash:     request.GetCommitHash(),
		Target:         request.Target,
		Flags:          optFlagsJSON,
		BuildFlags:     buildFlagsJSON,
		ContainerImage: artifactory.BuildContainerImage,
		BuildFlagsHash: request.HashTargetAndFlags(),
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

	return BuildJobDtoFromModel(job, artifactory.PrefixURL)
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

	firmwareBin, err := builder.Build(ctx, build.ContainerImage, build.Target, build.CommitRef, flags)
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
		Size:     (int64)(len(firmwareBin)),
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
		_ = jobsRepo.TimeoutBuilds(MaxBuildDuration)
		time.Sleep(time.Second * 1)
	}
}

func (artifactory *Artifactory) RunMetrics(
	queued, building, failed prometheus.Gauge,
) {
	log.Debugln("Start RunMetrics")
	jobsRepo := artifactory.BuildJobsRepository
	for {
		jobsRepo.UpdateMetrics(queued, building, failed)
		time.Sleep(time.Second * 30)
	}
}
