package config

import (
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"os"
)

type Config struct {
	Server struct {
		Port int `json:"port"`
	} `json:"server"`

	SQL struct {
		Host     string `json:"host"`
		Username string `json:"username"`
		Password string `json:"password"`
		Database string `json:"database"`
	} `json:"sql"`
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
