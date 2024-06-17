package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// GetRecentMultiplayerGames Retrieves the most recent multiplayer games from the database
// Endpoint: GET /v2/multiplayer/games
func GetRecentMultiplayerGames(c *gin.Context) *APIError {
	page, err := strconv.Atoi(c.Query("page"))

	if err != nil {
		page = 0
	}

	games, err := db.GetRecentMultiplayerGames(20, page)

	if err != nil {
		return APIErrorServerError("Error retrieving most recent multiplayer games", err)
	}

	c.JSON(http.StatusOK, gin.H{"games": games})
	return nil
}
