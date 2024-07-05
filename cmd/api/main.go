package main

import (
	"flag"
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
	configPath := flag.String("config", "../../config.json", "path to config file")
	flag.Parse()

	if err := config.Load(*configPath); err != nil {
		logrus.Panic(err)
	}

	if !config.Instance.IsProduction {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.Infof("Log level set to: `%v`", logrus.GetLevel())

	db.ConnectMySQL()
	db.InitializeRedis()
	db.InitializeElasticSearch()
	go db.CacheTotalUsersInRedis()
	go db.CacheTotalMapsetsInRedis()
	go db.CacheTotalScoresInRedis()

	azure.InitializeClient()
	webhooks.InitializeWebhooks()
	files.CreateDirectories()

	rand.Seed(time.Now().UnixNano())
	initializeServer(config.Instance.Server.Port)
}
