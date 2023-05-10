package database

import (
	"fmt"
)

var models []interface{}

func RegisterModels(objs ...interface{}) {
	models = append(models, objs...)
}

func Migrate(dsn string) error {
	if db, err := New(dsn); err != nil {
		return err
	} else {
		return db.AutoMigrate(models...)
	}
}

func DropSchema(dsn string) error {
	if db, err := New(dsn); err != nil {
		return err
	} else {
		err = db.Exec("DROP SCHEMA IF EXISTS public CASCADE;").Error
		if err != nil {
			return fmt.Errorf("failed to drop schema: %w", err)
		}

		err = db.Exec("CREATE SCHEMA IF NOT EXISTS public;").Error
		if err != nil {
			return fmt.Errorf("failed to create db: %w", err)
		}
		return nil
	}
}
