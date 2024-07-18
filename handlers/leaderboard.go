package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// GetGlobalLeaderboardForMode Retrieves the global leaderboard for a given game mode
// Endpoint: GET /v2/leaderboard/global?mode=&page=
func GetGlobalLeaderboardForMode(c *gin.Context) *APIError {
	mode, err := strconv.Atoi(c.Query("mode"))

	if err != nil {
		return APIErrorBadRequest("You must supply a valid `mode` query parameter.")
	}

	page, err := strconv.Atoi(c.Query("page"))

	if err != nil {
		page = 0
	}

	users, err := db.GetGlobalLeaderboard(enums.GameMode(mode), page, 50)

	if err != nil {
		return APIErrorServerError("Error retrieving users for global leaderboard", err)
	}

	userCount, err := db.GetTotalUnbannedUserCount()

	if err != nil {
		return APIErrorServerError("Error retrieving unbanned user count", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"total_users": userCount,
		"users":       users,
	})

	return nil
}

// GetCountryLeaderboard Retrieves the country leaderboard for a given country and mode
// Endpoint: GET /v2/leaderboard/country?country=&mode=&page=
func GetCountryLeaderboard(c *gin.Context) *APIError {
	country := c.Query("country")

	if country == "" {
		return APIErrorBadRequest("You must supply a valid `country` query parameter")
	}

	mode, err := strconv.Atoi(c.Query("mode"))

	if err != nil {
		return APIErrorBadRequest("You must supply a valid `mode` query parameter.")
	}

	page, err := strconv.Atoi(c.Query("page"))

	if err != nil {
		page = 0
	}

	users, err := db.GetCountryLeaderboard(country, enums.GameMode(mode), page, 50)

	if err != nil {
		return APIErrorServerError("Error retrieving users for country leaderboard", err)
	}

	userCount, err := db.GetCountryPlayerCountFromRedis(country)

	if err != nil {
		return APIErrorServerError("Error retrieving country player count from db", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"total_users": userCount,
		"users":       users,
	})

	return nil
}

// GetTotalHitsLeaderboard Retrieves the total hits leaderboard for a given game mode
// Endpoint: GET /v2/leaderboard/hits?mode=&page=
func GetTotalHitsLeaderboard(c *gin.Context) *APIError {
	page, err := strconv.Atoi(c.Query("page"))

	if err != nil {
		page = 0
	}

	users, err := db.GetTotalHitsLeaderboard(page, 50)

	if err != nil {
		return APIErrorServerError("Error retrieving users for total hits leaderboard", err)
	}

	userCount, err := db.GetTotalUnbannedUserCount()

	if err != nil {
		return APIErrorServerError("Error retrieving unbanned user count", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"total_users": userCount,
		"users":       users,
	})

	return nil
}
