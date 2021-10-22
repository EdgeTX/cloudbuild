package main

import (
	"flag"
	"os"

	"github.com/edgetx/cloudbuild/database"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func init() {
	log.SetOutput(os.Stdout)

	logLevel, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		log.Errorf("failed to parse LOG_LEVEL: %s", err)
		logLevel = log.DebugLevel
	}
	log.SetLevel(logLevel)
}

func main() {
	db, err := gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connecto to the database: %s", err)
	}

	var migrate bool
	flag.BoolVar(&migrate, "migrate", false, "specify -migrate to perform database migration")
	var dropSchema bool
	flag.BoolVar(&dropSchema, "drop-schema", false, "specify -dropSchema-schema to clear the database")
	flag.Parse()

	if !dropSchema && !migrate {
		log.Fatal("db cli options empty")
	}

	if migrate {
		err := database.Migrate(db)
		if err != nil {
			log.Fatalf("failed to migrate database: %s", err)
		} else {
			log.Infof("database was migrated successfully")
		}
	}

	if dropSchema {
		err := database.DropSchema(db)
		if err != nil {
			log.Fatalf("failed to dropSchema schema: %s", err)
		} else {
			log.Infof("schema successfully")
		}
	}
}
