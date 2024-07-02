package main

import (
	"github.com/Quaver/api2/cmd/console/commands"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use: "cli",
}

func Execute() error {
	return RootCmd.Execute()
}

func init() {
	if err := config.Load("../../config.json"); err != nil {
		logrus.Panic(err)
	}

	db.ConnectMySQL()
	db.InitializeRedis()

	RootCmd.AddCommand(commands.CacheClearCmd)
	RootCmd.AddCommand(commands.CacheLeaderboardCmd)
	RootCmd.AddCommand(commands.PlayerDonatorCheckCmd)
}
