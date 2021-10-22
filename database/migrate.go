package database

import (
	"github.com/edgetx/cloudbuild/artifactory"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&artifactory.BuildJobModel{},
		&artifactory.ArtifactModel{},
		&artifactory.AuditLogModel{},
	)
}

func DropSchema(db *gorm.DB) error {
	err := db.Exec("DROP SCHEMA IF EXISTS public CASCADE;").Error
	if err != nil {
		return errors.Wrap(err, "failed to drop schema")
	}

	err = db.Exec("CREATE SCHEMA IF NOT EXISTS public;").Error
	if err != nil {
		return errors.Wrap(err, "failed to create db")
	}
	return nil
}
