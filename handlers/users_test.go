package handlers

import (
	"fmt"
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
		"default_cover":"cover.jpg",
		"notif_action_mapset":false,
		"default_mode":2
	}`))
	if err != nil {
		t.Fatal(err)
	}

	if information.Discord != "discord" || information.Twitter != "twitter" ||
		information.Twitch != "twitch" || information.Youtube != "youtube" ||
		information.DefaultCover != "cover.jpg" || information.NotifyMapsetActions ||
		information.DefaultMode != enums.GameModeKeys7 {
		t.Fatalf("unexpected information: %#v", information)
	}
}

func TestParseUserInformationAcceptsEmptyDefaultCover(t *testing.T) {
	information, err := parseUserInformation(strings.NewReader(`{"default_cover":""}`))
	if err != nil {
		t.Fatal(err)
	}

	if information.DefaultCover != "" {
		t.Fatalf("expected empty default cover, got %q", information.DefaultCover)
	}
}

func TestParseUserInformationAcceptsValuesUpTo100Characters(t *testing.T) {
	value := strings.Repeat("a", maxUserInformationValueLength)
	body := fmt.Sprintf(`{
		"discord":%q,
		"twitter":%q,
		"twitch":%q,
		"youtube":%q
	}`, value, value, value, value)

	if _, err := parseUserInformation(strings.NewReader(body)); err != nil {
		t.Fatal(err)
	}
}

func TestParseUserInformationRejectsValuesOver100Characters(t *testing.T) {
	value := strings.Repeat("a", maxUserInformationValueLength+1)
	fields := []string{"discord", "twitter", "twitch", "youtube"}

	for _, field := range fields {
		t.Run(field, func(t *testing.T) {
			body := fmt.Sprintf(`{"%s":%q}`, field, value)
			if _, err := parseUserInformation(strings.NewReader(body)); err == nil {
				t.Fatalf("expected %s to be rejected when longer than %d characters", field, maxUserInformationValueLength)
			}
		})
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
