package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"strings"
	"testing"
)

func TestParseUserInformationAppliesDefaults(t *testing.T) {
	information, err := parseUserInformation(strings.NewReader(`{"discord":"user#1234"}`))
	if err != nil {
		t.Fatal(err)
	}

	expected := db.UserInformation{
		Discord:             "user#1234",
		NotifyMapsetActions: true,
		DefaultMode:         enums.GameModeKeys4,
	}

	if information != expected {
		t.Fatalf("expected %#v, got %#v", expected, information)
	}
}

func TestParseUserInformationAcceptsAllFields(t *testing.T) {
	information, err := parseUserInformation(strings.NewReader(`{
		"discord":"discord",
		"twitter":"twitter",
		"twitch":"twitch",
		"youtube":"youtube",
		"notif_action_mapset":false,
		"default_mode":2
	}`))
	if err != nil {
		t.Fatal(err)
	}

	if information.Discord != "discord" || information.Twitter != "twitter" ||
		information.Twitch != "twitch" || information.Youtube != "youtube" ||
		information.NotifyMapsetActions || information.DefaultMode != enums.GameModeKeys7 {
		t.Fatalf("unexpected information: %#v", information)
	}
}

func TestParseUserInformationRejectsInvalidBodies(t *testing.T) {
	tests := []string{
		`null`,
		`[]`,
		`{"unknown":"value"}`,
		`{"discord":null}`,
		`{"discord":123}`,
		`{"notif_action_mapset":"true"}`,
		`{"default_mode":0}`,
		`{"default_mode":3}`,
		`{"default_mode":11}`,
		`{"discord":"discord"} {}`,
	}

	for _, body := range tests {
		t.Run(body, func(t *testing.T) {
			if _, err := parseUserInformation(strings.NewReader(body)); err == nil {
				t.Fatalf("expected body to be rejected: %s", body)
			}
		})
	}
}
