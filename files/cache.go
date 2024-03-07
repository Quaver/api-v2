package files

import (
	"errors"
	"fmt"
	"github.com/Quaver/api2/azure"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/sirupsen/logrus"
	"os"
)

// CreateDirectories Creates the directories needed for the cache
func CreateDirectories() {
	err := os.MkdirAll(config.Instance.Cache.DataDirectory, os.ModePerm)

	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(getMapsDirectory(), os.ModePerm)

	if err != nil {
		panic(err)
	}

	logrus.Info("Created cache directories")
}

// CacheQuaFile Caches a .qua file. Returns the path of the file
func CacheQuaFile(mapQua *db.MapQua) (string, error) {
	fileName := fmt.Sprintf("%v.qua", mapQua.Id)
	path := fmt.Sprintf("%v/%v", getMapsDirectory(), fileName)

	// Check for existing file & see if md5 hash matches
	if _, err := os.Stat(path); err == nil {
		md5, err := GetFileMD5(path)

		if err != nil {
			return "", err
		}

		if md5 == mapQua.MD5 {
			return path, nil
		}
	}

	buffer, err := azure.Client.DownloadFile("maps", fileName, path)

	if err != nil {
		return "", err
	}

	// Handle donator-submitted maps (unsubmitted). These are gzipped, so un-gzip and rewrite file.
	if mapQua.RankedStatus == enums.RankedStatusNotSubmitted {
		if err = ungzipFile(&buffer, path); err != nil {
			return "", err
		}
		return path, nil
	}

	// Final MD5 Hash check for submitted maps
	md5, err := GetFileMD5(path)

	if err != nil {
		return "", err
	}

	if md5 != mapQua.MD5 {
		return "", errors.New("md5 hash does match during final check")
	}

	return path, nil
}

func getMapsDirectory() string {
	return fmt.Sprintf("%v/maps", config.Instance.Cache.DataDirectory)
}
