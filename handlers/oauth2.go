package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/stringutil"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	oauthIssuer                  = "quaver-api"
	oauthAudience                = "quaver-oauth2"
	oauthScopeRead               = "read"
	oauthAuthorizationCodePrefix = "quaver:oauth2:authorization_code:"

	oauthAuthorizationCodeLifetime = time.Minute
	oauthAccessTokenLifetime       = time.Hour
)

var errInvalidOAuthAccessToken = errors.New("invalid OAuth access token")
var errInvalidOAuthAuthorizationCode = errors.New("invalid OAuth authorization code")

type oauthAuthorizeRequest struct {
	ClientID     string `form:"client_id" json:"client_id"`
	RedirectURI  string `form:"redirect_uri" json:"redirect_uri"`
	ResponseType string `form:"response_type" json:"response_type"`
	State        string `form:"state" json:"state"`
}

type oauthTokenRequest struct {
	ClientID     string `form:"client_id" json:"client_id"`
	ClientSecret string `form:"client_secret" json:"client_secret"`
	GrantType    string `form:"grant_type" json:"grant_type"`
	Code         string `form:"code" json:"code"`
	RedirectURI  string `form:"redirect_uri" json:"redirect_uri"`
}

type oauthAuthorizationCode struct {
	ApplicationID int    `json:"application_id"`
	ClientID      string `json:"client_id"`
	UserID        int    `json:"user_id"`
	RedirectURI   string `json:"redirect_uri"`
}

type oauthAccessTokenClaims struct {
	ApplicationID int    `json:"application_id"`
	ClientID      string `json:"client_id"`
	GrantType     string `json:"grant_type"`
	Scope         string `json:"scope"`
	jwt.RegisteredClaims
}

type oauthAuthorizationCodeStore interface {
	Set(key string, value []byte, expiration time.Duration) error
	GetDel(key string) (string, error)
}

type redisOAuthAuthorizationCodeStore struct {
	client *redis.Client
}

func (store *redisOAuthAuthorizationCodeStore) Set(key string, value []byte, expiration time.Duration) error {
	return store.client.Set(db.RedisCtx, key, value, expiration).Err()
}

func (store *redisOAuthAuthorizationCodeStore) GetDel(key string) (string, error) {
	return store.client.GetDel(db.RedisCtx, key).Result()
}

type oauthErrorResponse struct {
	Status      int
	Code        string
	Description string
}

func (e *oauthErrorResponse) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Description)
}

func newOAuthError(status int, code string, description string) *oauthErrorResponse {
	return &oauthErrorResponse{
		Status:      status,
		Code:        code,
		Description: description,
	}
}

func respondOAuthError(c *gin.Context, response *oauthErrorResponse) {
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")

	if response.Status == http.StatusUnauthorized {
		c.Header("WWW-Authenticate", `Bearer realm="oauth2"`)
	}

	c.JSON(response.Status, gin.H{
		"error":             response.Code,
		"error_description": response.Description,
	})
}

func respondOAuthServerError(c *gin.Context, message string, err error) {
	logrus.Errorf("%s - %v", message, err)
	respondOAuthError(c, newOAuthError(http.StatusInternalServerError, "server_error", "The OAuth request could not be completed."))
}

func bindOAuthAuthorizeRequest(c *gin.Context) (oauthAuthorizeRequest, error) {
	request := oauthAuthorizeRequest{}

	if c.Request.Method == http.MethodGet {
		request.ClientID = c.Query("client_id")
		request.RedirectURI = c.Query("redirect_uri")
		request.ResponseType = c.Query("response_type")
		request.State = c.Query("state")
		return request, nil
	}

	if err := c.ShouldBind(&request); err != nil && !errors.Is(err, io.EOF) {
		return request, err
	}

	// Allow the approval request to be submitted with query parameters as well
	// as a JSON or form-encoded body.
	if request.ClientID == "" {
		request.ClientID = c.Query("client_id")
	}

	if request.RedirectURI == "" {
		request.RedirectURI = c.Query("redirect_uri")
	}

	if request.ResponseType == "" {
		request.ResponseType = c.Query("response_type")
	}

	if request.State == "" {
		request.State = c.Query("state")
	}

	return request, nil
}

func bindOAuthTokenRequest(c *gin.Context) (oauthTokenRequest, error) {
	request := oauthTokenRequest{}

	if err := c.ShouldBind(&request); err != nil && !errors.Is(err, io.EOF) {
		return request, err
	}

	return request, nil
}

// validateOAuthRedirectURL validates a registered or requested OAuth redirect URL.
func validateOAuthRedirectURL(value string) error {
	parsed, err := url.Parse(value)

	if err != nil || !parsed.IsAbs() || parsed.Host == "" || parsed.Fragment != "" || parsed.User != nil {
		return errors.New("redirect URL must be an absolute URL without a fragment")
	}

	if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
		return errors.New("redirect URL must use HTTP or HTTPS")
	}

	return nil
}

func loadOAuthAuthorizeApplication(request oauthAuthorizeRequest) (*db.Application, *url.URL, *oauthErrorResponse, error) {
	if strings.TrimSpace(request.ClientID) == "" {
		return nil, nil, newOAuthError(http.StatusBadRequest, "invalid_request", "client_id is required."), nil
	}

	if request.ResponseType != "code" {
		return nil, nil, newOAuthError(http.StatusBadRequest, "unsupported_response_type", "Only the code response type is supported."), nil
	}

	if request.RedirectURI == "" {
		return nil, nil, newOAuthError(http.StatusBadRequest, "invalid_request", "redirect_uri is required."), nil
	}

	if err := validateOAuthRedirectURL(request.RedirectURI); err != nil {
		return nil, nil, newOAuthError(http.StatusBadRequest, "invalid_request", "redirect_uri is invalid."), nil
	}

	application, err := db.GetActiveApplicationByClientId(request.ClientID)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, nil, err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) || application == nil {
		return nil, nil, newOAuthError(http.StatusUnauthorized, "invalid_client", "The client_id is invalid."), nil
	}

	if application.RedirectURL == "" {
		return nil, nil, newOAuthError(http.StatusBadRequest, "invalid_request", "This application does not have an authorization redirect URL."), nil
	}

	if err := validateOAuthRedirectURL(application.RedirectURL); err != nil {
		return nil, nil, newOAuthError(http.StatusBadRequest, "invalid_request", "The application redirect URL is invalid."), nil
	}

	if request.RedirectURI != application.RedirectURL {
		return nil, nil, newOAuthError(http.StatusBadRequest, "invalid_request", "redirect_uri does not match the registered application URL."), nil
	}

	redirect, err := url.Parse(application.RedirectURL)

	if err != nil {
		return nil, nil, nil, err
	}

	return application, redirect, nil, nil
}

// GetOAuthAuthorization returns the safe application metadata needed by the consent UI.
// Endpoint: GET /v2/oauth2/authorize
func GetOAuthAuthorization(c *gin.Context) {
	request, err := bindOAuthAuthorizeRequest(c)

	if err != nil {
		respondOAuthError(c, newOAuthError(http.StatusBadRequest, "invalid_request", "The authorization request is invalid."))
		return
	}

	application, redirect, response, err := loadOAuthAuthorizeApplication(request)

	if err != nil {
		respondOAuthServerError(c, "Error validating OAuth authorization request", err)
		return
	}

	if response != nil {
		respondOAuthError(c, response)
		return
	}

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{
		"client_id":    application.ClientId,
		"redirect_uri": application.RedirectURL,
		"site":         redirect.Hostname(),
		"application": gin.H{
			"id":   application.Id,
			"name": application.Name,
		},
	})
}

// ApproveOAuthAuthorization creates a one-time authorization code and redirects the user.
// Endpoint: POST /v2/oauth2/authorize
func ApproveOAuthAuthorization(c *gin.Context) {
	request, err := bindOAuthAuthorizeRequest(c)

	if err != nil {
		respondOAuthError(c, newOAuthError(http.StatusBadRequest, "invalid_request", "The authorization request is invalid."))
		return
	}

	application, redirect, response, err := loadOAuthAuthorizeApplication(request)

	if err != nil {
		respondOAuthServerError(c, "Error validating OAuth approval request", err)
		return
	}

	if response != nil {
		respondOAuthError(c, response)
		return
	}

	user := getAuthedUser(c)

	if user == nil {
		respondOAuthError(c, newOAuthError(http.StatusUnauthorized, "access_denied", "An authenticated user is required to approve this application."))
		return
	}

	code, err := stringutil.GenerateToken(32)

	if err != nil {
		respondOAuthServerError(c, "Error generating OAuth authorization code", err)
		return
	}

	codeStore, err := getOAuthAuthorizationCodeStore()

	if err != nil {
		respondOAuthServerError(c, "Error storing OAuth authorization code", err)
		return
	}

	if err := storeOAuthAuthorizationCode(codeStore, code, oauthAuthorizationCode{
		ApplicationID: application.Id,
		ClientID:      application.ClientId,
		UserID:        user.Id,
		RedirectURI:   application.RedirectURL,
	}); err != nil {
		respondOAuthServerError(c, "Error storing OAuth authorization code", err)
		return
	}

	if err := db.RecordApplicationUsage(application.Id); err != nil {
		logrus.WithError(err).WithField("application_id", application.Id).Warn("Unable to record OAuth application usage")
	}

	query := redirect.Query()
	query.Set("code", code)

	if request.State != "" {
		query.Set("state", request.State)
	}

	redirect.RawQuery = query.Encode()
	http.Redirect(c.Writer, c.Request, redirect.String(), http.StatusFound)
}

func oauthAuthorizationCodeKey(code string) string {
	return oauthAuthorizationCodePrefix + code
}

func getOAuthAuthorizationCodeStore() (oauthAuthorizationCodeStore, error) {
	if db.Redis == nil {
		return nil, errors.New("redis is not initialized")
	}

	return &redisOAuthAuthorizationCodeStore{client: db.Redis}, nil
}

func storeOAuthAuthorizationCode(store oauthAuthorizationCodeStore, code string, authorizationCode oauthAuthorizationCode) error {
	codeData, err := json.Marshal(authorizationCode)

	if err != nil {
		return err
	}

	return store.Set(oauthAuthorizationCodeKey(code), codeData, oauthAuthorizationCodeLifetime)
}

func consumeOAuthAuthorizationCode(store oauthAuthorizationCodeStore, code string) (oauthAuthorizationCode, error) {
	codeData, err := store.GetDel(oauthAuthorizationCodeKey(code))

	if err != nil {
		return oauthAuthorizationCode{}, err
	}

	storedCode := oauthAuthorizationCode{}

	if err := json.Unmarshal([]byte(codeData), &storedCode); err != nil {
		return oauthAuthorizationCode{}, errInvalidOAuthAuthorizationCode
	}

	return storedCode, nil
}

func getActiveOAuthApplication(clientID string, clientSecret string) (*db.Application, *oauthErrorResponse, error) {
	if strings.TrimSpace(clientID) == "" {
		return nil, newOAuthError(http.StatusBadRequest, "invalid_request", "client_id is required."), nil
	}

	if clientSecret == "" {
		return nil, newOAuthError(http.StatusUnauthorized, "invalid_client", "client_secret is required."), nil
	}

	application, err := db.GetActiveApplicationByClientIdAndSecret(clientID, clientSecret)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) || application == nil {
		return nil, newOAuthError(http.StatusUnauthorized, "invalid_client", "The client credentials are invalid."), nil
	}

	return application, nil, nil
}

// OAuthToken issues an OAuth access token for either supported grant.
// Endpoint: POST /v2/oauth2/token
func OAuthToken(c *gin.Context) {
	request, err := bindOAuthTokenRequest(c)

	if err != nil {
		respondOAuthError(c, newOAuthError(http.StatusBadRequest, "invalid_request", "The token request is invalid."))
		return
	}

	request.ClientID = strings.TrimSpace(request.ClientID)
	request.GrantType = strings.TrimSpace(request.GrantType)
	request.Code = strings.TrimSpace(request.Code)
	request.RedirectURI = strings.TrimSpace(request.RedirectURI)

	// The original OAuth documentation used client_credentials for the
	// authorization-code example. Accept that form only when a code is present.
	grantType := request.GrantType
	if grantType == "client_credentials" && request.Code != "" {
		grantType = "authorization_code"
	}

	if grantType != "authorization_code" && grantType != "client_credentials" {
		respondOAuthError(c, newOAuthError(http.StatusBadRequest, "unsupported_grant_type", "The requested grant type is not supported."))
		return
	}

	application, response, err := getActiveOAuthApplication(request.ClientID, request.ClientSecret)

	if err != nil {
		respondOAuthServerError(c, "Error authenticating OAuth client", err)
		return
	}

	if response != nil {
		respondOAuthError(c, response)
		return
	}

	var userID *int

	if grantType == "authorization_code" {
		if request.Code == "" || request.RedirectURI == "" {
			respondOAuthError(c, newOAuthError(http.StatusBadRequest, "invalid_request", "code and redirect_uri are required for the authorization-code grant."))
			return
		}

		if request.RedirectURI != application.RedirectURL {
			respondOAuthError(c, newOAuthError(http.StatusBadRequest, "invalid_grant", "redirect_uri does not match the registered application URL."))
			return
		}

		if err := validateOAuthRedirectURL(request.RedirectURI); err != nil {
			respondOAuthError(c, newOAuthError(http.StatusBadRequest, "invalid_grant", "redirect_uri is invalid."))
			return
		}

		codeStore, err := getOAuthAuthorizationCodeStore()

		if err != nil {
			respondOAuthServerError(c, "Error consuming OAuth authorization code", err)
			return
		}

		storedCode, err := consumeOAuthAuthorizationCode(codeStore, request.Code)

		if errors.Is(err, redis.Nil) {
			respondOAuthError(c, newOAuthError(http.StatusBadRequest, "invalid_grant", "The authorization code is invalid, expired, or already used."))
			return
		}

		if errors.Is(err, errInvalidOAuthAuthorizationCode) {
			respondOAuthError(c, newOAuthError(http.StatusBadRequest, "invalid_grant", "The authorization code is invalid."))
			return
		}

		if err != nil {
			respondOAuthServerError(c, "Error consuming OAuth authorization code", err)
			return
		}

		if storedCode.ApplicationID != application.Id ||
			storedCode.ClientID != application.ClientId ||
			storedCode.RedirectURI != application.RedirectURL ||
			storedCode.UserID <= 0 {
			respondOAuthError(c, newOAuthError(http.StatusBadRequest, "invalid_grant", "The authorization code does not belong to this client."))
			return
		}

		user, err := db.GetUserById(storedCode.UserID)

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			respondOAuthServerError(c, "Error retrieving OAuth user", err)
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) || user == nil || !user.Allowed {
			respondOAuthError(c, newOAuthError(http.StatusBadRequest, "invalid_grant", "The user authorization is no longer valid."))
			return
		}

		userID = &storedCode.UserID
	}

	accessToken, err := signOAuthAccessToken(application, grantType, userID)

	if err != nil {
		respondOAuthServerError(c, "Error signing OAuth access token", err)
		return
	}

	if err := db.RecordApplicationUsage(application.Id); err != nil {
		logrus.WithError(err).WithField("application_id", application.Id).Warn("Unable to record OAuth application usage")
	}

	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	c.JSON(http.StatusOK, gin.H{
		"token_type":   "Bearer",
		"expires_in":   int(oauthAccessTokenLifetime.Seconds()),
		"access_token": accessToken,
		"scope":        oauthScopeRead,
	})
}

func oauthAccessTokenSigningKey() ([]byte, error) {
	if config.Instance == nil || strings.TrimSpace(config.Instance.OAuthAccessTokenSecret) == "" {
		return nil, errors.New("OAuth access token signing secret is not configured")
	}

	return []byte(config.Instance.OAuthAccessTokenSecret), nil
}

func signOAuthAccessToken(application *db.Application, grantType string, userID *int) (string, error) {
	if application == nil || application.ClientId == "" {
		return "", errors.New("OAuth application is not configured")
	}

	signingKey, err := oauthAccessTokenSigningKey()

	if err != nil {
		return "", err
	}

	jti, err := stringutil.GenerateToken(16)

	if err != nil {
		return "", err
	}

	now := time.Now()
	subject := ""

	if userID != nil {
		subject = strconv.Itoa(*userID)
	}

	claims := oauthAccessTokenClaims{
		ApplicationID: application.Id,
		ClientID:      application.ClientId,
		GrantType:     grantType,
		Scope:         oauthScopeRead,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    oauthIssuer,
			Subject:   subject,
			Audience:  jwt.ClaimStrings{oauthAudience},
			ExpiresAt: jwt.NewNumericDate(now.Add(oauthAccessTokenLifetime)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(signingKey)
}

func parseUnverifiedOAuthAccessToken(rawToken string) (*oauthAccessTokenClaims, error) {
	rawToken = strings.TrimSpace(rawToken)

	if rawToken == "" {
		return nil, errInvalidOAuthAccessToken
	}

	unverifiedClaims := oauthAccessTokenClaims{}
	unverifiedToken, _, err := new(jwt.Parser).ParseUnverified(rawToken, &unverifiedClaims)

	if err != nil || unverifiedToken == nil || unverifiedToken.Method == nil ||
		unverifiedToken.Method.Alg() != jwt.SigningMethodHS256.Alg() ||
		unverifiedClaims.ClientID == "" {
		return nil, errInvalidOAuthAccessToken
	}

	return &unverifiedClaims, nil
}

func verifyOAuthAccessTokenWithApplication(rawToken string, application *db.Application) (*oauthAccessTokenClaims, error) {
	if application == nil || !application.Active {
		return nil, errInvalidOAuthAccessToken
	}

	unverifiedClaims, err := parseUnverifiedOAuthAccessToken(rawToken)

	if err != nil || unverifiedClaims.ClientID != application.ClientId {
		return nil, errInvalidOAuthAccessToken
	}

	signingKey, err := oauthAccessTokenSigningKey()

	if err != nil {
		return nil, err
	}

	claims := oauthAccessTokenClaims{}
	token, err := jwt.ParseWithClaims(
		rawToken,
		&claims,
		func(token *jwt.Token) (interface{}, error) {
			if token.Method == nil || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, errInvalidOAuthAccessToken
			}

			return signingKey, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(oauthIssuer),
		jwt.WithAudience(oauthAudience),
	)

	if err != nil || token == nil || !token.Valid {
		return nil, errInvalidOAuthAccessToken
	}

	if claims.ApplicationID != application.Id || claims.ClientID != application.ClientId ||
		claims.Scope != oauthScopeRead || claims.GrantType == "" || claims.ID == "" {
		return nil, errInvalidOAuthAccessToken
	}

	return &claims, nil
}

func verifyOAuthAccessToken(rawToken string) (*oauthAccessTokenClaims, *db.Application, error) {
	unverifiedClaims, err := parseUnverifiedOAuthAccessToken(rawToken)

	if err != nil {
		return nil, nil, errInvalidOAuthAccessToken
	}

	application, err := db.GetActiveApplicationByClientId(unverifiedClaims.ClientID)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, fmt.Errorf("error retrieving OAuth application: %w", err)
	}

	if errors.Is(err, gorm.ErrRecordNotFound) || application == nil {
		return nil, nil, errInvalidOAuthAccessToken
	}

	claims, err := verifyOAuthAccessTokenWithApplication(rawToken, application)

	if err != nil {
		return nil, nil, err
	}

	return claims, application, nil
}

func getOAuthAccessTokenFromRequest(c *gin.Context) (string, *oauthErrorResponse) {
	authorization := strings.TrimSpace(c.GetHeader("Authorization"))

	if authorization != "" {
		parts := strings.Fields(authorization)

		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			return "", newOAuthError(http.StatusUnauthorized, "invalid_token", "The bearer token is invalid.")
		}

		return parts[1], nil
	}

	body := struct {
		Code string `form:"code" json:"code"`
	}{}

	if err := c.ShouldBind(&body); err != nil && !errors.Is(err, io.EOF) {
		return "", newOAuthError(http.StatusBadRequest, "invalid_request", "The request body is invalid.")
	}

	return strings.TrimSpace(body.Code), nil
}

// OAuthMe verifies an OAuth access token and returns the current user.
// Endpoint: POST /v2/oauth2/me
func OAuthMe(c *gin.Context) {
	rawToken, response := getOAuthAccessTokenFromRequest(c)

	if response != nil {
		respondOAuthError(c, response)
		return
	}

	if rawToken == "" {
		respondOAuthError(c, newOAuthError(http.StatusUnauthorized, "invalid_token", "An OAuth access token is required."))
		return
	}

	claims, _, err := verifyOAuthAccessToken(rawToken)

	if errors.Is(err, errInvalidOAuthAccessToken) {
		respondOAuthError(c, newOAuthError(http.StatusUnauthorized, "invalid_token", "The OAuth access token is invalid or expired."))
		return
	}

	if err != nil {
		respondOAuthServerError(c, "Error verifying OAuth access token", err)
		return
	}

	if claims.GrantType != "authorization_code" || claims.Subject == "" {
		respondOAuthError(c, newOAuthError(http.StatusUnauthorized, "invalid_token", "The access token does not represent a user."))
		return
	}

	userID, err := strconv.Atoi(claims.Subject)

	if err != nil || userID <= 0 {
		respondOAuthError(c, newOAuthError(http.StatusUnauthorized, "invalid_token", "The access token does not represent a valid user."))
		return
	}

	user, err := db.GetUserById(userID)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		respondOAuthServerError(c, "Error retrieving OAuth user", err)
		return
	}

	if errors.Is(err, gorm.ErrRecordNotFound) || user == nil {
		respondOAuthError(c, newOAuthError(http.StatusUnauthorized, "invalid_token", "The user for this access token no longer exists."))
		return
	}

	if !user.Allowed {
		respondOAuthError(c, newOAuthError(http.StatusForbidden, "access_denied", "The user is not allowed to access this resource."))
		return
	}

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"user": user,
	})
}
