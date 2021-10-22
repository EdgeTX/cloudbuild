package artifactory_test

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/edgetx/cloudbuild/artifactory"
	"github.com/edgetx/cloudbuild/buildlogs"
	"github.com/edgetx/cloudbuild/database"
	"github.com/edgetx/cloudbuild/firmware"
	"github.com/edgetx/cloudbuild/storage"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

var (
	commitHash = "8620fe19289c36b574ab68008145a530d589f0fd"
	flags      = []firmware.BuildFlag{
		firmware.NewFlag("DISABLE_COMPANION", "YES"),
		firmware.NewFlag("CMAKE_BUILD_TYPE", "Release"),
		firmware.NewFlag("TRACE_SIMPGMSPACE", "NO"),
		firmware.NewFlag("VERBOSE_CMAKELISTS", "YES"),
		firmware.NewFlag("CMAKE_RULE_MESSAGES", "OFF"),
		firmware.NewFlag("PCB", "X10"),
		firmware.NewFlag("PCBREV", "T16"),
		firmware.NewFlag("INTERNAL_MODULE_MULTI", "ON"),
	}
)

func getDB() (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = database.DropSchema(db)
	if err != nil {
		return nil, err
	}

	err = database.Migrate(db)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func newArtifactory(db *gorm.DB, handler storage.Handler) *artifactory.Artifactory {
	if handler == nil {
		mck := &MockStorage{}
		mck.
			On("Upload", mock.Anything, mock.Anything, mock.Anything).
			Return(url.Parse("http://localhost:3000/firmware.bin"))
		handler = mck
	}

	repo := artifactory.NewBuildJobsDBRepository(db)
	buildImage := os.Getenv("BUILD_IMAGE")
	if len(buildImage) == 0 {
		log.Fatal("BUILD_IMAG env variable is not specified")
	}

	sourceRepository := os.Getenv("SOURCE_REPOSITORY")
	if len(sourceRepository) == 0 {
		log.Fatal("sourceRepository env variable is not specified")
	}

	return artifactory.New(repo, handler, buildImage, sourceRepository)
}

func createBuildModel(
	db *gorm.DB,
	status artifactory.BuildStatus,
	commitHash string,
	flags []firmware.BuildFlag,
) (*artifactory.BuildJobModel, error) {
	repository := artifactory.NewBuildJobsDBRepository(db)
	buildFlagsJSON, err := json.Marshal(flags)
	if err != nil {
		return nil, err
	}
	build := artifactory.BuildJobModel{
		Status:         status,
		CommitHash:     commitHash,
		BuildFlags:     buildFlagsJSON,
		BuildFlagsHash: artifactory.HashBuildFlags(flags),
		Artifacts:      nil,
	}
	if status == artifactory.BuildSuccess {
		build.Artifacts = append(build.Artifacts, artifactory.ArtifactModel{
			Slug:        "firmware",
			DownloadURL: "http://localhost:3000/firmware.bin",
		})
	}
	buildModel, err := repository.Create(build)
	if err != nil {
		return nil, err
	}
	return buildModel, nil
}

func TestCreateBuildJob(t *testing.T) {
	db, err := getDB()
	assert.Nil(t, err)
	art := newArtifactory(db, nil)

	job, err := art.CreateBuildJob("127.0.0.1", commitHash, flags)
	t.Logf("job: %+v err: %s", job, err)

	assert.Nil(t, err)
	assert.NotNil(t, job)

	assert.Equal(t, artifactory.WaitingForBuild, job.Status)
	assert.Equal(t, artifactory.HashBuildFlags(flags), job.BuildFlagsHash)
	assert.Equal(t, len(flags), len(job.BuildFlags))
}

func TestCreatesBuildJobWhenBuildExistsInErrorState(t *testing.T) {
	db, err := getDB()
	assert.Nil(t, err)
	art := newArtifactory(db, nil)

	job, err := art.CreateBuildJob("127.0.0.1", commitHash, flags)
	assert.Nil(t, err)
	repository := artifactory.NewBuildJobsDBRepository(db)
	err = repository.Save(&artifactory.BuildJobModel{
		ID:     uuid.Must(uuid.FromString(job.ID)),
		Status: artifactory.BuildError,
	})
	assert.Nil(t, err)

	job, err = art.CreateBuildJob("127.0.0.1", commitHash, flags)
	assert.Nil(t, err)
	assert.Equal(t, artifactory.WaitingForBuild, job.Status)
}

func TestGetBuildWhenBuildDoesNotExist(t *testing.T) {
	db, err := getDB()
	assert.Nil(t, err)
	art := newArtifactory(db, nil)
	job, err := art.GetBuild(commitHash, flags)
	assert.Nil(t, job)
	assert.NotNil(t, err)
}

func TestGetBuildWhenBuildExists(t *testing.T) {
	db, err := getDB()
	assert.Nil(t, err)
	art := newArtifactory(db, nil)

	model, err := createBuildModel(db, artifactory.BuildSuccess, commitHash, flags)
	assert.Nil(t, err)
	dto, err := art.GetBuild(commitHash, flags)
	assert.Nil(t, err)
	assert.Equal(t, model.ID.String(), dto.ID)
	assert.True(t, len(dto.Artifacts[0].DownloadURL) > 0)
	assert.Equal(t, model.Artifacts[0].DownloadURL, dto.Artifacts[0].DownloadURL)
}

func TestGetBuildWhenInProgress(t *testing.T) {
	db, err := getDB()
	assert.Nil(t, err)
	art := newArtifactory(db, nil)

	model, err := createBuildModel(db, artifactory.BuildInProgress, commitHash, flags)
	assert.Nil(t, err)
	dto, err := art.GetBuild(commitHash, flags)
	assert.Nil(t, err)
	assert.Equal(t, model.ID.String(), dto.ID)
}

func TestGetBuildWhenWaitingForBuild(t *testing.T) {
	db, err := getDB()
	assert.Nil(t, err)
	art := newArtifactory(db, nil)

	model, err := createBuildModel(db, artifactory.WaitingForBuild, commitHash, flags)
	assert.Nil(t, err)
	dto, err := art.GetBuild(commitHash, flags)
	assert.Nil(t, err)
	assert.Equal(t, model.ID.String(), dto.ID)
}

func TestGetBuildWhenBuildIsInError(t *testing.T) {
	db, err := getDB()
	assert.Nil(t, err)
	art := newArtifactory(db, nil)

	model, err := createBuildModel(db, artifactory.BuildError, commitHash, flags)
	assert.Nil(t, err)
	dto, err := art.GetBuild(commitHash, flags)
	assert.Nil(t, err)
	assert.Equal(t, model.ID.String(), dto.ID)
}

func TestProcessNextBuildJobWhenNoneAvailable(t *testing.T) {
	db, err := getDB()
	assert.Nil(t, err)
	art := newArtifactory(db, nil)

	ctx := context.Background()
	recorder := buildlogs.NewRecorder()
	downloader := &MockDownloader{}
	builder := &MockFirmwareBuilder{}
	model, err := art.ProcessNextBuildJob(ctx, recorder, downloader, builder)
	assert.Nil(t, err)
	assert.Nil(t, model)
}

func TestProcessNextBuildJobWhenFailingToDownload(t *testing.T) {
	db, err := getDB()
	assert.Nil(t, err)
	art := newArtifactory(db, nil)

	model1, err := createBuildModel(db, artifactory.WaitingForBuild, commitHash, flags)
	assert.Nil(t, err)

	ctx := context.Background()
	recorder := buildlogs.NewRecorder()
	downloader := &MockDownloader{}
	downloader.
		On("Download", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("failed to download"))
	builder := &MockFirmwareBuilder{}

	model2, err := art.ProcessNextBuildJob(ctx, recorder, downloader, builder)
	assert.Error(t, err, "failed to download")
	downloader.AssertNumberOfCalls(t, "Download", 1)
	assert.Equal(t, model1.ID.String(), model2.ID.String())
	assert.Equal(t, artifactory.WaitingForBuild, model2.Status)
	assert.Equal(t, int64(1), model2.BuildAttempts)
}

func TestProcessNextBuildJobWhenFailingToBuild(t *testing.T) {
	db, err := getDB()
	assert.Nil(t, err)
	art := newArtifactory(db, nil)

	model1, err := createBuildModel(db, artifactory.WaitingForBuild, commitHash, flags)
	assert.Nil(t, err)

	ctx := context.Background()
	recorder := buildlogs.NewRecorder()
	downloader := &MockDownloader{}
	downloader.
		On("Download", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	builder := &MockFirmwareBuilder{}
	builder.
		On("Build", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("failed to build"))
	model2, err := art.ProcessNextBuildJob(ctx, recorder, downloader, builder)
	assert.Error(t, err, "failed to build")
	downloader.AssertNumberOfCalls(t, "Download", 1)
	builder.AssertNumberOfCalls(t, "Build", 1)
	assert.Equal(t, model1.ID.String(), model2.ID.String())
	assert.Equal(t, artifactory.WaitingForBuild, model2.Status)
	assert.Equal(t, int64(1), model2.BuildAttempts)
}

func TestProcessNextBuildJobWhenFailingToUpload(t *testing.T) {
	db, err := getDB()
	assert.Nil(t, err)
	uploader := &MockStorage{}
	uploader.
		On("Upload", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("failed to upload"))
	art := newArtifactory(db, uploader)

	model1, err := createBuildModel(db, artifactory.WaitingForBuild, commitHash, flags)
	assert.Nil(t, err)

	ctx := context.Background()
	recorder := buildlogs.NewRecorder()
	downloader := &MockDownloader{}
	downloader.
		On("Download", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	builder := &MockFirmwareBuilder{}
	builder.
		On("Build", mock.Anything, mock.Anything, mock.Anything).
		Return([]byte("test"), nil)
	model2, err := art.ProcessNextBuildJob(ctx, recorder, downloader, builder)
	assert.Error(t, err, "failed to upload")

	assert.Equal(t, model1.ID.String(), model2.ID.String())
	assert.Equal(t, artifactory.WaitingForBuild, model2.Status)
	assert.Equal(t, int64(1), model2.BuildAttempts)
}

func TestSuccessfulBuildJobFlow(t *testing.T) {
	db, err := getDB()
	assert.Nil(t, err)
	art := newArtifactory(db, nil)

	model1, err := createBuildModel(db, artifactory.WaitingForBuild, commitHash, flags)
	assert.Nil(t, err)

	ctx := context.Background()
	recorder := buildlogs.NewRecorder()
	downloader := &MockDownloader{}
	downloader.
		On("Download", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	builder := &MockFirmwareBuilder{}
	builder.
		On("Build", mock.Anything, mock.Anything, mock.Anything).
		Return([]byte("edgetx"), nil)
	model2, err := art.ProcessNextBuildJob(ctx, recorder, downloader, builder)
	assert.Nil(t, err)

	assert.Equal(t, model1.ID.String(), model2.ID.String())
	assert.Equal(t, artifactory.BuildSuccess, model2.Status)
	assert.Equal(t, int64(1), model2.BuildAttempts)
	assert.Equal(t, model2.Artifacts[0].DownloadURL, "http://localhost:3000/firmware.bin")
}

func TestJobGoesToErrorAfterTooManyFailures(t *testing.T) {
	db, err := getDB()
	assert.Nil(t, err)
	art := newArtifactory(db, nil)

	model1, err := createBuildModel(db, artifactory.WaitingForBuild, commitHash, flags)
	assert.Nil(t, err)

	model1.BuildAttempts = 10

	repository := artifactory.NewBuildJobsDBRepository(db)
	err = repository.Save(model1)
	assert.Nil(t, err)

	ctx := context.Background()
	recorder := buildlogs.NewRecorder()
	downloader := &MockDownloader{}
	downloader.
		On("Download", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("download error"))
	builder := &MockFirmwareBuilder{}
	model2, err := art.ProcessNextBuildJob(ctx, recorder, downloader, builder)
	assert.Error(t, err, "download error")

	assert.Equal(t, model1.ID.String(), model2.ID.String())
	assert.Equal(t, artifactory.BuildError, model2.Status)
	assert.Equal(t, model1.BuildAttempts+1, model2.BuildAttempts)
}

func TestBackoffDurationForFailedBuild(t *testing.T) {
	db, err := getDB()
	assert.Nil(t, err)
	art := newArtifactory(db, nil)

	model1, err := createBuildModel(db, artifactory.WaitingForBuild, commitHash, flags)
	assert.Nil(t, err)

	model1.BuildAttempts = 1
	model1.BuildEndedAt = time.Now()

	repository := artifactory.NewBuildJobsDBRepository(db)
	err = repository.Save(model1)
	assert.Nil(t, err)

	ctx := context.Background()
	recorder := buildlogs.NewRecorder()
	downloader := &MockDownloader{}
	builder := &MockFirmwareBuilder{}

	model2, err := art.ProcessNextBuildJob(ctx, recorder, downloader, builder)
	assert.Nil(t, model2)
	assert.Nil(t, err)

	model3, err := repository.FindByID(model1.ID)
	assert.Nil(t, err)
	assert.Equal(t, model1.BuildAttempts, model3.BuildAttempts)
}
