package main

import (
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/files"
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
	files.InitializeAzure()
	initializeServer(config.Instance.Server.Port)
}
