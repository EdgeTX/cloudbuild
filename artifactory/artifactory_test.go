package artifactory_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/edgetx/cloudbuild/artifactory"
	"github.com/edgetx/cloudbuild/buildlogs"
	"github.com/edgetx/cloudbuild/config"
	"github.com/edgetx/cloudbuild/database"
	"github.com/edgetx/cloudbuild/storage"
	"github.com/edgetx/cloudbuild/targets"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
	err := targets.ReadTargetsDefFromBytes([]byte(targetsJSON))
	if err != nil {
		panic(err)
	}
}

var (
	targetsJSON = `{
	  "releases": { "v1.2.3": { "sha": "3ca63cbb9bb7fe14c22e0349b668900f125e2d09" }},
	  "flags": {
	    "language": {
	      "build_flag": "TRANSLATIONS",
	      "values": [ "CZ", "FR", "FI" ]
	    },
	    "foo": {
	      "build_flag": "FOO",
	      "values": [ "BAR" ]
	    }
	  },
	  "tags": {
	    "colorlcd": {
	      "flags": {
		"language": {
		  "values": [ "CN", "JP", "TW" ]
		}
	      }
	    }
	  },
	  "targets": {
	    "mydreamradio": {
	      "description": "Acme Dream Radio",
	      "build_flags": {
		"PCB": "ACME",
		"PCBREV": "DREAM"
	      }
	    }
	  }
	}`

	commitRef  = "v1.2.3"
	commitHash = "3ca63cbb9bb7fe14c22e0349b668900f125e2d09"
	target     = "mydreamradio"
	flags      = []artifactory.OptionFlag{
		{Name: "language", Value: "FR"},
		{Name: "foo", Value: "BAR"},
	}
	request = &artifactory.BuildRequest{
		Release: commitRef,
		Target:  target,
		Flags:   flags,
	}
	testCfg *config.CloudbuildOpts
	testDB  *gorm.DB
)

func resetDB(dsn string) error {
	if err := database.DropSchema(dsn); err != nil {
		return err
	}
	if err := database.Migrate(dsn); err != nil {
		return err
	}
	return nil
}

func setupTestDB(c *config.CloudbuildOpts) (*gorm.DB, error) {
	return database.New(c.DatabaseDSN)
}

func newArtifactory(db *gorm.DB, handler storage.Handler) *artifactory.Artifactory {
	if handler == nil {
		mck := &MockStorage{}
		mck.
			On("Upload", mock.Anything, mock.Anything, mock.Anything).
			Return(nil)
		handler = mck
	}

	repo := artifactory.NewBuildJobsDBRepository(db)
	buildImage := testCfg.BuildImage
	if len(buildImage) == 0 {
		log.Fatal("BUILD_IMAGE env variable is not specified")
	}

	sourceRepository := testCfg.SourceRepository
	if len(sourceRepository) == 0 {
		log.Fatal("sourceRepository env variable is not specified")
	}

	return artifactory.New(repo, handler, buildImage, sourceRepository, &url.URL{})
}

func createBuildModel(
	db *gorm.DB,
	status artifactory.BuildStatus,
	request *artifactory.BuildRequest,
) (*artifactory.BuildJobModel, error) {
	repository := artifactory.NewBuildJobsDBRepository(db)
	buildFlagsJSON, err := json.Marshal(flags)
	if err != nil {
		return nil, err
	}
	build := artifactory.BuildJobModel{
		Status:         status,
		CommitHash:     commitHash,
		CommitRef:      commitRef,
		Target:         target,
		BuildFlags:     buildFlagsJSON,
		BuildFlagsHash: request.HashTargetAndFlags(),
		Artifacts:      nil,
	}
	if status == artifactory.BuildSuccess {
		build.Artifacts = append(build.Artifacts, artifactory.ArtifactModel{
			Slug:     "firmware",
			Filename: "firmware.bin",
		})
	}
	buildModel, err := repository.Create(build)
	if err != nil {
		return nil, err
	}
	return buildModel, nil
}

func TestCreateBuildJob(t *testing.T) {
	resetDB(testCfg.DatabaseDSN) //nolint:errcheck
	art := newArtifactory(testDB, nil)

	job, err := art.CreateBuildJob("127.0.0.1", request)
	t.Logf("job: %+v err: %s", job, err)

	assert.Nil(t, err)
	assert.NotNil(t, job)

	assert.Equal(t, artifactory.WaitingForBuild, job.Status)
	assert.Equal(t, request.HashTargetAndFlags(), job.BuildFlagsHash)
	// 2 from target + 2 additional flags
	assert.Equal(t, len(flags)+2, len(job.BuildFlags))
}

func TestCreatesBuildJobWhenBuildExistsInErrorState(t *testing.T) {
	resetDB(testCfg.DatabaseDSN) //nolint:errcheck
	art := newArtifactory(testDB, nil)

	job, err := art.CreateBuildJob("127.0.0.1", request)
	assert.Nil(t, err)
	repository := artifactory.NewBuildJobsDBRepository(testDB)
	err = repository.Save(&artifactory.BuildJobModel{
		ID:     uuid.Must(uuid.FromString(job.ID)),
		Status: artifactory.BuildError,
	})
	assert.Nil(t, err)

	job, err = art.CreateBuildJob("127.0.0.1", request)
	assert.Nil(t, err)
	assert.NotNil(t, job)
	if job != nil {
		assert.Equal(t, artifactory.WaitingForBuild, job.Status)
	}
}

func TestGetBuildWhenBuildDoesNotExist(t *testing.T) {
	resetDB(testCfg.DatabaseDSN) //nolint:errcheck
	art := newArtifactory(testDB, nil)
	job, err := art.GetBuild(request)
	assert.Nil(t, job)
	assert.NotNil(t, err)
}

func TestGetBuildWhenBuildExists(t *testing.T) {
	resetDB(testCfg.DatabaseDSN) //nolint:errcheck
	art := newArtifactory(testDB, nil)
	model, err := createBuildModel(testDB, artifactory.BuildSuccess, request)
	assert.Nil(t, err)
	dto, err := art.GetBuild(request)
	assert.Nil(t, err)
	assert.NotNil(t, dto)
	if dto != nil {
		assert.Equal(t, model.ID.String(), dto.ID)
		assert.True(t, len(dto.Artifacts[0].DownloadURL) > 0)
		assert.True(t,
			strings.Contains(
				dto.Artifacts[0].DownloadURL,
				model.Artifacts[0].Filename,
			),
		)
	}
}

func TestGetBuildWhenInProgress(t *testing.T) {
	resetDB(testCfg.DatabaseDSN) //nolint:errcheck
	art := newArtifactory(testDB, nil)
	model, err := createBuildModel(testDB, artifactory.BuildInProgress, request)
	assert.Nil(t, err)
	dto, err := art.GetBuild(request)
	assert.Nil(t, err)
	assert.NotNil(t, dto)
	if dto != nil {
		assert.Equal(t, model.ID.String(), dto.ID)
	}
}

func TestGetBuildWhenWaitingForBuild(t *testing.T) {
	resetDB(testCfg.DatabaseDSN) //nolint:errcheck
	art := newArtifactory(testDB, nil)
	model, err := createBuildModel(testDB, artifactory.WaitingForBuild, request)
	assert.Nil(t, err)
	dto, err := art.GetBuild(request)
	assert.Nil(t, err)
	assert.NotNil(t, dto)
	if dto != nil {
		assert.Equal(t, model.ID.String(), dto.ID)
	}
}

func TestGetBuildWhenBuildIsInError(t *testing.T) {
	resetDB(testCfg.DatabaseDSN) //nolint:errcheck
	art := newArtifactory(testDB, nil)
	model, err := createBuildModel(testDB, artifactory.BuildError, request)
	assert.Nil(t, err)
	dto, err := art.GetBuild(request)
	assert.Nil(t, err)
	assert.NotNil(t, dto)
	if dto != nil {
		assert.Equal(t, model.ID.String(), dto.ID)
	}
}

func TestReservePendingBuildWhenNoneAvailable(t *testing.T) {
	resetDB(testCfg.DatabaseDSN) //nolint:errcheck
	art := newArtifactory(testDB, nil)
	model, err := art.ReservePendingBuild()
	assert.Nil(t, err)
	assert.Nil(t, model)
}

func TestBuildWhenFailingToDownload(t *testing.T) {
	resetDB(testCfg.DatabaseDSN) //nolint:errcheck
	art := newArtifactory(testDB, nil)
	model1, err := createBuildModel(testDB, artifactory.WaitingForBuild, request)
	assert.Nil(t, err)

	ctx := context.Background()
	recorder := buildlogs.NewRecorder()
	downloader := &MockDownloader{}
	downloader.
		On("Download", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("failed to download"))
	builder := &MockFirmwareBuilder{}

	model2, err := art.Build(ctx, model1, recorder, downloader, builder)
	assert.Error(t, err, "failed to download")
	downloader.AssertNumberOfCalls(t, "Download", 1)
	assert.Equal(t, model1.ID.String(), model2.ID.String())
	assert.Equal(t, artifactory.WaitingForBuild, model2.Status)
	assert.Equal(t, int64(1), model2.BuildAttempts)
}

func TestBuildWhenFailingToBuild(t *testing.T) {
	resetDB(testCfg.DatabaseDSN) //nolint:errcheck
	art := newArtifactory(testDB, nil)
	model1, err := createBuildModel(testDB, artifactory.WaitingForBuild, request)
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
	model2, err := art.Build(ctx, model1, recorder, downloader, builder)
	assert.Error(t, err, "failed to build")
	downloader.AssertNumberOfCalls(t, "Download", 1)
	builder.AssertNumberOfCalls(t, "Build", 1)
	assert.Equal(t, model1.ID.String(), model2.ID.String())
	assert.Equal(t, artifactory.WaitingForBuild, model2.Status)
	assert.Equal(t, int64(1), model2.BuildAttempts)
}

func TestBuildWhenFailingToUpload(t *testing.T) {
	uploader := &MockStorage{}
	uploader.
		On("Upload", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("failed to upload"))

	resetDB(testCfg.DatabaseDSN) //nolint:errcheck
	art := newArtifactory(testDB, uploader)
	model1, err := createBuildModel(testDB, artifactory.WaitingForBuild, request)
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
	model2, err := art.Build(ctx, model1, recorder, downloader, builder)
	assert.Error(t, err, "failed to upload")

	assert.Equal(t, model1.ID.String(), model2.ID.String())
	assert.Equal(t, artifactory.WaitingForBuild, model2.Status)
	assert.Equal(t, int64(1), model2.BuildAttempts)
}

func TestSuccessfulBuildJobFlow(t *testing.T) {
	resetDB(testCfg.DatabaseDSN) //nolint:errcheck
	art := newArtifactory(testDB, nil)
	model1, err := createBuildModel(testDB, artifactory.WaitingForBuild, request)
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
	model2, err := art.Build(ctx, model1, recorder, downloader, builder)
	assert.Nil(t, err)

	assert.Equal(t, model1.ID.String(), model2.ID.String())
	assert.Equal(t, artifactory.BuildSuccess, model2.Status)
	assert.Equal(t, int64(1), model2.BuildAttempts)
	assert.Equal(t,
		fmt.Sprintf("%s-%s.bin", model2.CommitHash, model2.BuildFlagsHash),
		model2.Artifacts[0].Filename,
	)
}

func TestJobGoesToErrorAfterTooManyFailures(t *testing.T) {
	resetDB(testCfg.DatabaseDSN) //nolint:errcheck
	art := newArtifactory(testDB, nil)
	model1, err := createBuildModel(testDB, artifactory.WaitingForBuild, request)
	assert.Nil(t, err)

	model1.BuildAttempts = 10

	repository := artifactory.NewBuildJobsDBRepository(testDB)
	err = repository.Save(model1)
	assert.Nil(t, err)

	ctx := context.Background()
	recorder := buildlogs.NewRecorder()
	downloader := &MockDownloader{}
	downloader.
		On("Download", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("download error"))
	builder := &MockFirmwareBuilder{}

	model2, err := art.Build(ctx, model1, recorder, downloader, builder)
	assert.Error(t, err, "download error")

	assert.Equal(t, model1.ID.String(), model2.ID.String())
	assert.Equal(t, artifactory.BuildError, model2.Status)
	assert.Equal(t, int64(11), model2.BuildAttempts)
}

func TestBackoffDurationForFailedBuild(t *testing.T) {
	resetDB(testCfg.DatabaseDSN) //nolint:errcheck
	art := newArtifactory(testDB, nil)
	model1, err := createBuildModel(testDB, artifactory.WaitingForBuild, request)
	assert.Nil(t, err)

	model1.BuildAttempts = 1
	model1.BuildEndedAt = time.Now()

	repository := artifactory.NewBuildJobsDBRepository(testDB)
	err = repository.Save(model1)
	assert.Nil(t, err)

	model2, err := art.ReservePendingBuild()
	assert.Nil(t, model2)
	assert.Nil(t, err)

	model3, err := repository.FindByID(model1.ID)
	assert.Nil(t, err)
	assert.Equal(t, model1.BuildAttempts, model3.BuildAttempts)
}

func TestMain(m *testing.M) {
	v := viper.New()
	v.Set("config-path", "./../test_config.yaml")
	config.InitConfig(v)()

	testCfg = config.NewOpts(v)
	if err := testCfg.Unmarshal(); err != nil {
		fmt.Println("failed to unmarshal config:", err)
		os.Exit(1)
	}

	log.Debugf("Config: %+v", v.AllSettings())
	log.Debugf("Opts: %+v", testCfg)

	if db, err := setupTestDB(testCfg); err != nil {
		fmt.Println("failed to setup test DB:", err)
		os.Exit(1)
	} else {
		testDB = db
	}

	os.Exit(m.Run())
}
