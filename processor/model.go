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

func Heartbeat(c *config.CloudbuildOpts) {
	db, err := database.New(c.DatabaseDSN)
	if err != nil {
		panic(err)
	}
	var workerModel WorkerModel
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

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
	db, err := database.New(c.DatabaseDSN)
	if err != nil {
		panic(err)
	}

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
