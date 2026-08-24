package handlers

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type testOAuthAuthorizationCodeStore struct {
	values      map[string]string
	expirations map[string]time.Duration
}

func newTestOAuthAuthorizationCodeStore() *testOAuthAuthorizationCodeStore {
	return &testOAuthAuthorizationCodeStore{
		values:      make(map[string]string),
		expirations: make(map[string]time.Duration),
	}
}

func (store *testOAuthAuthorizationCodeStore) Set(key string, value []byte, expiration time.Duration) error {
	store.values[key] = string(value)
	store.expirations[key] = expiration
	return nil
}

func (store *testOAuthAuthorizationCodeStore) GetDel(key string) (string, error) {
	value, exists := store.values[key]

	if !exists {
		return "", redis.Nil
	}

	delete(store.values, key)
	delete(store.expirations, key)
	return value, nil
}

func TestValidateOAuthRedirectURL(t *testing.T) {
	validURLs := []string{
		"https://example.com/oauth/callback",
		"http://localhost:8080/oauth/callback?source=quaver",
	}

	for _, value := range validURLs {
		if err := validateOAuthRedirectURL(value); err != nil {
			t.Fatalf("expected %q to be valid: %v", value, err)
		}
	}

	invalidURLs := []string{
		"/oauth/callback",
		"javascript:alert(1)",
		"https://example.com/oauth/callback#fragment",
		"https://user:password@example.com/oauth/callback",
	}

	for _, value := range invalidURLs {
		if err := validateOAuthRedirectURL(value); err == nil {
			t.Fatalf("expected %q to be invalid", value)
		}
	}
}

func TestOAuthAuthorizationCodeIsOneTimeAndShortLived(t *testing.T) {
	store := newTestOAuthAuthorizationCodeStore()
	code := "authorization-code"
	want := oauthAuthorizationCode{
		ApplicationID: 4,
		ClientID:      "client-id",
		UserID:        23,
		RedirectURI:   "https://example.com/callback",
	}

	if err := storeOAuthAuthorizationCode(store, code, want); err != nil {
		t.Fatalf("storeOAuthAuthorizationCode returned an error: %v", err)
	}

	if got := store.expirations[oauthAuthorizationCodeKey(code)]; got != oauthAuthorizationCodeLifetime {
		t.Fatalf("expected authorization code lifetime %v, got %v", oauthAuthorizationCodeLifetime, got)
	}

	got, err := consumeOAuthAuthorizationCode(store, code)

	if err != nil {
		t.Fatalf("consumeOAuthAuthorizationCode returned an error: %v", err)
	}

	if got != want {
		t.Fatalf("expected authorization code %+v, got %+v", want, got)
	}

	_, err = consumeOAuthAuthorizationCode(store, code)

	if !errors.Is(err, redis.Nil) {
		t.Fatalf("expected a reused code to return redis.Nil, got %v", err)
	}
}

func TestOAuthAccessTokenSigningRequiresServerSecret(t *testing.T) {
	previousConfig := config.Instance
	config.Instance = nil
	t.Cleanup(func() {
		config.Instance = previousConfig
	})

	_, err := signOAuthAccessToken(&db.Application{
		ClientId: "client-id",
		Active:   true,
	}, "client_credentials", nil)

	if err == nil {
		t.Fatal("expected OAuth access-token signing to fail without the server secret")
	}
}

func TestOAuthAccessTokenClaimsAreBoundToApplication(t *testing.T) {
	previousConfig := config.Instance
	config.Instance = &config.Config{
		JWTSecret:              "user-jwt-signing-secret",
		OAuthAccessTokenSecret: "oauth-access-token-signing-secret",
	}
	t.Cleanup(func() {
		config.Instance = previousConfig
	})

	application := &db.Application{
		Id:           7,
		ClientId:     "client-id",
		ClientSecret: "client-secret",
		Active:       true,
	}
	userID := 23

	token, err := signOAuthAccessToken(application, "authorization_code", &userID)

	if err != nil {
		t.Fatalf("signOAuthAccessToken returned an error: %v", err)
	}

	claims, err := verifyOAuthAccessTokenWithApplication(token, application)

	if err != nil {
		t.Fatalf("verifyOAuthAccessTokenWithApplication returned an error: %v", err)
	}

	if claims.Subject != "23" || claims.ClientID != application.ClientId || claims.ApplicationID != application.Id {
		t.Fatalf("unexpected OAuth claims: %+v", claims)
	}

	rotatedClientSecretApplication := &db.Application{
		Id:           application.Id,
		ClientId:     application.ClientId,
		ClientSecret: "different-secret",
		Active:       true,
	}

	if _, err := verifyOAuthAccessTokenWithApplication(token, rotatedClientSecretApplication); err != nil {
		t.Fatalf("expected client-secret rotation not to change the OAuth signing key, got %v", err)
	}

	inactiveApplication := *application
	inactiveApplication.Active = false

	if _, err := verifyOAuthAccessTokenWithApplication(token, &inactiveApplication); !errors.Is(err, errInvalidOAuthAccessToken) {
		t.Fatalf("expected a token for an inactive application to be rejected, got %v", err)
	}

	signingKey, err := oauthAccessTokenSigningKey()
	if err != nil {
		t.Fatalf("could not load OAuth access-token signing key: %v", err)
	}

	expiredToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, oauthAccessTokenClaims{
		ApplicationID: application.Id,
		ClientID:      application.ClientId,
		GrantType:     "authorization_code",
		Scope:         oauthScopeRead,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    oauthIssuer,
			Subject:   "23",
			Audience:  jwt.ClaimStrings{oauthAudience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Minute)),
			ID:        "expired-token",
		},
	}).SignedString(signingKey)

	if err != nil {
		t.Fatalf("could not create expired access token: %v", err)
	}

	if _, err := verifyOAuthAccessTokenWithApplication(expiredToken, application); !errors.Is(err, errInvalidOAuthAccessToken) {
		t.Fatalf("expected an expired access token to be rejected, got %v", err)
	}

	forgedClaims := oauthAccessTokenClaims{
		ApplicationID: application.Id,
		ClientID:      application.ClientId,
		GrantType:     "authorization_code",
		Scope:         oauthScopeRead,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    oauthIssuer,
			Subject:   "24",
			Audience:  jwt.ClaimStrings{oauthAudience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        "forged-token",
		},
	}

	for _, test := range []struct {
		name string
		key  string
	}{
		{name: "client secret", key: application.ClientSecret},
		{name: "user JWT secret", key: config.Instance.JWTSecret},
	} {
		forgedToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, forgedClaims).SignedString([]byte(test.key))

		if err != nil {
			t.Fatalf("could not create token signed with the %s: %v", test.name, err)
		}

		if _, err := verifyOAuthAccessTokenWithApplication(forgedToken, application); !errors.Is(err, errInvalidOAuthAccessToken) {
			t.Fatalf("expected a token signed with the %s to be rejected, got %v", test.name, err)
		}
	}

	clientToken, err := signOAuthAccessToken(application, "client_credentials", nil)

	if err != nil {
		t.Fatalf("signOAuthAccessToken returned an error for client credentials: %v", err)
	}

	clientClaims, err := verifyOAuthAccessTokenWithApplication(clientToken, application)

	if err != nil {
		t.Fatalf("client-credentials token could not be verified: %v", err)
	}

	if clientClaims.Subject != "" || clientClaims.GrantType != "client_credentials" {
		t.Fatalf("unexpected client-credentials claims: %+v", clientClaims)
	}
}

func TestGetOAuthAccessTokenFromRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	be := func(request *httptest.ResponseRecorder, method string, body string, contentType string) *gin.Context {
		ginContext, _ := gin.CreateTestContext(request)
		ginContext.Request = httptest.NewRequest(method, "/v2/oauth2/me", strings.NewReader(body))
		if contentType != "" {
			ginContext.Request.Header.Set("Content-Type", contentType)
		}
		return ginContext
	}

	bearerRecorder := httptest.NewRecorder()
	bearerContext := be(bearerRecorder, "POST", "", "")
	bearerContext.Request.Header.Set("Authorization", "Bearer bearer-token")

	token, response := getOAuthAccessTokenFromRequest(bearerContext)

	if response != nil || token != "bearer-token" {
		t.Fatalf("expected bearer token extraction, got token=%q response=%v", token, response)
	}

	bodyRecorder := httptest.NewRecorder()
	bodyContext := be(bodyRecorder, "POST", `{"code":"body-token"}`, "application/json")
	token, response = getOAuthAccessTokenFromRequest(bodyContext)

	if response != nil || token != "body-token" {
		t.Fatalf("expected body token extraction, got token=%q response=%v", token, response)
	}

	invalidRecorder := httptest.NewRecorder()
	invalidContext := be(invalidRecorder, "POST", "", "")
	invalidContext.Request.Header.Set("Authorization", "Basic credentials")
	_, response = getOAuthAccessTokenFromRequest(invalidContext)

	if response == nil || response.Code != "invalid_token" {
		t.Fatalf("expected invalid bearer token response, got %v", response)
	}
}
