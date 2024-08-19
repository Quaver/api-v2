package main

import (
	"context"
	"fmt"
	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/handlers"
	"github.com/Quaver/api2/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"
)

// Starts the server on a given port
func initializeServer(port int) {
	gin.SetMode(gin.ReleaseMode)
	logrus.Info("Starting gin server in mode: ", gin.Mode())

	engine := gin.New()
	engine.Use(cors.Default())
	engine.Use(gin.Recovery())

	initializeRateLimiter(engine)
	initializeRoutes(engine)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: engine.Handler(),
	}

	go func() {
		logrus.Info(fmt.Sprintf("API is now being served on port :%v", port))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Listen: %s\n", err)
		}
	}()

	handleGracefulShutdown(server)
}

// Initializes the rate limiter for the server
func initializeRateLimiter(engine *gin.Engine) {
	store := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  time.Minute,
		Limit: 100,
	})

	engine.Use(ratelimit.RateLimiter(store, &ratelimit.Options{
		ErrorHandler: func(c *gin.Context, info ratelimit.Info) {
			isWhitelisted := slices.Contains(config.Instance.Server.RateLimitIpWhitelist, c.ClientIP())

			if !config.Instance.IsProduction || isWhitelisted {
				c.Next()
				return
			}

			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
		},
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
	}))
}

// Initializes all the routes for the server.
func initializeRoutes(engine *gin.Engine) {
	// Clan Invites
	engine.POST("/v2/clan/invite", middleware.RequireAuth, handlers.CreateHandler(handlers.InviteUserToClan))
	engine.GET("/v2/clan/invite/:id", middleware.RequireAuth, handlers.CreateHandler(handlers.GetClanInvite))
	engine.GET("/v2/clan/invites", middleware.RequireAuth, handlers.CreateHandler(handlers.GetClanPendingInvites))
	engine.GET("/v2/clan/user/invites", middleware.RequireAuth, handlers.CreateHandler(handlers.GetUserPendingClanInvites))
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

	// Clan Scores
	engine.GET("/v2/clan/:id/scores/:mode", handlers.CreateHandler(handlers.GetClanScoresForMode))
	engine.GET("/v2/clan/scores/:id", handlers.CreateHandler(handlers.GetUserScoresForClanScore))

	// Users
	engine.POST("/v2/user", handlers.CreateHandler(handlers.RegisterNewUser))
	engine.GET("/v2/user/:id", middleware.AllowAuth, handlers.CreateHandler(handlers.GetUser))
	engine.GET("/v2/user/:id/aboutme", middleware.AllowAuth, handlers.CreateHandler(handlers.GetUserAboutMe))
	engine.GET("/v2/user/:id/achievements", middleware.AllowAuth, handlers.CreateHandler(handlers.GetUserAchievements))
	engine.GET("/v2/user/:id/activity", middleware.AllowAuth, handlers.CreateHandler(handlers.GetUserActivity))
	engine.GET("/v2/user/:id/badges", middleware.AllowAuth, handlers.CreateHandler(handlers.GetUserBadges))
	engine.GET("/v2/user/:id/mapsets", middleware.AllowAuth, handlers.CreateHandler(handlers.GetUserMapsets))
	engine.GET("/v2/user/:id/playlists", middleware.AllowAuth, handlers.CreateHandler(handlers.GetUserPlaylists))
	engine.GET("/v2/user/:id/mostplayed", middleware.AllowAuth, handlers.CreateHandler(handlers.GetUserMostPlayedMaps))
	engine.GET("/v2/user/:id/scores/:mode/best", middleware.AllowAuth, handlers.CreateHandler(handlers.GetUserBestScoresForMode))
	engine.GET("/v2/user/:id/scores/:mode/recent", middleware.AllowAuth, handlers.CreateHandler(handlers.GetUserRecentScoresForMode))
	engine.GET("/v2/user/:id/scores/:mode/firstplace", middleware.AllowAuth, handlers.CreateHandler(handlers.GetUserFirstPlaceScoresForMode))
	engine.GET("/v2/user/:id/scores/:mode/grades/:grade", middleware.AllowAuth, handlers.CreateHandler(handlers.GetUserGradesForMode))
	engine.GET("/v2/user/:id/scores/:mode/pinned", middleware.AllowAuth, handlers.CreateHandler(handlers.GetPinnedScoresForMode))
	engine.GET("/v2/user/:id/statistics/:mode/rank", middleware.AllowAuth, handlers.CreateHandler(handlers.GetUserRankStatisticsForMode))
	engine.POST("/v2/user/:id/ban", middleware.RequireAuth, handlers.CreateHandler(handlers.BanUser))
	engine.POST("/v2/user/:id/unban", middleware.RequireAuth, handlers.CreateHandler(handlers.UnbanUser))
	engine.POST("/v2/user/:id/discord", middleware.RequireAuth, handlers.CreateHandler(handlers.UpdateUserDiscordId))
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
	engine.GET("/v2/mapset/search", handlers.CreateHandler(handlers.GetMapsetsSearch))
	engine.POST("/v2/mapset", middleware.RequireAuth, handlers.CreateHandler(handlers.HandleMapsetSubmission))
	engine.GET("/v2/mapset/:id", handlers.CreateHandler(handlers.GetMapsetById))
	engine.POST("/v2/mapset/:id/delete", middleware.RequireAuth, handlers.CreateHandler(handlers.DeleteMapset))
	engine.GET("/v2/mapset/ranked", handlers.CreateHandler(handlers.GetRankedMapsetIds))
	engine.GET("/v2/mapset/offsets", handlers.CreateHandler(handlers.GetMapsetOnlineOffsets))
	engine.GET("/v2/mapset/:id/elastic", middleware.RequireAuth, handlers.CreateHandler(handlers.UpdateElasticSearchMapset))
	engine.POST("/v2/mapset/:id/description", middleware.RequireAuth, handlers.CreateHandler(handlers.UpdateMapsetDescription))
	engine.POST("/v2/mapset/:id/explicit", middleware.RequireAuth, handlers.CreateHandler(handlers.MarkMapsetAsExplicit))
	engine.POST("/v2/mapset/:id/unexplicit", middleware.RequireAuth, handlers.CreateHandler(handlers.MarkMapsetAsNotExplicit))

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
	engine.POST("/v2/download/multiplayer/:id/upload", middleware.RequireAuth, handlers.CreateHandler(handlers.UploadMultiplayerMapset))
	engine.GET("/v2/download/multiplayer/:id", middleware.RequireAuth, handlers.CreateHandler(handlers.DownloadMultiplayerMapset))

	// Logs
	engine.POST("/v2/logs/crash", middleware.RequireAuth, handlers.CreateHandler(handlers.AddCrashLog))

	// Leaderboards
	engine.GET("/v2/leaderboard/global", handlers.CreateHandler(handlers.GetGlobalLeaderboardForMode))
	engine.GET("/v2/leaderboard/country", handlers.CreateHandler(handlers.GetCountryLeaderboard))
	engine.GET("/v2/leaderboard/hits", handlers.CreateHandler(handlers.GetTotalHitsLeaderboard))
	engine.GET("/v2/leaderboard/clans", handlers.CreateHandler(handlers.GetClanLeaderboard))

	// Scores
	engine.GET("/v2/scores/:md5/stats", middleware.RequireAuth, handlers.CreateHandler(handlers.GetVirtualReplayPlayerOutput))

	engine.GET("/v2/scores/:md5/global", middleware.AllowAuth, handlers.CreateHandler(handlers.GetGlobalScoresForMap))
	engine.GET("/v2/scores/:md5/country/:country", middleware.RequireAuth, handlers.CreateHandler(handlers.GetCountryScoresForMap))
	engine.GET("/v2/scores/:md5/mods/:mods", middleware.AllowAuth, handlers.CreateHandler(handlers.GetModifierScoresForMap))
	engine.GET("/v2/scores/:md5/rate/:mods", middleware.AllowAuth, handlers.CreateHandler(handlers.GetRateScoresForMap))
	engine.GET("/v2/scores/:md5/all", middleware.RequireAuth, handlers.CreateHandler(handlers.GetAllScoresForMap))
	engine.GET("/v2/scores/:md5/friends", middleware.RequireAuth, handlers.CreateHandler(handlers.GetFriendScoresForMap))
	engine.GET("/v2/scores/:md5/clans", middleware.AllowAuth, handlers.CreateHandler(handlers.GetClanScoresForMap))
	// Scores (Personal Best)
	engine.GET("/v2/scores/:md5/:user_id/global", middleware.AllowAuth, handlers.CreateHandler(handlers.GetUserPersonalBestScoreGlobal))
	engine.GET("/v2/scores/:md5/:user_id/all", middleware.AllowAuth, handlers.CreateHandler(handlers.GetUserPersonalBestScoreAll))
	engine.GET("/v2/scores/:md5/:user_id/mods/:mods", middleware.AllowAuth, handlers.CreateHandler(handlers.GetUserPersonalBestScoreMods))
	engine.GET("/v2/scores/:md5/:user_id/rate/:mods", middleware.AllowAuth, handlers.CreateHandler(handlers.GetUserPersonalBestScoreRate))
	engine.GET("/v2/scores/:md5/:user_id/clan", middleware.AllowAuth, handlers.CreateHandler(handlers.GetClanPersonalBestScore)) // user_id is clan_id

	// Pinned Scores
	engine.POST("/v2/scores/:id/pin", middleware.RequireAuth, handlers.CreateHandler(handlers.CreatePinnedScore))
	engine.POST("/v2/scores/:id/unpin", middleware.RequireAuth, handlers.CreateHandler(handlers.RemovePinnedScore))
	engine.POST("/v2/scores/pinned/:mode/sort", middleware.RequireAuth, handlers.CreateHandler(handlers.SortPinnedScores))

	// Ranking Queue
	engine.GET("/v2/ranking/config", handlers.CreateHandler(handlers.GetRankingQueueConfig))
	engine.GET("/v2/ranking/queue/mode/:mode", handlers.CreateHandler(handlers.GetRankingQueue))
	engine.GET("/v2/ranking/queue/supervisors/actions", middleware.RequireAuth, handlers.CreateHandler(handlers.GetRankingSupervisorActions))
	engine.GET("/v2/ranking/queue/:id", handlers.CreateHandler(handlers.GetRankingQueueMapset))
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
	engine.GET("/v2/orders/donations/prices", handlers.CreateHandler(handlers.GetDonatorPrices))
	engine.POST("/v2/orders/checkout", middleware.RequireAuth, handlers.CreateHandler(handlers.CreateOrderCheckoutSession))

	// Orders Steam
	engine.POST("/v2/orders/steam/initiate/donation", middleware.RequireAuth, handlers.CreateHandler(handlers.InitiateSteamDonatorTransaction))
	engine.GET("/v2/orders/steam/finalize", handlers.CreateHandler(handlers.FinalizeSteamTransaction))

	// Orders Stripe
	engine.GET("/v2/orders/stripe/subscriptions", middleware.RequireAuth, handlers.CreateHandler(handlers.GetActiveSubscriptions))
	engine.GET("/v2/orders/stripe/subscriptions/modify", middleware.RequireAuth, handlers.CreateHandler(handlers.ModifyStripeSubscription))
	engine.POST("/v2/orders/stripe/initiate/donation", middleware.RequireAuth, handlers.CreateHandler(handlers.InitiateStripeDonatorCheckoutSession))
	engine.POST("/v2/orders/stripe/webhook", handlers.CreateHandler(handlers.HandleStripeWebhook))

	// Applications
	engine.GET("/v2/developers/applications", middleware.RequireAuth, handlers.CreateHandler(handlers.GetUserApplications))
	engine.POST("/v2/developers/applications", middleware.RequireAuth, handlers.CreateHandler(handlers.CreateNewApplication))
	engine.GET("/v2/developers/applications/:id", middleware.RequireAuth, handlers.CreateHandler(handlers.GetUserApplication))
	engine.POST("/v2/developers/applications/:id", middleware.RequireAuth, handlers.CreateHandler(handlers.UpdateApplication))
	engine.DELETE("/v2/developers/applications/:id", middleware.RequireAuth, handlers.CreateHandler(handlers.DeleteUserApplication))
	engine.POST("/v2/developers/applications/:id/secret", middleware.RequireAuth, handlers.CreateHandler(handlers.ResetApplicationSecret))

	// Notifications
	engine.GET("/v2/notifications", middleware.RequireAuth, handlers.CreateHandler(handlers.GetUserNotifications))
	engine.POST("/v2/notifications", middleware.RequireAuth, handlers.CreateHandler(handlers.CreateUserNotification))
	engine.POST("/v2/notifications/:id/read", middleware.RequireAuth, handlers.CreateHandler(handlers.MarkUserNotificationAsRead))
	engine.POST("/v2/notifications/:id/unread", middleware.RequireAuth, handlers.CreateHandler(handlers.MarkUserNotificationAsUnread))
	engine.DELETE("/v2/notifications/:id", middleware.RequireAuth, handlers.CreateHandler(handlers.DeleteNotification))

	// Artists
	engine.POST("/v2/artists", middleware.RequireAuth, handlers.CreateHandler(handlers.InsertMusicArtist))
	engine.POST("/v2/artists/:id", middleware.RequireAuth, handlers.CreateHandler(handlers.UpdateMusicArtist))
	engine.DELETE("/v2/artists/:id", middleware.RequireAuth, handlers.CreateHandler(handlers.DeleteMusicArtist))
	engine.GET("/v2/artists", handlers.CreateHandler(handlers.GetMusicArtists))
	engine.GET("/v2/artists/:id", handlers.CreateHandler(handlers.GetSingleMusicArtist))
	engine.POST("/v2/artists/:id/avatar", middleware.RequireAuth, handlers.CreateHandler(handlers.UploadMusicArtistAvatar))
	engine.POST("/v2/artists/:id/banner", middleware.RequireAuth, handlers.CreateHandler(handlers.UploadMusicArtistBanner))
	engine.POST("/v2/artists/sort", middleware.RequireAuth, handlers.CreateHandler(handlers.SortMusicArtists))

	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	})
}

func handleGracefulShutdown(server *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	timeout := 5 * time.Second

	if !config.Instance.IsProduction {
		timeout = time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logrus.Fatal("Server Shutdown: ", err)
	}

	select {
	case <-ctx.Done():
		logrus.Infof("Server Shutdown")
	}
}
