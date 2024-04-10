package config

import (
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"os"
)

type Config struct {
	IsProduction bool `json:"is_production"`

	JWTSecret string `json:"jwt_secret"`

	Server struct {
		Port int `json:"port"`
	} `json:"server"`

	Steam struct {
		AppId        int    `json:"app_id"`
		PublisherKey string `json:"publisher_key"`
	} `json:"steam"`

	SQL struct {
		Host     string `json:"host"`
		Username string `json:"username"`
		Password string `json:"password"`
		Database string `json:"database"`
	} `json:"sql"`

	Redis struct {
		Host     string `json:"host"`
		Password string `json:"password"`
		Database int    `json:"database"`
	} `json:"redis"`

	Azure struct {
		AccountName string `json:"account_name"`
		AccountKey  string `json:"account_key"`
	} `json:"azure"`

	Cache struct {
		DataDirectory string `json:"data_directory"`
	}

	QuaverToolsPath string `json:"quaver_tools_path"`
}

var Instance *Config = nil

func Load(path string) error {
	if Instance != nil {
		return errors.New("config already loaded")
	}

	data, err := os.ReadFile(path)

	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &Instance)

	if err != nil {
		return err
	}

	logrus.Info("Config file has been loaded")
	return nil
}
