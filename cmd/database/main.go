package main

import (
	"flag"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sirupsen/logrus"
	"log"
	"strconv"
)

func main() {
	direction := flag.String("direction", "up", "Direction to run migration: up or down")
	version := flag.String("version", "", "Target migration version")
	steps := flag.Int("steps", 0, "Number of migration steps to apply")
	flag.Parse()

	if err := config.Load("../../config.json"); err != nil {
		logrus.Panic(err)
	}

	db.ConnectMySQL()

	sqlDB, err := db.SQL.DB()
	if err != nil {
		log.Fatal(err)
	}

	driver, err := mysql.WithInstance(sqlDB, &mysql.Config{})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "mysql", driver)
	if err != nil {
		log.Fatal(err)
	}

	if *version != "" {
		// Migrate to a specific version
		v, err := strconv.ParseUint(*version, 10, 64)
		if err != nil {
			log.Fatalf("Invalid version format: %v", err)
		}
		if err := m.Migrate(uint(v)); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		logrus.Infof("Migrated to version %d successfully!", v)
	} else if *steps != 0 {
		// Migrate by specific steps
		if err := m.Steps(*steps); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		logrus.Infof("Migrated %d steps successfully!", *steps)
	} else {
		// Migrate up or down to the latest or initial version
		switch *direction {
		case "up":
			if err := m.Up(); err != nil && err != migrate.ErrNoChange {
				log.Fatal(err)
			}
			logrus.Info("Migrations applied successfully!")
		case "down":
			if err := m.Down(); err != nil && err != migrate.ErrNoChange {
				log.Fatal(err)
			}
			logrus.Info("Migrations rolled back successfully!")
		default:
			log.Fatalf("Unknown migration direction: %s", *direction)
		}
	}
}
