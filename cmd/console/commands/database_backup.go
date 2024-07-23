package commands

import (
	"bytes"
	"fmt"
	"github.com/Quaver/api2/azure"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/webhooks"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var DatabaseBackupCmd = &cobra.Command{
	Use:   "backup:database",
	Short: "Backs up the database and uploads to azure",
	Run: func(cmd *cobra.Command, args []string) {
		backupDir := fmt.Sprintf("%v/backups", config.Instance.Cache.DataDirectory)

		if err := os.MkdirAll(backupDir, os.ModePerm); err != nil {
			logrus.Error("[Database Backup] Error creating backup directory", err)
			_ = webhooks.SendBackupWebhook(false, err)
			return
		}

		var err error
		path, _ := filepath.Abs(fmt.Sprintf("%v/backup.sql", backupDir))

		if err := os.Remove(path); err != nil {
			logrus.Error("[Database Backup] Error deleting existing backup")
			_ = webhooks.SendBackupWebhook(false, err)
			return
		}

		logrus.Info("[Database Backup] Dumping database...")

		if err := dumpDatabase(path); err != nil {
			logrus.Error("[Database Backup] Error dumping database: ", err)
			_ = webhooks.SendBackupWebhook(false, err)
			return
		}

		logrus.Info("[Database Backup] Finished dumping database at path: ", path)

		err = azure.Client.UploadFileFromDisk("databasebackup", "latest-backup.sql", path, nil)

		if err != nil {
			logrus.Error("[Database Backup] Error uploading database backup", err)
			_ = webhooks.SendBackupWebhook(false, err)
			return
		}

		logrus.Info("[Database Backup] Database backup complete!")
		_ = webhooks.SendBackupWebhook(true)
	},
}

func dumpDatabase(path string) error {
	hostSplit := strings.Split(config.Instance.SQL.Host, ":")

	cmd := exec.Command(
		"mysqldump",
		"--single-transaction",
		"-P",
		hostSplit[1],
		"-h",
		hostSplit[0],
		"-u",
		config.Instance.SQL.Username,
		fmt.Sprintf("-p%v", config.Instance.SQL.Password),
		config.Instance.SQL.Database,
		fmt.Sprintf("--result-file=%v", path),
	)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		return fmt.Errorf("%v\n\n```%v```", err, stderr.String())
	}

	return nil
}
