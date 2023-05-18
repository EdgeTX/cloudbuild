package processor

import (
	"database/sql"
	"errors"
	"os"
	"time"

	"github.com/edgetx/cloudbuild/config"
	"github.com/edgetx/cloudbuild/database"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

const (
	HeartbeatInterval = time.Second * 10
	Timeout           = time.Second * 30
)

type WorkerModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	Hostname  string    `gorm:"index:worker_hostname_idx"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (WorkerModel) TableName() string {
	return "workers"
}

func (base *WorkerModel) BeforeCreate(db *gorm.DB) error {
	base.ID = uuid.NewV4()
	return nil
}

type WorkerDB struct {
	db *gorm.DB
}

func NewWorkerDB(c *config.CloudbuildOpts) *WorkerDB {
	return &WorkerDB{db: newDB(c)}
}

func (w *WorkerDB) List() (*[]WorkerModel, error) {
	var workers []WorkerModel
	err := w.db.Find(&workers).Error
	return &workers, err
}

func newDB(c *config.CloudbuildOpts) *gorm.DB {
	db, err := database.New(c.DatabaseDSN)
	if err != nil {
		panic(err)
	}
	return db
}

func Heartbeat(c *config.CloudbuildOpts) {
	var workerModel WorkerModel
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	db := newDB(c)
	err = db.Where(&WorkerModel{
		Hostname: hostname,
	}).First(&workerModel).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		workerModel.Hostname = hostname
		if err := db.Create(&workerModel).Error; err != nil {
			panic(err)
		}
	}

	for {
		db.Save(&workerModel)
		time.Sleep(HeartbeatInterval)
	}
}

func GarbageCollector(c *config.CloudbuildOpts) {
	db := newDB(c)
	for {
		db.Where(
			"updated_at < @updatedAt",
			sql.Named("updatedAt", time.Now().Add(-Timeout)),
		).Delete(&WorkerModel{})
		time.Sleep(HeartbeatInterval)
	}
}

func init() {
	database.RegisterModels(
		&WorkerModel{},
	)
}
