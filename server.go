package main

import (
	"fmt"
	"github.com/Quaver/api2/handlers"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

// Starts the server on a given port
func initializeServer(port int) {
	gin.SetMode(gin.ReleaseMode)
	logrus.Info("Starting gin server in mode: ", gin.Mode())

	engine := gin.New()
	engine.Use(gin.Recovery())
	initializeRoutes(engine)

	logrus.Info(fmt.Sprintf("API is now being served on port :%v", port))
	logrus.Fatal(engine.Run(fmt.Sprintf(":%v", port)))
}

// Initializes all the routes for the server.
func initializeRoutes(engine *gin.Engine) {
	// Clan Invites
	engine.POST(createRoute("/clan/invite"), createHandler(handlers.InviteUserToClan))

	// Clans
	engine.POST(createRoute("/clan"), createHandler(handlers.CreateClan))
	engine.GET(createRoute("/clan/:id"), createHandler(handlers.GetClan))
	engine.PATCH(createRoute("/clan/:id"), createHandler(handlers.UpdateClan))
	engine.DELETE(createRoute("/clan/:id"), createHandler(handlers.DeleteClan))

	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	})

	logrus.Info("Initialized router")
}

// Creates a route with /v2. Example usage: createRoute("/foo/:id)
func createRoute(route string) string {
	return fmt.Sprintf("/v2%v", route)
}

// Creates a handler with automatic error handling
func createHandler(fn func(*gin.Context) *handlers.APIError) func(*gin.Context) {
	return func(c *gin.Context) {
		err := fn(c)

		if err == nil {
			return
		}

		if err.Error != nil {
			logrus.Errorf("%v - %v", err.Message, err.Error)
		}

		if err.Status == http.StatusInternalServerError {
			c.JSON(err.Status, gin.H{"error": "Internal Server Error"})
			return
		}

		c.JSON(err.Status, gin.H{"error": err.Message})
	}
}
