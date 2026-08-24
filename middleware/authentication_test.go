package middleware

import (
	"testing"

	"github.com/Quaver/api2/config"
	"github.com/golang-jwt/jwt/v5"
)

func TestAuthenticateJWTRejectsInvalidSignatureBeforeReadingClaims(t *testing.T) {
	previousConfig := config.Instance
	config.Instance = &config.Config{JWTSecret: "global-secret"}
	defer func() {
		config.Instance = previousConfig
	}()

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTClaims{
		UserId:   1,
		Username: "user",
	}).SignedString([]byte("different-secret"))

	if err != nil {
		t.Fatalf("could not create test token: %v", err)
	}

	user, err := authenticateJWT("Bearer " + token)

	if err != nil {
		t.Fatalf("expected an invalid signature to be treated as unauthenticated, got %v", err)
	}

	if user != nil {
		t.Fatalf("expected an invalid signature to produce no authenticated user, got %+v", user)
	}
}
