package artifactory

import (
	"database/sql"
	"strings"
	"time"

	"github.com/edgetx/cloudbuild/config"
	"github.com/edgetx/cloudbuild/database"
	"github.com/edgetx/cloudbuild/targets"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BuildJobsRepository interface {
	Get(request *BuildRequest) (*BuildJobModel, error)
	List(query *JobQuery) (*database.Pagination, error)
	Delete(id uuid.UUID) error
	FindByID(ID uuid.UUID) (*BuildJobModel, error)
	Create(model BuildJobModel) (*BuildJobModel, error)
	Save(model *BuildJobModel) error
	ReservePendingBuild() (*BuildJobModel, error)
	TimeoutBuilds(timeout time.Duration) error
	UpdateMetrics(queued, building, failed prometheus.Gauge)
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

func (repository *BuildJobsDBRepository) Get(request *BuildRequest) (*BuildJobModel, error) {
	var buildJob BuildJobModel
	err := repository.db.Where(&BuildJobModel{
		CommitHash:     targets.GetCommitHashByRef(request.Release),
		Target:         request.Target,
		BuildFlagsHash: request.HashTargetAndFlags(),
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

var statusLookup = map[string]interface{}{
	"all":         "",
	"success":     string(BuildSuccess),
	"error":       string(BuildError),
	"queued":      string(WaitingForBuild),
	"building":    string(BuildInProgress),
	"in-progress": []string{string(WaitingForBuild), string(BuildInProgress)},
}

func jobQueryClause(query *JobQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if query.Status != "" {
			status, ok := statusLookup[query.Status]
			if ok && status != "" {
				db = db.Where("status IN (?)", status)
			}
		}
		if query.Target != "" {
			targets := strings.Split(query.Target, ",")
			db = db.Where("target IN(?)", targets)
		}
		if query.Release != "" {
			refs := strings.Split(query.Release, ",")
			db = db.Where("commit_ref IN(?)", refs)
		}
		return db
	}
}

func (repository *BuildJobsDBRepository) List(query *JobQuery) (*database.Pagination, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}

	var jobs []BuildJobModel
	tx := repository.db.Preload("Artifacts").Scopes(jobQueryClause(query))
	err := tx.Scopes(
		database.Paginate(
			&BuildJobModel{}, query, tx,
		),
	).Find(&jobs).Error

	res := query.Pagination
	res.Rows = &jobs
	return &res, err
}

func (repository *BuildJobsDBRepository) Delete(id uuid.UUID) error {
	return repository.db.Select(clause.Associations).Delete(&BuildJobModel{ID: id}).Error
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

func countRequestsByStatus(status interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Model(&BuildJobModel{}).Where(
			"status = ?", status,
		)
	}
}

func updateMetricsByStatus(db *gorm.DB, gauge prometheus.Gauge, status interface{}) {
	var count int64
	err := db.Scopes(countRequestsByStatus(status)).Count(&count).Error
	if err != nil {
		log.Errorf("failed to query: %s", err.Error())
	}
	gauge.Set(float64(count))
}

func (repository *BuildJobsDBRepository) UpdateMetrics(
	queued, building, failed prometheus.Gauge,
) {
	log.Debugln("UpdateMetrics")
	updateMetricsByStatus(repository.db, queued, WaitingForBuild)
	updateMetricsByStatus(repository.db, building, BuildInProgress)
	updateMetricsByStatus(repository.db, failed, BuildError)
}
