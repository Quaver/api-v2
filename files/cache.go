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
	if err := os.RemoveAll(GetTempDirectory()); err != nil {
		panic(err)
	}

	directories := []string{
		config.Instance.Cache.DataDirectory,
		getMapsDirectory(),
		getMapsetDirectory(),
		getReplayDirectory(),
		GetTempDirectory(),
		fmt.Sprintf("%v/multiplayer", GetTempDirectory()),
	}

	for _, directory := range directories {
		if err := os.MkdirAll(directory, os.ModePerm); err != nil {
			panic(err)
		}
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

// CacheMapset Caches a mapset file and returns the path to it
func CacheMapset(mapset *db.Mapset) (string, error) {
	fileName := fmt.Sprintf("%v.qp", mapset.Id)
	path := fmt.Sprintf("%v/%v", getMapsetDirectory(), fileName)

	// Check MD5 hash of existing file
	if _, err := os.Stat(path); err == nil {
		md5, err := GetFileMD5(path)

		if err != nil {
			return "", err
		}

		if md5 == mapset.PackageMD5 {
			return path, nil
		}
	}

	if _, err := azure.Client.DownloadFile("mapsets", fileName, path); err != nil {
		return "", err
	}

	return path, nil
}

// CacheReplay Caches a replay file and returns the path to it
func CacheReplay(scoreId int) (string, error) {
	fileName := fmt.Sprintf("%v.qr", scoreId)
	path := fmt.Sprintf("%v/%v", getReplayDirectory(), fileName)

	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	if _, err := azure.Client.DownloadFile("replays", fileName, path); err != nil {
		logrus.Error(err)
		return "", err
	}

	return path, nil
}

func getMapsDirectory() string {
	return fmt.Sprintf("%v/maps", config.Instance.Cache.DataDirectory)
}

func getMapsetDirectory() string {
	return fmt.Sprintf("%v/mapsets", config.Instance.Cache.DataDirectory)
}

func getReplayDirectory() string {
	return fmt.Sprintf("%v/replays", config.Instance.Cache.DataDirectory)
}

func GetTempDirectory() string {
	return "./temp"
}
