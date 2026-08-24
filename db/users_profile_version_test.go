package db

import (
	"encoding/hex"
	"encoding/json"
	"testing"
)

func TestGenerateUserProfileVersion(t *testing.T) {
	first, err := generateUserProfileVersion()
	if err != nil {
		t.Fatal(err)
	}

	second, err := generateUserProfileVersion()
	if err != nil {
		t.Fatal(err)
	}

	if len(first) != userProfileVersionBytes*2 {
		t.Fatalf("expected profile version length %d, got %d", userProfileVersionBytes*2, len(first))
	}

	if _, err := hex.DecodeString(first); err != nil {
		t.Fatalf("profile version is not valid hexadecimal: %v", err)
	}

	if first == second {
		t.Fatal("expected consecutive profile versions to differ")
	}
}

func TestUserProfileVersionIsSerializedWhenEmpty(t *testing.T) {
	data, err := json.Marshal(&User{ProfileVersion: ""})
	if err != nil {
		t.Fatal(err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatal(err)
	}

	value, exists := payload["profile_version"]
	if !exists {
		t.Fatal("expected profile_version to be present in serialized user")
	}

	if value != "" {
		t.Fatalf("expected empty profile_version, got %v", value)
	}
}

func TestUserBeforeCreateLeavesProfileVersionEmpty(t *testing.T) {
	user := &User{}

	if err := user.BeforeCreate(nil); err != nil {
		t.Fatal(err)
	}

	if user.ProfileVersion != "" {
		t.Fatalf("expected profile_version to remain empty, got %q", user.ProfileVersion)
	}
}
