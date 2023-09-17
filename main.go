package main

import (
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := config.Load("./config.json"); err != nil {
		logrus.Panic(err)
	}

	db.ConnectMySQL()
	initializeServer(config.Instance.Server.Port)
}
