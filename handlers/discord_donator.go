package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/Quaver/api2/internal/discord"
	"github.com/gin-gonic/gin"
)

var discordDonatorClient = discord.NewClient(&http.Client{})

func AddDiscordDonatorRole(c *gin.Context) *APIError {
	return changeDiscordDonatorRole(c, discordDonatorClient.AddDonatorRole)
}

func RemoveDiscordDonatorRole(c *gin.Context) *APIError {
	return changeDiscordDonatorRole(c, discordDonatorClient.RemoveDonatorRole)
}

func changeDiscordDonatorRole(c *gin.Context, changeRole func(context.Context, string) error) *APIError {
	var body struct {
		ID string `json:"id" form:"id"`
	}
	if err := c.ShouldBind(&body); err != nil && !errors.Is(err, io.EOF) {
		return APIErrorBadRequest("Invalid Discord `id`")
	}
	if err := validateDiscordDonatorID(body.ID); err != nil {
		return err
	}
	if err := changeRole(c.Request.Context(), body.ID); err != nil {
		return discordDonatorRequestError(err)
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "OK!"})
	return nil
}

func CheckDiscordPremiumRole(c *gin.Context) *APIError {
	id := c.Param("id")
	if err := validateDiscordDonatorID(id); err != nil {
		return err
	}
	hasPremiumRole, err := discordDonatorClient.HasPremiumRole(c.Request.Context(), id)
	if err != nil {
		return discordDonatorRequestError(err)
	}

	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "has_donator": hasPremiumRole})
	return nil
}

func validateDiscordDonatorID(id string) *APIError {
	if id == "" {
		return APIErrorBadRequest("Missing Discord `id`")
	}
	if !discord.ValidID(id) {
		return APIErrorBadRequest("Invalid Discord `id`")
	}
	return nil
}

func discordDonatorRequestError(err error) *APIError {
	if errors.Is(err, discord.ErrMemberNotFound) {
		return APIErrorNotFound("User")
	}

	return APIErrorServerError("Discord donator request failed", err)
}
