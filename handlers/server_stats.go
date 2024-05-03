package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetServerStats Retrieves statistics about Quaver
// Endpoint: /v2/server/stats
func GetServerStats(c *gin.Context) *APIError {
	onlineUsers, err := db.GetOnlineUserCountFromRedis()

	if err != nil {
		return APIErrorServerError("Error retrieving online user count from redis", err)
	}

	totalUsers, err := db.GetTotalUserCountFromRedis()

	if err != nil {
		return APIErrorServerError("Error retrieving total user count from redis", err)
	}

	totalMapsets, err := db.GetTotalMapsetCountFromRedis()

	if err != nil {
		return APIErrorServerError("Error retrieving total mapset count from redis", err)
	}

	totalScores, err := db.GetTotalScoreCountFromRedis()

	if err != nil {
		return APIErrorServerError("Error retrieving total score count from redis", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"online_users":  onlineUsers,
		"total_users":   totalUsers,
		"total_mapsets": totalMapsets,
		"total_scores":  totalScores,
	})
	return nil
}

// GetCountryPlayers Returns the countries and their total user counts
// Endpoint: /v2/server/stats/countries
func GetCountryPlayers(c *gin.Context) *APIError {
	countries, err := db.CacheCountryPlayersInRedis()

	if err != nil {
		return APIErrorServerError("Failed to cache country players in redis", err)
	}

	c.JSON(http.StatusOK, gin.H{"countries": countries})
	return nil
}
