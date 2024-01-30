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

	// Clan Members
	engine.GET("/v2/clan/:id/members", handlers.CreateHandler(handlers.GetClanMembers))
	engine.POST("/v2/clan/leave", middleware.RequireAuth, handlers.CreateHandler(handlers.LeaveClan))
	engine.POST("/v2/clan/transfer/:user_id", middleware.RequireAuth, handlers.CreateHandler(handlers.TransferClanOwnership))
	engine.POST("/v2/clan/kick/:user_id", middleware.RequireAuth, handlers.CreateHandler(handlers.KickClanMember))

	// Clan Activity
	engine.GET("/v2/clan/:id/activity", handlers.CreateHandler(handlers.GetClanActivity))

	// Clan Images
	engine.POST("/v2/clan/avatar", middleware.RequireAuth, handlers.CreateHandler(handlers.UploadClanAvatar))
	engine.POST("/v2/clan/banner", middleware.RequireAuth, handlers.CreateHandler(handlers.UploadClanBanner))

	// Users
	engine.GET("/v2/user/:id", handlers.CreateHandler(handlers.GetUser))
	engine.GET("/v2/user/:id/achievements", handlers.CreateHandler(handlers.GetUserAchievements))
	engine.GET("/v2/user/:id/activity", handlers.CreateHandler(handlers.GetUserActivity))
	engine.GET("/v2/user/:id/badges", handlers.CreateHandler(handlers.GetUserBadges))
	engine.GET("/v2/user/search/:name", handlers.CreateHandler(handlers.SearchUsers))

	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	})
}
