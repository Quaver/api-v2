package discord

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/Quaver/api2/config"
	"github.com/go-resty/resty/v2"
)

const unknownMemberCode = 10007

var ErrMemberNotFound = errors.New("Discord guild member not found")

type Client struct {
	http *resty.Client
}

func NewClient(httpClient *http.Client) *Client {
	client := resty.NewWithClient(httpClient).
		SetBaseURL("https://discord.com/api/v10").
		SetTimeout(10 * time.Second).
		SetRedirectPolicy(resty.NoRedirectPolicy())
	return &Client{http: client}
}

func ValidID(value string) bool {
	id, err := strconv.ParseUint(value, 10, 64)
	return err == nil && id != 0 && strconv.FormatUint(id, 10) == value
}

func (client *Client) AddDonatorRole(ctx context.Context, memberID string) error {
	return client.changeDonatorRole(ctx, http.MethodPut, memberID)
}

func (client *Client) RemoveDonatorRole(ctx context.Context, memberID string) error {
	return client.changeDonatorRole(ctx, http.MethodDelete, memberID)
}

func (client *Client) changeDonatorRole(ctx context.Context, method, memberID string) error {
	roleID := config.Instance.Discord.DonatorRoleID
	if !ValidID(roleID) {
		return errors.New("invalid Discord donator role ID")
	}
	_, err := client.requestMember(ctx, method, memberID, "/roles/"+roleID, http.StatusNoContent)
	return err
}

func (client *Client) HasPremiumRole(ctx context.Context, memberID string) (bool, error) {
	roleID := config.Instance.Discord.PremiumMemberRoleID
	if !ValidID(roleID) {
		return false, errors.New("invalid Discord premium member role ID")
	}

	body, err := client.requestMember(ctx, http.MethodGet, memberID, "", http.StatusOK)
	if err != nil {
		return false, err
	}

	var member struct {
		Roles []string `json:"roles"`
	}
	if err := json.Unmarshal(body, &member); err != nil {
		return false, fmt.Errorf("decode Discord member: %w", err)
	}
	if member.Roles == nil {
		return false, errors.New("Discord member response is missing roles")
	}
	return slices.Contains(member.Roles, roleID), nil
}

func (client *Client) requestMember(ctx context.Context, method, memberID, suffix string, expectedStatus int) ([]byte, error) {
	if !ValidID(memberID) {
		return nil, errors.New("invalid Discord member ID")
	}
	settings := config.Instance.Discord
	if settings.BotToken == "" || !ValidID(settings.GuildID) {
		return nil, errors.New("missing Discord bot token or invalid guild ID")
	}

	path := fmt.Sprintf("/guilds/%s/members/%s%s", settings.GuildID, memberID, suffix)
	response, err := client.http.R().
		SetContext(ctx).
		SetAuthScheme("Bot").
		SetAuthToken(settings.BotToken).
		Execute(method, path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() == expectedStatus {
		return response.Body(), nil
	}

	var discordError struct {
		Code int `json:"code"`
	}
	_ = json.Unmarshal(response.Body(), &discordError)
	if response.StatusCode() == http.StatusNotFound && discordError.Code == unknownMemberCode {
		return nil, ErrMemberNotFound
	}
	return nil, fmt.Errorf("Discord %s returned HTTP %d (code %d)", method, response.StatusCode(), discordError.Code)
}
