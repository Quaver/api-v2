package handlers

import "github.com/gin-gonic/gin"

// GetClans Retrieves basic info / leaderboard data about clans
// GET /v2/clans?page=1
func GetClans(c *gin.Context) {
}

// CreateClan Creates a new clan if the user is eligible to.
// POST /v2/clan
func CreateClan(c *gin.Context) {
	// TODO: Authenticate user
}

// GetClan Retrieves data about an individual clan
// GET /v2/clan/:id
func GetClan(c *gin.Context) {
}

// UpdateClan Updates data about a clan
// PATCH /v2/clan/:id
func UpdateClan(c *gin.Context) {
}

// DeleteClan Deletes an individual clan
// DELETE /v2/clan/:id
func DeleteClan(c *gin.Context) {
}
