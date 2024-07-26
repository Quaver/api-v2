package commands

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/Quaver/api2/azure"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/files"
	"github.com/Quaver/api2/webhooks"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	databaseBackupContainer string = "databasebackup"
)

var DatabaseBackupCmd = &cobra.Command{
	Use:   "backup:database",
	Short: "Backs up the database and uploads to azure",
	Run: func(cmd *cobra.Command, args []string) {
		sqlPath, _ := filepath.Abs(fmt.Sprintf("%v/backup.sql", files.GetBackupsDirectory()))
		zipPath, _ := filepath.Abs(fmt.Sprintf("%v/backup.sql.zip", files.GetBackupsDirectory()))

		if err := deletePreviousBackups(databaseBackupContainer, 2); err != nil {
			logrus.Error(err)
			_ = webhooks.SendBackupWebhook(false, err)
			return
		}

		azureFileName := "1.zip"

		blobs, err := azure.Client.ListBlobs(databaseBackupContainer)

		if err != nil {
			logrus.Error(err)
			_ = webhooks.SendBackupWebhook(false, err)
			return
		}

		// Get incremented value for azureFileName
		if len(blobs) > 0 {
			fileNumber, err := strconv.Atoi(strings.Replace(blobs[len(blobs)-1], ".zip", "", -1))

			if err != nil {
				logrus.Error(err)
				_ = webhooks.SendBackupWebhook(false, err)
				return
			}

			azureFileName = fmt.Sprintf("%v.zip", fileNumber+1)
		}

		if err := performDatabaseBackupBackup(sqlPath, zipPath, databaseBackupContainer, azureFileName); err != nil {
			_ = webhooks.SendBackupWebhook(false, err)
			return
		}

		_ = webhooks.SendBackupWebhook(true)
	},
}

func performDatabaseBackupBackup(sqlFilePath string, zipFilePath string, azureContainer string, azureFileName string) error {
	if err := os.Remove(sqlFilePath); err != nil {
		logrus.Error("[Database Backup] Error deleting existing backup (sql file).")
	}

	if err := os.Remove(zipFilePath); err != nil {
		logrus.Error("[Database Backup] Error deleting existing backup (zip file).")
	}

	if err := dumpDatabase(sqlFilePath); err != nil {
		logrus.Error("[Database Backup] Error dumping database: ", err)
		return err
	}

	if err := zipBackup(sqlFilePath, zipFilePath); err != nil {
		logrus.Error("[Database Backup] Error zipping database: ", err)
		return err
	}

	if err := uploadToAzure(zipFilePath, azureContainer, azureFileName); err != nil {
		logrus.Error("[Database Backup] Error uploading database backup", err)
		return err
	}

	return nil
}

func deletePreviousBackups(container string, maxBackups int) error {
	blobs, err := azure.Client.ListBlobs(container)

	if err != nil {
		return err
	}

	if len(blobs) < maxBackups {
		return nil
	}

	if err := azure.Client.DeleteBlob(container, blobs[0]); err != nil {
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

func zipBackup(inputPath string, outputPath string) error {
	logrus.Info("[Database Backup] Zipping backup file...")

	inputFile, err := os.Open(inputPath)

	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	zipFile, err := os.Create(outputPath)

	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}

	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	zipEntry, err := zipWriter.Create("backup.sql")

	if err != nil {
		return fmt.Errorf("failed to create zip entry: %w", err)
	}

	if _, err = io.Copy(zipEntry, inputFile); err != nil {
		return fmt.Errorf("failed to copy file content into zip entry: %w", err)
	}

	logrus.Info("[Database Backup] Successfully zipped backup file")
	return nil
}

func uploadToAzure(path string, container string, fileName string) error {
	logrus.Info("[Database Backup] Uploading to azure...:")

	err := azure.Client.UploadFileFromDisk(container, fileName, path, azblob.AccessTierHot)

	if err != nil {
		return err
	}

	logrus.Info("[Database Backup] Finished uploading to azure!")
	return nil
}
