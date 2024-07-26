package commands

import (
	"fmt"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/webhooks"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

const (
	databaseBackupHourlyContainer string = "databasebackuphourly"
)

var DatabaseBackupHourlyCmd = &cobra.Command{
	Use:   "backup:database:hourly",
	Short: "Backs up the database hourly and uploads to azure",
	Run: func(cmd *cobra.Command, args []string) {
		backupDir := fmt.Sprintf("%v/backups", config.Instance.Cache.DataDirectory)
		sqlPath, _ := filepath.Abs(fmt.Sprintf("%v/backup-hourly.sql", backupDir))
		azureFileName := "backup-hourly.sql.zip"
		zipPath, _ := filepath.Abs(fmt.Sprintf("%v/%v", backupDir, azureFileName))

		if err := os.MkdirAll(backupDir, os.ModePerm); err != nil {
			logrus.Error("[Hourly Database Backup] Error creating backup directory", err)
			_ = webhooks.SendBackupWebhook(false, err)
			return
		}

		if err := os.Remove(sqlPath); err != nil {
			logrus.Error("[Hourly Database Backup] Error deleting existing backup")
		}

		if err := os.Remove(zipPath); err != nil {
			logrus.Error("[Hourly Database Backup] Error deleting existing backup")
		}

		if err := dumpDatabase(sqlPath); err != nil {
			logrus.Error("[Hourly Database Backup] Error dumping database: ", err)
			_ = webhooks.SendBackupWebhook(false, err)
			return
		}

		if err := zipBackup(sqlPath, zipPath); err != nil {
			logrus.Error("[Hourly Database Backup] Error zipping database: ", err)
			_ = webhooks.SendBackupWebhook(false, err)
			return
		}

		if err := uploadToAzure(zipPath, databaseBackupHourlyContainer, azureFileName); err != nil {
			logrus.Error("[Hourly Database Backup] Error uploading database backup", err)
			_ = webhooks.SendBackupWebhook(false, err)
			return
		}

		_ = webhooks.SendBackupWebhook(true)
	},
}
