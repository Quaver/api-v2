package db

import (
	"encoding/json"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/enums"
	"gorm.io/gorm"
	"reflect"
	"testing"
)

func TestUserInformationJSONOmitsEmptyFields(t *testing.T) {
	marshaled, err := json.Marshal(UserInformation{
		Discord:             "discord",
		NotifyMapsetActions: false,
		DefaultMode:         enums.GameModeKeys7,
	})
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(marshaled, &got); err != nil {
		t.Fatal(err)
	}

	expected := map[string]any{
		"discord":      "discord",
		"default_mode": float64(enums.GameModeKeys7),
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %#v, got %#v", expected, got)
	}
}

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
