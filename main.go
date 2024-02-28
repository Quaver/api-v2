package main

import (
	"github.com/Quaver/api2/azure"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := config.Load("./config.json"); err != nil {
		logrus.Panic(err)
	}

	if !config.Instance.IsProduction {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.Infof("Log level set to: `%v`", logrus.GetLevel())

	db.ConnectMySQL()
	db.InitializeRedis()
	azure.InitializeClient()

	db.CacheTotalUsersInRedis()
	db.CacheTotalMapsetsInRedis()

	if config.Instance.IsProduction {
		// Extremely long start-up time. Only really need to do this once during debug to populate
		db.CacheTotalScoresInRedis()
	}

	initializeServer(config.Instance.Server.Port)
}
