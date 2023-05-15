package artifactory

import (
	"database/sql"
	"time"

	"github.com/edgetx/cloudbuild/config"
	"github.com/edgetx/cloudbuild/database"
	"github.com/edgetx/cloudbuild/firmware"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type JobQuery struct {
	database.Pagination
	Status string `form:"status"`
}

func (q *JobQuery) Validate() error {
	switch q.Sort {
	case "", "created_at", "updated_at", "build_started_at", "build_ended_at":
		return nil
	default:
		return database.ErrBadSortAttribute
	}
}

type BuildJobsRepository interface {
	Get(commitHash string, flags []firmware.BuildFlag) (*BuildJobModel, error)
	List(query *JobQuery) (*database.Pagination, error)
	FindByID(ID uuid.UUID) (*BuildJobModel, error)
	Create(model BuildJobModel) (*BuildJobModel, error)
	ReservePendingBuild() (*BuildJobModel, error)
	TimeoutBuilds(timeout time.Duration) error
	Save(model *BuildJobModel) error
}

type BuildJobsDBRepository struct {
	db *gorm.DB
}

func NewBuildJobsDBRepository(db *gorm.DB) *BuildJobsDBRepository {
	return &BuildJobsDBRepository{
		db: db,
	}
}

func NewBuildJobsDBRepositoryFromConfig(c *config.CloudbuildOpts) (*BuildJobsDBRepository, error) {
	db, err := database.New(c.DatabaseDSN)
	if err != nil {
		return nil, err
	}
	return NewBuildJobsDBRepository(db), nil
}

func (repository *BuildJobsDBRepository) Get(commitHash string, flags []firmware.BuildFlag) (*BuildJobModel, error) {
	var buildJob BuildJobModel
	err := repository.db.Where(&BuildJobModel{
		CommitHash:     commitHash,
		BuildFlagsHash: HashBuildFlags(flags),
	}).Preload("Artifacts").First(&buildJob).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &buildJob, nil
}

func (repository *BuildJobsDBRepository) FindByID(id uuid.UUID) (*BuildJobModel, error) {
	var buildJob BuildJobModel
	err := repository.db.Where(&BuildJobModel{
		ID: id,
	}).Preload("Artifacts").First(&buildJob).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &buildJob, nil
}

func (repository *BuildJobsDBRepository) List(query *JobQuery) (*database.Pagination, error) {

	if err := query.Validate(); err != nil {
		return nil, err
	}
	
	tx := repository.db.Preload("Artifacts")
	if query.Status != "" {
		tx = tx.Where("status = ?", query.Status)
	}

	var jobs []BuildJobModel
	log.Debugln("Sort:", query.Pagination.Sort, "SoftDesc:", query.Pagination.SortDesc)
	err := tx.Debug().Scopes(database.Paginate(
		&BuildJobModel{}, &query.Pagination, tx),
	).Find(&jobs).Error

	res := query.Pagination
	res.Rows = &jobs
	return &res, err
}

func (repository *BuildJobsDBRepository) Create(model BuildJobModel) (*BuildJobModel, error) {
	err := repository.db.Session(
		&gorm.Session{FullSaveAssociations: true},
	).Create(&model).Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to create build job")
	}
	return &model, nil
}

func (repository *BuildJobsDBRepository) TimeoutBuilds(timeout time.Duration) error {
	return repository.db.Exec(
		`
			UPDATE build_jobs
			SET status = @newStatus
			WHERE status = @currentStatus AND build_started_at < @buildStartedAt
		`,
		sql.Named("currentStatus", BuildInProgress),
		sql.Named("newStatus", BuildError),
		sql.Named("buildStartedAt", time.Now().Add(-1*timeout)),
	).Error
}

func (repository *BuildJobsDBRepository) ReservePendingBuild() (*BuildJobModel, error) {
	var jobs []BuildJobModel
	backoffDuration := time.Minute
	err := repository.db.Raw(
		`
			UPDATE build_jobs
			SET status = @newStatus, build_started_at = @buildStartedAt
			WHERE id = (
				SELECT id FROM build_jobs
				WHERE status = @currentStatus AND build_ended_at < @minBuildEndedAt
				LIMIT 1
				FOR UPDATE SKIP LOCKED
			)
			RETURNING id
		`,
		sql.Named("newStatus", BuildInProgress),
		sql.Named("buildStartedAt", time.Now()),
		sql.Named("currentStatus", WaitingForBuild),
		sql.Named("minBuildEndedAt", time.Now().Add(backoffDuration*-1)),
	).Scan(&jobs).Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to reserve job for build")
	}

	if len(jobs) == 0 {
		return nil, nil
	}

	var buildJob BuildJobModel
	err = repository.db.Where(&BuildJobModel{
		ID: jobs[0].ID,
	}).Preload("Artifacts").First(&buildJob).Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to find reserved job")
	}

	return &buildJob, nil
}

func (repository *BuildJobsDBRepository) Save(model *BuildJobModel) error {
	return repository.db.Session(
		&gorm.Session{FullSaveAssociations: true},
	).Save(model).Error
}
