package main

import (
	"fmt"
	"github.com/Quaver/api2/handlers"
	"github.com/Quaver/api2/middleware"
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
	engine.POST("/v2/clan/invite", middleware.RequireAuth, handlers.CreateHandler(handlers.InviteUserToClan))
	engine.GET("/v2/clan/invite/:id", middleware.RequireAuth, handlers.CreateHandler(handlers.GetClanInvite))
	engine.GET("/v2/clan/invites", middleware.RequireAuth, handlers.CreateHandler(handlers.GetPendingClanInvites))
	engine.POST("/v2/clan/invite/:id/accept", middleware.RequireAuth, handlers.CreateHandler(handlers.AcceptClanInvite))
	engine.POST("/v2/clan/invite/:id/decline", middleware.RequireAuth, handlers.CreateHandler(handlers.DeclineClanInvite))

	// Clans
	engine.POST("/v2/clan", middleware.RequireAuth, handlers.CreateHandler(handlers.CreateClan))
	engine.GET("/v2/clan/:id", handlers.CreateHandler(handlers.GetClan))
	engine.PATCH("/v2/clan/:id", middleware.RequireAuth, handlers.CreateHandler(handlers.UpdateClan))
	engine.DELETE("/v2/clan/:id", middleware.RequireAuth, handlers.CreateHandler(handlers.DeleteClan))

	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	})
}
