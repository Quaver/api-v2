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
	engine.POST("/v2/clan/:id", middleware.RequireAuth, handlers.CreateHandler(handlers.UpdateClan))
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
	engine.POST("/v2/user", handlers.CreateHandler(handlers.RegisterNewUser))
	engine.GET("/v2/user/:id", handlers.CreateHandler(handlers.GetUser))
	engine.GET("/v2/user/:id/achievements", handlers.CreateHandler(handlers.GetUserAchievements))
	engine.GET("/v2/user/:id/activity", handlers.CreateHandler(handlers.GetUserActivity))
	engine.GET("/v2/user/:id/badges", handlers.CreateHandler(handlers.GetUserBadges))
	engine.GET("/v2/user/:id/mapsets", handlers.CreateHandler(handlers.GetUserMapsets))
	engine.GET("/v2/user/:id/playlists", handlers.CreateHandler(handlers.GetUserPlaylists))
	engine.GET("/v2/user/:id/mostplayed", handlers.CreateHandler(handlers.GetUserMostPlayedMaps))
	engine.GET("/v2/user/:id/scores/:mode/best", handlers.CreateHandler(handlers.GetUserBestScoresForMode))
	engine.GET("/v2/user/:id/scores/:mode/recent", middleware.AllowAuth, handlers.CreateHandler(handlers.GetUserRecentScoresForMode))
	engine.GET("/v2/user/:id/scores/:mode/firstplace", handlers.CreateHandler(handlers.GetUserFirstPlaceScoresForMode))
	engine.GET("/v2/user/:id/scores/:mode/grades/:grade", handlers.CreateHandler(handlers.GetUserGradesForMode))
	engine.GET("/v2/user/:id/statistics/:mode/rank", handlers.CreateHandler(handlers.GetUserRankStatisticsForMode))
	engine.POST("/v2/user/:id/ban", middleware.RequireAuth, handlers.CreateHandler(handlers.BanUser))
	engine.POST("/v2/user/:id/unban", middleware.RequireAuth, handlers.CreateHandler(handlers.UnbanUser))
	engine.GET("/v2/user/search/:name", handlers.CreateHandler(handlers.SearchUsers))
	engine.GET("/v2/user/team/members", handlers.CreateHandler(handlers.GetTeamMembers))

	// User Profile
	engine.POST("/v2/user/profile/aboutme", middleware.RequireAuth, handlers.CreateHandler(handlers.UpdateUserAboutMe))
	engine.POST("/v2/user/profile/cover", middleware.RequireAuth, handlers.CreateHandler(handlers.UploadUserProfileCover))
	engine.GET("/v2/user/profile/username/eligible", middleware.RequireAuth, handlers.CreateHandler(handlers.GetCanUserChangeUsername))
	engine.GET("/v2/user/profile/username/available", middleware.RequireAuth, handlers.CreateHandler(handlers.IsUsernameAvailable))
	engine.POST("/v2/user/profile/username/", middleware.RequireAuth, handlers.CreateHandler(handlers.ChangeUserUsername))

	// User Relationships
	engine.GET("/v2/user/relationship/friends", middleware.RequireAuth, handlers.CreateHandler(handlers.GetFriendsList))
	engine.POST("/v2/user/:id/relationship/add", middleware.RequireAuth, handlers.CreateHandler(handlers.AddFriend))
	engine.POST("/v2/user/:id/relationship/remove", middleware.RequireAuth, handlers.CreateHandler(handlers.RemoveFriend))

	// Maps
	engine.GET("/v2/map/:id", handlers.CreateHandler(handlers.GetMap))
	engine.POST("/v2/map", middleware.RequireAuth, handlers.CreateHandler(handlers.UploadUnsubmittedMap))

	// Map Mods
	engine.GET("/v2/map/:id/mods", handlers.CreateHandler(handlers.GetMapMods))
	engine.POST("/v2/map/:id/mods", middleware.RequireAuth, handlers.CreateHandler(handlers.SubmitMapMod))
	engine.POST("/v2/map/:id/mods/:mod_id/status", middleware.RequireAuth, handlers.CreateHandler(handlers.UpdateMapModStatus))
	engine.POST("/v2/map/:id/mods/:mod_id/comment", middleware.RequireAuth, handlers.CreateHandler(handlers.SubmitMapModComment))

	// Mapsets
	engine.GET("/v2/mapset/:id", handlers.CreateHandler(handlers.GetMapsetById))
	engine.POST("/v2/mapset/:id/delete", middleware.RequireAuth, handlers.CreateHandler(handlers.DeleteMapset))
	engine.GET("/v2/mapset/ranked", handlers.CreateHandler(handlers.GetRankedMapsetIds))
	engine.GET("/v2/mapset/offsets", handlers.CreateHandler(handlers.GetMapsetOnlineOffsets))
	engine.POST("/v2/mapset/:id/description", middleware.RequireAuth, handlers.CreateHandler(handlers.UpdateMapsetDescription))

	// Chat
	engine.GET("/v2/chat/:channel/history", middleware.RequireAuth, handlers.CreateHandler(handlers.GetChatHistory))

	// Server
	engine.GET("/v2/server/stats", handlers.CreateHandler(handlers.GetServerStats))
	engine.GET("/v2/server/stats/country", handlers.CreateHandler(handlers.GetCountryPlayers))
	engine.GET("/v2/server/stats/mostplayed", handlers.CreateHandler(handlers.GetWeeklyMostPlayedMapsets))

	// Download
	engine.GET("/v2/download/map/:id", handlers.CreateHandler(handlers.DownloadQua))
	engine.GET("/v2/download/replay/:id", handlers.CreateHandler(handlers.DownloadReplay))
	engine.GET("/v2/download/mapset/:id", middleware.RequireAuth, handlers.CreateHandler(handlers.DownloadMapset))
	// engine.POST("/v2/download/multiplayer/:id/upload", middleware.RequireAuth, handlers.CreateHandler(handlers.UploadMultiplayerMapset))
	engine.GET("/v2/download/multiplayer/:id", middleware.RequireAuth, handlers.CreateHandler(handlers.DownloadMultiplayerMapset))

	// Logs
	engine.POST("/v2/logs/crash", middleware.RequireAuth, handlers.CreateHandler(handlers.AddCrashLog))

	// Leaderboards
	engine.GET("/v2/leaderboard/global", handlers.CreateHandler(handlers.GetGlobalLeaderboardForMode))
	engine.GET("/v2/leaderboard/country", handlers.CreateHandler(handlers.GetCountryLeaderboard))
	engine.GET("/v2/leaderboard/hits", handlers.CreateHandler(handlers.GetTotalHitsLeaderboard))

	// Scores
	engine.GET("/v2/scores/:md5/global", middleware.AllowAuth, handlers.CreateHandler(handlers.GetGlobalScoresForMap))
	engine.GET("/v2/scores/:md5/country/:country", middleware.RequireAuth, handlers.CreateHandler(handlers.GetCountryScoresForMap))
	engine.GET("/v2/scores/:md5/mods/:mods", middleware.AllowAuth, handlers.CreateHandler(handlers.GetModifierScoresForMap))
	engine.GET("/v2/scores/:md5/rate/:mods", middleware.AllowAuth, handlers.CreateHandler(handlers.GetRateScoresForMap))
	engine.GET("/v2/scores/:md5/all", middleware.RequireAuth, handlers.CreateHandler(handlers.GetAllScoresForMap))
	engine.GET("/v2/scores/:md5/friends", middleware.RequireAuth, handlers.CreateHandler(handlers.GetFriendScoresForMap))
	// Scores (Personal Best)
	engine.GET("/v2/scores/:md5/:user_id/global", middleware.AllowAuth, handlers.CreateHandler(handlers.GetUserPersonalBestScoreGlobal))
	engine.GET("/v2/scores/:md5/:user_id/all", middleware.AllowAuth, handlers.CreateHandler(handlers.GetUserPersonalBestScoreAll))
	engine.GET("/v2/scores/:md5/:user_id/mods/:mods", middleware.AllowAuth, handlers.CreateHandler(handlers.GetUserPersonalBestScoreMods))
	engine.GET("/v2/scores/:md5/:user_id/rate/:mods", middleware.AllowAuth, handlers.CreateHandler(handlers.GetUserPersonalBestScoreRate))

	// Ranking Queue
	engine.GET("/v2/ranking/config", handlers.CreateHandler(handlers.GetRankingQueueConfig))
	engine.GET("/v2/ranking/queue/mode/:mode", handlers.CreateHandler(handlers.GetRankingQueue))
	engine.POST("/v2/ranking/queue/:id/submit", middleware.RequireAuth, handlers.CreateHandler(handlers.SubmitMapsetToRankingQueue))
	engine.POST("/v2/ranking/queue/:id/remove", middleware.RequireAuth, handlers.CreateHandler(handlers.RemoveFromRankingQueue))
	engine.GET("/v2/ranking/queue/:id/comments", handlers.CreateHandler(handlers.GetRankingQueueComments))
	engine.POST("/v2/ranking/queue/:id/comment", middleware.RequireAuth, handlers.CreateHandler(handlers.AddRankingQueueComment))
	engine.POST("/v2/ranking/queue/comment/:id/edit", middleware.RequireAuth, handlers.CreateHandler(handlers.EditRankingQueueComment))
	engine.POST("/v2/ranking/queue/:id/vote", middleware.RequireAuth, handlers.CreateHandler(handlers.VoteForRankingQueueMapset))
	engine.POST("/v2/ranking/queue/:id/deny", middleware.RequireAuth, handlers.CreateHandler(handlers.DenyRankingQueueMapset))
	engine.POST("/v2/ranking/queue/:id/blacklist", middleware.RequireAuth, handlers.CreateHandler(handlers.BlacklistRankingQueueMapset))
	engine.POST("/v2/ranking/queue/:id/hold", middleware.RequireAuth, handlers.CreateHandler(handlers.OnHoldRankingQueueMapset))

	// Game Builds
	engine.POST("/v2/builds", middleware.RequireAuth, handlers.CreateHandler(handlers.AddNewGameBuild))

	// Multiplayer
	engine.GET("/v2/multiplayer/games", handlers.CreateHandler(handlers.GetRecentMultiplayerGames))
	engine.GET("/v2/multiplayer/game/:id", handlers.CreateHandler(handlers.GetMultiplayerGame))

	// Playlists
	engine.POST("/v2/playlists", middleware.RequireAuth, handlers.CreateHandler(handlers.CreatePlaylist))
	engine.GET("/v2/playlists/search", handlers.CreateHandler(handlers.SearchPlaylists))
	engine.GET("/v2/playlists/:id", handlers.CreateHandler(handlers.GetPlaylist))
	engine.POST("/v2/playlists/:id/update", middleware.RequireAuth, handlers.CreateHandler(handlers.UpdatePlaylist))
	engine.DELETE("/v2/playlists/:id", middleware.RequireAuth, handlers.CreateHandler(handlers.DeletePlaylist))
	engine.GET("/v2/playlists/:id/contains/:map_id", handlers.CreateHandler(handlers.GetPlaylistContainsMap))
	engine.POST("/v2/playlists/:id/add/:map_id", middleware.RequireAuth, handlers.CreateHandler(handlers.AddMapToPlaylist))
	engine.POST("/v2/playlists/:id/remove/:map_id", middleware.RequireAuth, handlers.CreateHandler(handlers.RemoveMapFromPlaylist))
	engine.POST("/v2/playlists/:id/cover", middleware.RequireAuth, handlers.CreateHandler(handlers.UploadPlaylistCover))

	// Orders
	engine.GET("/v2/orders", middleware.RequireAuth, handlers.CreateHandler(handlers.GetUserOrders))
	engine.POST("/v2/orders/checkout", middleware.RequireAuth, handlers.CreateHandler(handlers.CreateOrderCheckoutSession))

	// Orders Steam
	engine.POST("/v2/orders/steam/initiate/donation", middleware.RequireAuth, handlers.CreateHandler(handlers.InitiateSteamDonatorTransaction))
	engine.GET("/v2/orders/steam/finalize", handlers.CreateHandler(handlers.FinalizeSteamTransaction))

	// Orders Stripe
	engine.GET("/v2/orders/stripe/subscriptions", middleware.RequireAuth, handlers.CreateHandler(handlers.GetActiveSubscriptions))
	engine.GET("/v2/orders/stripe/subscriptions/modify", middleware.RequireAuth, handlers.CreateHandler(handlers.ModifyStripeSubscription))
	engine.POST("/v2/orders/stripe/initiate/donation", middleware.RequireAuth, handlers.CreateHandler(handlers.InitiateStripeDonatorCheckoutSession))
	engine.POST("/v2/orders/stripe/webhook", handlers.CreateHandler(handlers.HandleStripeWebhook))

	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	})
}
