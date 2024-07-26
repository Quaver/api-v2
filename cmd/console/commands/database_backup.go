package commands

import (
	"bytes"
	"fmt"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/Quaver/api2/azure"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/webhooks"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	databaseBackupContainer string = "databasebackup"
)

var DatabaseBackupCmd = &cobra.Command{
	Use:   "backup:database",
	Short: "Backs up the database and uploads to azure",
	Run: func(cmd *cobra.Command, args []string) {
		if err := deletePreviousBackups(); err != nil {
			logrus.Error(err)
			_ = webhooks.SendBackupWebhook(false, err)
			return
		}

		backupDir := fmt.Sprintf("%v/backups", config.Instance.Cache.DataDirectory)
		path, _ := filepath.Abs(fmt.Sprintf("%v/backup.sql", backupDir))

		if err := os.MkdirAll(backupDir, os.ModePerm); err != nil {
			logrus.Error("[Database Backup] Error creating backup directory", err)
			_ = webhooks.SendBackupWebhook(false, err)
			return
		}

		if err := os.Remove(path); err != nil {
			logrus.Error("[Database Backup] Error deleting existing backup")
		}

		if err := dumpDatabase(path); err != nil {
			logrus.Error("[Database Backup] Error dumping database: ", err)
			_ = webhooks.SendBackupWebhook(false, err)
			return
		}

		if err := uploadToAzure(path); err != nil {
			logrus.Error("[Database Backup] Error uploading database backup", err)
			_ = webhooks.SendBackupWebhook(false, err)
			return
		}

		_ = webhooks.SendBackupWebhook(true)
	},
}

func deletePreviousBackups() error {
	blobs, err := azure.Client.ListBlobs(databaseBackupContainer)

	if err != nil {
		return err
	}

	if len(blobs) < 12 {
		return nil
	}

	if err := azure.Client.DeleteBlob(databaseBackupContainer, blobs[0]); err != nil {
		return err
	}

	logrus.Info("[Database Backup] Deleted previous backup: ", blobs[0])
	return nil
}

func dumpDatabase(path string) error {
	logrus.Info("[Database Backup] Dumping database...")

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

	logrus.Info("[Database Backup] Finished dumping database at path: ", path)
	return nil
}

func uploadToAzure(path string) error {
	logrus.Info("[Database Backup] Uploading to azure...:")

	currentTime := time.Now()

	fileName := fmt.Sprintf("%d-%d-%d-time-%d-%d.sql", currentTime.Year(), currentTime.Month(), currentTime.Day(), currentTime.Hour(), currentTime.Minute())

	err := azure.Client.UploadFileFromDisk(databaseBackupContainer, fileName, path, azblob.AccessTierHot)

	if err != nil {
		return err
	}

	logrus.Info("[Database Backup] Finished uploading to azure!")
	return nil
}
