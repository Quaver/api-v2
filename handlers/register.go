package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

// RegisterNewUser Registers a new user account
// Endpoint: POST /v2/user
func RegisterNewUser(c *gin.Context) *APIError {
	if c.GetHeader("User-Agent") != "Quaver" {
		return APIErrorForbidden("You are not allowed to access this resource.")
	}

	body := struct {
		Username string `form:"username" json:"username" binding:"required"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	steamId, apiErr := authenticateSteamTicket(c)

	if apiErr != nil {
		return apiErr
	}

	user, err := db.GetUserBySteamId(steamId)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving user by Steam Id", err)
	}

	if user != nil {
		return APIErrorForbidden("You already have an account and cannot access this resource.")
	}

	usernameAvailable, err := db.IsUsernameAvailable(-1, body.Username)

	if err != nil {
		return APIErrorServerError("Error checking if username is available", err)
	}

	if !usernameAvailable {
		return APIErrorBadRequest("The username you have chosen is unavailable.")
	}

	newUser := &db.User{
		SteamId:        steamId,
		Username:       body.Username,
		TimeRegistered: time.Now().UnixMilli(),
		Allowed:        true,
		Privileges:     1,
		UserGroups:     enums.UserGroupNormal,
		MuteEndTime:    0,
		LatestActivity: time.Now().UnixMilli(),
		Country:        "XX",
		IP:             c.ClientIP(),
	}

	if err = newUser.Insert(); err != nil {
		return APIErrorServerError("Error inserting user", err)
	}

	c.JSON(200, gin.H{"message": "Your account has been successfully created."})
	return nil
}

// Authenticates a Steam ticket from an incoming request. Returns the user's steam id.
func authenticateSteamTicket(c *gin.Context) (string, *APIError) {
	body := struct {
		PTicket string `form:"p_ticket" json:"p_ticket" binding:"required"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return "", APIErrorBadRequest("Invalid request body")
	}

	resp, err := resty.New().R().
		SetQueryParams(map[string]string{
			"key":    config.Instance.Steam.PublisherKey,
			"appid":  strconv.Itoa(config.Instance.Steam.AppId),
			"ticket": strings.Replace(body.PTicket, "-", "", -1),
		}).
		Get("https://api.steampowered.com/ISteamUserAuth/AuthenticateUserTicket/v1/")

	if err != nil {
		return "", APIErrorServerError("Cannot complete steam ticket authentication request", err)
	}

	const failed string = "failed to authenticate steam ticket"

	if resp.IsError() {
		return "", APIErrorServerError("Steam auth failed", fmt.Errorf("%v %v - %v", failed, resp.StatusCode(), string(resp.Body())))
	}

	type authenticateSteamTicketResponse struct {
		Response struct {
			Params struct {
				Result          string `json:"result,omitempty"`
				SteamId         string `json:"steamid,omitempty"`
				OwnerSteamId    string `json:"ownersteamid,omitempty"`
				VacBanned       bool   `json:"vacbanned,omitempty"`
				PublisherBanned bool   `json:"publisherbanned,omitempty"`
			} `json:"params"`

			Error interface{} `json:"error,omitempty"`
		} `json:"response"`
	}

	var parsed authenticateSteamTicketResponse
	err = json.Unmarshal(resp.Body(), &parsed)

	if err != nil {
		return "", APIErrorServerError("Steam auth failed", fmt.Errorf("%v - json unmarshal - %v - %v", failed, err, string(resp.Body())))
	}

	if parsed.Response.Error != nil || parsed.Response.Params.Result != "OK" {
		return "", APIErrorServerError("Steam auth failed", fmt.Errorf("%v - invalid response result - %v", failed, string(resp.Body())))
	}

	if parsed.Response.Params.VacBanned {
		return "", APIErrorServerError("Steam auth failed", fmt.Errorf("%v - user is vac banned", failed))
	}

	if parsed.Response.Params.PublisherBanned {
		return "", APIErrorServerError("Steam auth failed", fmt.Errorf("%v - user is publisher banned", failed))
	}

	return parsed.Response.Params.SteamId, nil
}
