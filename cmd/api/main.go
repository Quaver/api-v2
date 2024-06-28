package main

import (
	"github.com/Quaver/api2/azure"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/files"
	"github.com/Quaver/api2/webhooks"
	"github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

func main() {
	if err := config.Load("../../config.json"); err != nil {
		logrus.Panic(err)
	}

	if !config.Instance.IsProduction {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.Infof("Log level set to: `%v`", logrus.GetLevel())

	db.ConnectMySQL()
	db.InitializeRedis()
	db.InitializeElasticSearch()
	db.CacheTotalUsersInRedis()
	db.CacheTotalMapsetsInRedis()

	if config.Instance.IsProduction {
		db.CacheTotalScoresInRedis()
	}

	azure.InitializeClient()
	webhooks.InitializeWebhooks()
	files.CreateDirectories()

	rand.Seed(time.Now().UnixNano())
	initializeServer(config.Instance.Server.Port)
}
