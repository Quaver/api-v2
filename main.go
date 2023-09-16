package main

import (
	"github.com/Quaver/api2/config"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := config.Load("./config.json"); err != nil {
		logrus.Panic(err)
	}

	initializeServer(8080)
}
