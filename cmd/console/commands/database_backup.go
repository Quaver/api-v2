package commands

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/files"
	"github.com/Quaver/api2/s3util"
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
	s3BackupFolderName string = "backups"
)

var DatabaseBackupCmd = &cobra.Command{
	Use:   "backup:database",
	Short: "Backs up the database and uploads to s3",
	Run: func(cmd *cobra.Command, args []string) {
		sqlPath, _ := filepath.Abs(fmt.Sprintf("%v/backup.sql", files.GetBackupsDirectory()))
		zipPath, _ := filepath.Abs(fmt.Sprintf("%v/backup.sql.zip", files.GetBackupsDirectory()))

		if err := deletePreviousBackups(s3BackupFolderName, 28); err != nil {
			logrus.Error(err)
			_ = webhooks.SendBackupWebhook(false, err)
			return
		}

		files, err := s3util.Instance().ListFiles(s3BackupFolderName)

		if err != nil {
			logrus.Error(err)
			_ = webhooks.SendBackupWebhook(false, err)
			return
		}

		s3FileName := "001.zip"

		// Get incremented value for s3 file name
		if len(files) > 0 {
			name := strings.Replace(files[len(files)-1], ".zip", "", -1)
			name = strings.Replace(name, fmt.Sprintf("%v/", s3BackupFolderName), "", -1)

			fileNumber, err := strconv.Atoi(name)

			if err != nil {
				logrus.Error(err)
				_ = webhooks.SendBackupWebhook(false, err)
				return
			}

			s3FileName = fmt.Sprintf("%03d.zip", fileNumber+1)
		}

		if err := performDatabaseBackupBackup(sqlPath, zipPath, s3BackupFolderName, s3FileName); err != nil {
			_ = webhooks.SendBackupWebhook(false, err)
			return
		}

		_ = webhooks.SendBackupWebhook(true)
	},
}

func performDatabaseBackupBackup(sqlFilePath string, zipFilePath string, s3Folder string, s3FileName string) error {
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

	if err := uploadToS3(zipFilePath, s3Folder, s3FileName); err != nil {
		logrus.Error("[Database Backup] Error uploading database backup", err)
		return err
	}

	if err := os.Remove(sqlFilePath); err != nil {
		logrus.Error("[Database Backup] Error deleting existing backup (sql file).")
	}

	return nil
}

func deletePreviousBackups(folderName string, maxBackups int) error {
	backupFiles, err := s3util.Instance().ListFiles(folderName)

	if err != nil {
		return err
	}

	if len(backupFiles) < maxBackups {
		return nil
	}

	if err := s3util.Instance().DeleteFile(folderName, filepath.Base(backupFiles[0])); err != nil {
		return err
	}

	logrus.Info("[Database Backup] Deleted previous backup: ", backupFiles[0])
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

func uploadToS3(path string, folder string, fileName string) error {
	logrus.Info("[Database Backup] Uploading to S3...")

	err := s3util.Instance().UploadFile(folder, fileName, path)

	if err != nil {
		return err
	}

	logrus.Info("[Database Backup] Finished uploading to S3!")
	return nil
}
