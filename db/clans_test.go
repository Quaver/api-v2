package db

import (
	"github.com/Quaver/api2/config"
	"testing"
)

func TestGetClanByName(t *testing.T) {
	_ = config.Load(testConfigPath)
	ConnectMySQL()

	_, err := GetClanByName("Act Broke Stay Broke")

	if err != nil {
		t.Fatal(err)
	}

	CloseMySQL()
}

func TestInsertClan(t *testing.T) {
	_ = config.Load(testConfigPath)
	ConnectMySQL()

	clan := Clan{
		OwnerId: 1,
		Name:    "Test",
		Tag:     "TEST",
	}

	err := clan.Insert()

	if err != nil {
		t.Fatal(err)
	}

	CloseMySQL()
}
