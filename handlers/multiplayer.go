package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

	count, err := db.GetTotalMultiplayerGameCount()

	if err != nil {
		return APIErrorServerError("Error retrieving multiplayer game count", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"total_game_count": count,
		"games":            games,
	})

	return nil
}

// GetMultiplayerGame Gets an individual multiplayer game
// Endpoint: GET /v2/multiplayer/games/:id
func GetMultiplayerGame(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("You must provide a valid game id")
	}

	game, err := db.GetMultiplayerGame(id)

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving multiplayer game", err)
	}

	if game == nil {
		return APIErrorNotFound("Multiplayer Game")
	}

	c.JSON(http.StatusOK, gin.H{"game": game})
	return nil
}
