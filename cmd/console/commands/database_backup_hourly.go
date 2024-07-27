package commands

import (
	"fmt"
	"github.com/Quaver/api2/files"
	"github.com/Quaver/api2/webhooks"
	"github.com/spf13/cobra"
	"path/filepath"
)

const (
	databaseBackupHourlyContainer string = "databasebackuphourly"
)

var DatabaseBackupHourlyCmd = &cobra.Command{
	Use:   "backup:database:hourly",
	Short: "Backs up the database hourly and uploads to azure",
	Run: func(cmd *cobra.Command, args []string) {
		sqlPath, _ := filepath.Abs(fmt.Sprintf("%v/backup-hourly.sql", files.GetBackupsDirectory()))
		azureFileName := "backup-hourly.sql.zip"
		zipPath, _ := filepath.Abs(fmt.Sprintf("%v/%v", files.GetBackupsDirectory(), azureFileName))

		if err := performDatabaseBackupBackup(sqlPath, zipPath, databaseBackupHourlyContainer, azureFileName); err != nil {
			_ = webhooks.SendBackupWebhook(false, err)
			return
		}

		_ = webhooks.SendBackupWebhook(true)
	},
}
