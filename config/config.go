package config

import (
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"os"
)

type Config struct {
	IsProduction bool `json:"is_production"`

	APIUrl string `json:"api_url"`

	JWTSecret string `json:"jwt_secret"`

	Server struct {
		Port                 int      `json:"port"`
		RateLimitIpWhitelist []string `json:"rate_limit_ip_whitelist"`
	} `json:"server"`

	Steam struct {
		AppId                   int    `json:"app_id"`
		PublisherKey            string `json:"publisher_key"`
		DonateRedirectUrl       string `json:"donate_redirect_url"`
		StorePaymentRedirectUrl string `json:"store_payment_redirect_url"`
	} `json:"steam"`

	SQL struct {
		Host     string `json:"host"`
		Username string `json:"username"`
		Password string `json:"password"`
		Database string `json:"database"`
	} `json:"sql"`

	Redis struct {
		Host     string `json:"host"`
		Password string `json:"password"`
		Database int    `json:"database"`
	} `json:"redis"`

	ElasticSearch struct {
		Host string `json:"host"`
	} `json:"elasticsearch"`

	Azure struct {
		AccountName string `json:"account_name"`
		AccountKey  string `json:"account_key"`
	} `json:"azure"`

	Cache struct {
		DataDirectory string `json:"data_directory"`
	}

	QuaverToolsPath string `json:"quaver_tools_path"`

	RankingQueue struct {
		Webhook               string `json:"webhook"`
		RankedWebhook         string `json:"ranked_webhook"`
		VotesRequired         int    `json:"votes_required"`
		DenialsRequired       int    `json:"denials_required"`
		MapsetUploadsRequired int    `json:"mapset_uploads_required"`
		ResubmissionDays      int    `json:"resubmission_days"`
	} `json:"ranking_queue"`

	BundledMapsets []int `json:"bundled_mapsets"`

	OrdersWebhook string `json:"orders_webhook"`

	Stripe struct {
		APIKey                  string `json:"api_key"`
		WebhookSigningSecret    string `json:"webhook_signing_secret"`
		DonateRedirectUrl       string `json:"donate_redirect_url"`
		StorePaymentRedirectUrl string `json:"store_payment_redirect_url"`
	} `json:"stripe"`

	Discord struct {
		BotAPI string `json:"bot_api"`
	} `json:"discord"`

	OpenAIAPIKey string `json:"openai_api_key"`

	CacheServer struct {
		URL string `json:"url"`
		Key string `json:"key"`
	} `json:"cache_server"`

	Cron struct {
		DonatorCheck         CronJob `json:"donator_check"`
		ElasticIndexMapsets  CronJob `json:"elastic_index_mapsets"`
		WeeklyMostPlayed     CronJob `json:"weekly_most_played"`
		UserRank             CronJob `json:"user_rank"`
		CacheLeaderboard     CronJob `json:"cache_leaderboard"`
		MigratePlaylists     CronJob `json:"migrate_playlists"`
		DatabaseBackup       CronJob `json:"database_backup"`
		DatabaseBackupHourly CronJob `json:"database_backup_hourly"`
	} `json:"cron"`
}

type CronJob struct {
	Job
}

type Job struct {
	Enabled  bool   `json:"enabled"`
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
}

var Instance *Config = nil

func Load(path string) error {
	if Instance != nil {
		return errors.New("config already loaded")
	}

	data, err := os.ReadFile(path)

	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &Instance)

	if err != nil {
		return err
	}

	if Instance.RankingQueue.VotesRequired < 1 || Instance.RankingQueue.DenialsRequired < 1 ||
		Instance.RankingQueue.ResubmissionDays < 1 {
		panic("ranking_queue configuration must be set and greater than 1")
	}

	logrus.Info("Config file has been loaded")
	return nil
}
