package commands

import (
	"fmt"
	"github.com/Quaver/api2/files"
	"github.com/Quaver/api2/webhooks"
	"github.com/spf13/cobra"
	"path/filepath"
)

const (
	s3HourlyBackupFolderName string = "backups-hourly"
)

var DatabaseBackupHourlyCmd = &cobra.Command{
	Use:   "backup:database:hourly",
	Short: "Backs up the database hourly and uploads to s3",
	Run: func(cmd *cobra.Command, args []string) {
		sqlPath, _ := filepath.Abs(fmt.Sprintf("%v/backup-hourly.sql", files.GetBackupsDirectory()))
		s3FileName := "backup-hourly.sql.zip"
		zipPath, _ := filepath.Abs(fmt.Sprintf("%v/%v", files.GetBackupsDirectory(), s3FileName))

		if err := performDatabaseBackupBackup(sqlPath, zipPath, s3HourlyBackupFolderName, s3FileName); err != nil {
			_ = webhooks.SendBackupWebhook(false, err)
			return
		}

		_ = webhooks.SendBackupWebhook(true)
	},
}
