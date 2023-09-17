package db

import (
	"github.com/Quaver/api2/config"
	"gorm.io/gorm"
	"testing"
)

func TestGetUserById(t *testing.T) {
	_ = config.Load(testConfigPath)
	ConnectMySQL()

	_, err := GetUserById(1)

	if err != nil {
		t.Fatal(err)
	}

	CloseMySQL()
}

func TestGetUserByIdNotFound(t *testing.T) {
	_ = config.Load(testConfigPath)
	ConnectMySQL()

	_, err := GetUserById(-100)

	if err != nil && err != gorm.ErrRecordNotFound {
		t.Fatal(err)
	}

	CloseMySQL()
}
