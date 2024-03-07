package files

import (
	"fmt"
	"github.com/Quaver/api2/config"
	"github.com/sirupsen/logrus"
	"os"
)

// CreateDirectories Creates the directories needed for the cache
func CreateDirectories() {
	err := os.MkdirAll(config.Instance.Cache.DataDirectory, os.ModePerm)

	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(GetMapsDirectory(), os.ModePerm)

	if err != nil {
		panic(err)
	}

	logrus.Info("Created cache directories")
}

func GetMapsDirectory() string {
	return fmt.Sprintf("%v/maps", config.Instance.Cache.DataDirectory)
}
