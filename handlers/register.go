package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/Quaver/api2/config"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"strconv"
	"strings"
)

// RegisterNewUser Registers a new user account
// Endpoint: POST /v2/user
func RegisterNewUser(c *gin.Context) *APIError {
	if c.GetHeader("User-Agent") != "Quaver" {
		return APIErrorForbidden("You are not allowed to access this resource.")
	}

	return nil
}

// Authenticates a Steam ticket from an incoming request. Returns the user's steam id.
func authenticateSteamTicket(c *gin.Context) (string, error) {
	body := struct {
		PTicket string `form:"p_ticket" json:"p_ticket" binding:"required"`
	}{}

	resp, err := resty.New().R().
		SetQueryParams(map[string]string{
			"key":    config.Instance.Steam.PublisherKey,
			"appid":  strconv.Itoa(config.Instance.Steam.AppId),
			"ticket": strings.Replace(body.PTicket, "-", "", -1),
		}).
		Get("https://api.steampowered.com/ISteamUserAuth/AuthenticateUserTicket/v1/")

	if err != nil {
		return "", err
	}

	const failed string = "failed to authenticate steam ticket"

	if resp.IsError() {
		return "", fmt.Errorf("%v %v - %v", failed, resp.StatusCode(), string(resp.Body()))
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
		return "", fmt.Errorf("%v - json unmarshal - %v - %v", failed, err, string(resp.Body()))
	}

	if parsed.Response.Error != nil || parsed.Response.Params.Result != "OK" {
		return "", fmt.Errorf("%v - invalid response result - %v", failed, string(resp.Body()))
	}

	if parsed.Response.Params.VacBanned {
		return "", fmt.Errorf("%v - user is vac banned", failed)
	}

	if parsed.Response.Params.PublisherBanned {
		return "", fmt.Errorf("%v - user is publisher banned", failed)
	}

	return parsed.Response.Params.SteamId, nil
}
