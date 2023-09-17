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
