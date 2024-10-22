package main

import (
	"flag"
	"github.com/Quaver/api2/azure"
	"github.com/Quaver/api2/cmd/console/commands"
	"github.com/Quaver/api2/cmd/console/migrations"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/s3util"
	"github.com/Quaver/api2/webhooks"
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
	configPath := flag.String("config", "../../config.json", "path to config file")
	flag.Parse()

	if err := config.Load(*configPath); err != nil {
		logrus.Panic(err)
	}

	db.ConnectMySQL()
	db.InitializeRedis()
	db.InitializeElasticSearch()
	azure.InitializeClient()
	s3util.Initialize()
	webhooks.InitializeWebhooks()

	// Commands
	RootCmd.AddCommand(commands.CacheClearCmd)
	RootCmd.AddCommand(commands.CacheLeaderboardCmd)
	RootCmd.AddCommand(commands.CacheClanLeaderboard)
	RootCmd.AddCommand(commands.ElasticIndexMapsets)
	RootCmd.AddCommand(commands.PlayerDonatorCheckCmd)
	RootCmd.AddCommand(commands.WeeklyMostPlayedMapsetsCmd)
	RootCmd.AddCommand(commands.UserRankCmd)
	RootCmd.AddCommand(commands.CacheScoreboardClearCmd)
	RootCmd.AddCommand(commands.DatabaseBackupCmd)
	RootCmd.AddCommand(commands.DatabaseBackupHourlyCmd)
	RootCmd.AddCommand(commands.SupervisorActivityCmd)
	RootCmd.AddCommand(commands.ClanRankMapCmd)
	RootCmd.AddCommand(commands.DenyOnHoldCmd)
	RootCmd.AddCommand(commands.DatabaseScoresBatchDeleteFailed)
	RootCmd.AddCommand(commands.DeleteScoreCmd)
	RootCmd.AddCommand(commands.FixStatsCmd)
	RootCmd.AddCommand(commands.UpdateStripePriceId)

	// Migrations
	RootCmd.AddCommand(migrations.MigrationPlaylistMapsetCmd)
}
