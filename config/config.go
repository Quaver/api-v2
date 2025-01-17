package config

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/sirupsen/logrus"
)

type Config struct {
	IsProduction bool `json:"is_production"`

	APIUrl string `json:"api_url"`

	WebsiteUrl string `json:"website_url"`

	JWTSecret string `json:"jwt_secret"`

	Server struct {
		Port                 int      `json:"port"`
		RateLimitIpWhitelist []string `json:"rate_limit_ip_whitelist"`
	} `json:"server"`

	ApiV1 struct {
		Url       string `json:"url"`
		SecretKey string `json:"secret_key"`
	} `json:"api_v1"`

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
		Webhook                         string `json:"webhook"`
		RankedWebhook                   string `json:"ranked_webhook"`
		VotesRequired                   int    `json:"votes_required"`
		DenialsRequired                 int    `json:"denials_required"`
		MapsetUploadsRequired           int    `json:"mapset_uploads_required"`
		ResubmissionDays                int    `json:"resubmission_days"`
		WeeklyRequiredSupervisorActions int    `json:"weekly_required_supervisor_actions"`
	} `json:"ranking_queue"`

	BundledMapsets []int `json:"bundled_mapsets"`

	EventsWebhook          string `json:"events_webhook"`
	TeamAnnounceWebhook    string `json:"team_announce_webhook"`
	ClansFirstPlaceWebhook string `json:"clans_first_place_webhook"`
	ClansMapRankedWebhook  string `json:"clans_map_ranked_webhook"`
	CrashLogWebhook        string `json:"crash_log_webhook"`

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

	S3 struct {
		Endpoint  string `json:"endpoint"`
		Region    string `json:"region"`
		AccessKey string `json:"access_key"`
		Secret    string `json:"secret"`
		Bucket    string `json:"bucket"`
	} `json:"s3"`

	Cron struct {
		DonatorCheck         CronJob `json:"donator_check"`
		ElasticIndexMapsets  CronJob `json:"elastic_index_mapsets"`
		WeeklyMostPlayed     CronJob `json:"weekly_most_played"`
		UserRank             CronJob `json:"user_rank"`
		CacheLeaderboard     CronJob `json:"cache_leaderboard"`
		MigratePlaylists     CronJob `json:"migrate_playlists"`
		DatabaseBackup       CronJob `json:"database_backup"`
		DatabaseBackupHourly CronJob `json:"database_backup_hourly"`
		SupervisorActivity   CronJob `json:"supervisor_activity"`
		RankClanMap          CronJob `json:"rank_clan_map"`
		DenyOnHoldOneMonth   CronJob `json:"deny_on_hold_one_month"`
		ClanRecalculate      CronJob `json:"clan_recalculate"`
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
