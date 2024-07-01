package db

import (
	"encoding/json"
	"fmt"
	"github.com/Quaver/api2/enums"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"time"
)

type Score struct {
	Id                          int              `gorm:"column:id; PRIMARY_KEY" json:"id"`
	UserId                      int              `gorm:"column:user_id" json:"user_id"`
	MapMD5                      string           `gorm:"column:map_md5" json:"map_md5"`
	ReplayMD5                   string           `gorm:"column:replay_md5" json:"replay_md5"`
	Timestamp                   int64            `gorm:"column:timestamp" json:"-"`
	TimestampJSON               time.Time        `gorm:"-:all" json:"timestamp"`
	IsPersonalBest              bool             `gorm:"column:personal_best" json:"is_personal_best"`
	PerformanceRating           float64          `gorm:"column:performance_rating" json:"performance_rating"`
	Modifiers                   int64            `gorm:"column:mods" json:"modifiers"`
	Failed                      bool             `gorm:"column:failed" json:"failed"`
	TotalScore                  int              `gorm:"column:total_score" json:"total_score"`
	Accuracy                    float64          `gorm:"column:accuracy" json:"accuracy"`
	MaxCombo                    int              `gorm:"column:max_combo" json:"max_combo"`
	CountMarvelous              int              `gorm:"column:count_marv" json:"count_marvelous"`
	CountPerfect                int              `gorm:"column:count_perf" json:"count_perfect"`
	CountGreat                  int              `gorm:"column:count_great" json:"count_great"`
	CountGood                   int              `gorm:"column:count_good" json:"count_good"`
	CountOkay                   int              `gorm:"column:count_okay" json:"count_okay"`
	CountMiss                   int              `gorm:"column:count_miss" json:"count_miss"`
	Grade                       string           `gorm:"column:grade" json:"grade"`
	ScrollSpeed                 int              `gorm:"column:scroll_speed" json:"scroll_speed"`
	TimePlayStart               int64            `gorm:"column:time_play_start" json:"-"`
	TimePlayEnd                 int64            `gorm:"column:time_play_end" json:"-"`
	IP                          string           `gorm:"column:ip" json:"-"`
	ExecutingAssembly           string           `gorm:"column:executing_assembly" json:"-"`
	EntryAssembly               string           `gorm:"column:entry_assembly" json:"-"`
	QuaverVersion               string           `gorm:"column:quaver_version" json:"-"`
	PauseCount                  int              `gorm:"column:pause_count" json:"-"`
	PerformanceProcessorVersion string           `gorm:"column:performance_processor_version" json:"-"`
	DifficultyProcessorVersion  string           `gorm:"column:difficulty_processor_version" json:"-"`
	IsDonatorScore              bool             `gorm:"column:is_donator_score" json:"is_donator_score"`
	TournamentGameId            *int             `gorm:"column:tournament_game_id" json:"tournament_game_id"`
	ClanId                      *int             `gorm:"column:clan_id" json:"clan_id"`
	Map                         *MapQua          `gorm:"foreignKey:MapMD5; references:MD5" json:"map,omitempty"`
	User                        *User            `gorm:"foreignKey:UserId; references:Id" json:"user,omitempty"`
	FirstPlace                  *ScoreFirstPlace `gorm:"foreignKey:Id; references:ScoreId" json:"-"`
}

func (s *Score) TableName() string {
	return "scores"
}

func (s *Score) BeforeCreate(*gorm.DB) (err error) {
	s.TimestampJSON = time.Now()
	return nil
}

func (s *Score) AfterFind(*gorm.DB) (err error) {
	s.TimestampJSON = time.UnixMilli(s.Timestamp)
	return nil
}

// GetUserBestScoresForMode Retrieves a user's best scores for a given game mode
func GetUserBestScoresForMode(id int, mode enums.GameMode, limit int, page int) ([]*Score, error) {
	var scores []*Score

	result := SQL.
		InnerJoins("Map").
		Where("scores.personal_best = 1 AND "+
			"scores.user_id = ? AND "+
			"scores.mode = ? AND "+
			"scores.is_donator_score = 0", id, mode).
		Order("scores.performance_rating DESC").
		Limit(limit).
		Offset(page * limit).
		Find(&scores)

	if result.Error != nil {
		return nil, result.Error
	}

	return scores, nil
}

// GetUserRecentScoresForMode Retrieves a user's recent scores for a given mode
func GetUserRecentScoresForMode(id int, mode enums.GameMode, isDonator bool, limit int, page int) ([]*Score, error) {
	var scores []*Score

	donatorScore := " AND scores.is_donator_score = 0"

	if isDonator {
		donatorScore = ""
	}

	result := SQL.
		InnerJoins("Map").
		Where("scores.user_id = ? AND "+
			"scores.mode = ?"+
			donatorScore, id, mode).
		Order("scores.timestamp DESC").
		Limit(limit).
		Offset(page * limit).
		Find(&scores)

	if result.Error != nil {
		return nil, result.Error
	}

	return scores, nil
}

// GetUserFirstPlaceScoresForMode Retrieves a user's first place scores for a given mode
func GetUserFirstPlaceScoresForMode(id int, mode enums.GameMode, limit int, page int) ([]*Score, error) {
	var scores []*Score

	result := SQL.
		Joins("FirstPlace").
		Joins("Map").
		Where("FirstPlace.user_id = ? AND Map.game_mode = ?", id, mode).
		Order("FirstPlace.performance_rating DESC").
		Limit(limit).
		Offset(page * limit).
		Find(&scores)

	if result.Error != nil {
		return nil, result.Error
	}

	return scores, nil
}

// GetUserGradeScoresForMode Retrieves a user's scores for a particular grade
func GetUserGradeScoresForMode(id int, mode enums.GameMode, grade string, limit int, page int) ([]*Score, error) {
	var scores []*Score

	result := SQL.
		Joins("Map").
		Where("scores.user_id = ? "+
			"AND Map.game_mode = ? "+
			"AND scores.grade = ? "+
			"AND scores.personal_best = 1 "+
			"AND scores.is_donator_score = 0", id, mode, grade).
		Order("scores.performance_rating DESC").
		Limit(limit).
		Offset(page * limit).
		Find(&scores)

	if result.Error != nil {
		return nil, result.Error
	}

	return scores, nil
}

// GetScoreById Retrieves a score from the database by its id
func GetScoreById(id int) (*Score, error) {
	var score *Score

	result := SQL.
		Joins("Map").
		Where("scores.id = ?", id).
		First(&score)

	if result.Error != nil {
		return nil, result.Error
	}

	return score, nil
}

// GetGlobalScoresForMap Retrieves the global scores for a map
func GetGlobalScoresForMap(md5 string, limit int, page int, useCache bool) ([]*Score, error) {
	if useCache {
		cached, err := getCachedScoreboard(scoreboardGlobal, md5)

		if err != nil {
			return nil, err
		}

		if cached != nil {
			return cached, nil
		}
	}

	var scores []*Score

	result := SQL.
		Joins("User").
		Where("scores.map_md5 = ? "+
			"AND scores.personal_best = 1 "+
			"AND User.allowed = 1", md5).
		Order("scores.performance_rating DESC").
		Limit(limit).
		Offset(page * limit).
		Find(&scores)

	if result.Error != nil {
		return nil, result.Error
	}

	if useCache {
		if err := cacheScoreboard(scoreboardGlobal, md5, scores); err != nil {
			return nil, err
		}
	}

	return scores, nil
}

// GetCountryScoresForMap Retrieves the country scores for a map
func GetCountryScoresForMap(md5 string, country string, limit int, page int) ([]*Score, error) {
	cached, err := getCachedScoreboard(scoreboardCountry, md5)

	if err != nil {
		return nil, err
	}

	if cached != nil {
		return cached, nil
	}

	var scores []*Score

	result := SQL.
		Joins("User").
		Where("scores.map_md5 = ? "+
			"AND scores.personal_best = 1 "+
			"AND User.country = ? "+
			"AND User.allowed = 1", md5, country).
		Order("scores.performance_rating DESC").
		Limit(limit).
		Offset(page * limit).
		Find(&scores)

	if result.Error != nil {
		return nil, result.Error
	}

	if err := cacheScoreboard(scoreboardCountry, md5, scores); err != nil {
		return nil, err
	}

	return scores, nil
}

// GetModifierScoresForMap Retrieves the modifier scores for a map
func GetModifierScoresForMap(md5 string, mods int64, limit int, page int) ([]*Score, error) {
	cached, err := getCachedScoreboard(scoreboardMods, md5)

	if err != nil {
		return nil, err
	}

	if cached != nil {
		return cached, nil
	}

	var scores []*Score

	result := SQL.
		Joins("User").
		Where("scores.map_md5 = ? "+
			"AND scores.failed = 0 "+
			"AND (scores.mods & ?) != 0 "+
			"AND User.allowed = 1", md5, mods).
		Group("User.id").
		Order("scores.performance_rating DESC").
		Limit(limit).
		Offset(page * limit).
		Find(&scores)

	if result.Error != nil {
		return nil, result.Error
	}

	if err := cacheScoreboard(scoreboardMods, md5, scores); err != nil {
		return nil, err
	}

	return scores, nil
}

// GetRateScoresForMap Retrieves the rate scores for a map
func GetRateScoresForMap(md5 string, mods int64, limit int, page int) ([]*Score, error) {
	cached, err := getCachedScoreboard(scoreboardRate, md5)

	if err != nil {
		return nil, err
	}

	if cached != nil {
		return cached, nil
	}

	var scores []*Score

	modsQuery := ""

	if mods == 0 {
		modsQuery = "AND (scores.mods = 0 OR scores.mods = ?) "
		mods = 2147483648 // TODO: USE ENUM
	} else {
		modsQuery = "AND (scores.mods & ?) != 0 "
	}

	result := SQL.
		Joins("User").
		Where("scores.map_md5 = ? "+
			"AND scores.failed = 0 "+
			modsQuery+
			"AND User.allowed = 1", md5, mods).
		Group("User.id").
		Order("scores.performance_rating DESC").
		Limit(limit).
		Offset(page * limit).
		Find(&scores)

	if result.Error != nil {
		return nil, result.Error
	}

	if err := cacheScoreboard(scoreboardRate, md5, scores); err != nil {
		return nil, err
	}

	return scores, nil
}

// GetAllScoresForMap Retrieves all scores for a map
func GetAllScoresForMap(md5 string, limit int, page int) ([]*Score, error) {
	cached, err := getCachedScoreboard(scoreboardAll, md5)

	if err != nil {
		return nil, err
	}

	if cached != nil {
		return cached, nil
	}

	var scores []*Score

	result := SQL.
		Joins("User").
		Where("scores.map_md5 = ? "+
			"AND scores.failed = 0 "+
			"AND User.allowed = 1", md5).
		Order("scores.performance_rating DESC").
		Group("User.id").
		Limit(limit).
		Offset(page * limit).
		Find(&scores)

	if result.Error != nil {
		return nil, result.Error
	}

	if err := cacheScoreboard(scoreboardAll, md5, scores); err != nil {
		return nil, err
	}

	return scores, nil
}

// GetFriendScoresForMap Retrieves the friend scores for a map
func GetFriendScoresForMap(md5 string, userId int, friends []*UserFriend, limit int, page int) ([]*Score, error) {
	friendLookup := fmt.Sprintf("AND (scores.user_id = %v", userId)

	for _, friend := range friends {
		friendLookup += fmt.Sprintf(" OR scores.user_id = %v", friend.Id)
	}

	friendLookup += ") "

	var scores []*Score

	result := SQL.
		Joins("User").
		Where("scores.map_md5 = ? "+
			"AND scores.personal_best = 1 "+
			friendLookup+
			"AND User.allowed = 1", md5).
		Order("scores.performance_rating DESC").
		Limit(limit).
		Offset(page * limit).
		Find(&scores)

	if result.Error != nil {
		return nil, result.Error
	}

	return scores, nil
}

// GetUserPersonalBestScoreGlobal Retrieves a user's personal best global score on a given map
func GetUserPersonalBestScoreGlobal(userId int, md5 string) (*Score, error) {
	var score *Score

	result := SQL.
		Joins("User").
		Where("scores.map_md5 = ? "+
			"AND scores.personal_best = 1 "+
			"AND User.id =  ? "+
			"AND User.allowed = 1", md5, userId).
		First(&score)

	if result.Error != nil {
		return nil, result.Error
	}

	return score, nil
}

// GetUserPersonalBestScoreAll Retrieves a user's personal best ALL score on a given map
func GetUserPersonalBestScoreAll(userId int, md5 string) (*Score, error) {
	var score *Score

	result := SQL.
		Joins("User").
		Where("scores.map_md5 = ? "+
			"AND scores.failed = 0 "+
			"AND User.id =  ? "+
			"AND User.allowed = 1", md5, userId).
		Order("scores.performance_rating DESC").
		First(&score)

	if result.Error != nil {
		return nil, result.Error
	}

	return score, nil
}

// GetUserPersonalBestScoreMods Retrieves a user's personal best modifier score on a given map
func GetUserPersonalBestScoreMods(userId int, md5 string, mods int64) (*Score, error) {
	var score *Score

	modsQueryStr := ""

	if mods == 0 {
		modsQueryStr = "AND scores.mods = ? "
	} else {
		modsQueryStr = "AND (scores.mods & ?) != 0 "
	}

	result := SQL.
		Joins("User").
		Where("scores.map_md5 = ? "+
			"AND scores.failed = 0 "+
			modsQueryStr+
			"AND User.id = ? "+
			"AND User.allowed = 1", md5, mods, userId).
		Order("scores.performance_rating DESC").
		First(&score)

	if result.Error != nil {
		return nil, result.Error
	}

	return score, nil
}

// GetUserPersonalBestScoreRate Retrieves a user's personal best rate score on a given map
func GetUserPersonalBestScoreRate(userId int, md5 string, mods int64) (*Score, error) {
	var score *Score

	modsQuery := ""

	if mods == 0 {
		modsQuery = "AND (scores.mods = 0 OR scores.mods = ?) "
		mods = 2147483648 // TODO: USE ENUM
	} else {
		modsQuery = "AND (scores.mods & ?) != 0 "
	}

	result := SQL.
		Joins("User").
		Where("scores.map_md5 = ? "+
			"AND scores.failed = 0 "+
			modsQuery+
			"AND User.id = ? "+
			"AND User.allowed = 1", md5, mods, userId).
		Order("scores.performance_rating DESC").
		First(&score)

	if result.Error != nil {
		return nil, result.Error
	}

	return score, nil
}

type scoreboardType string

const (
	scoreboardGlobal  scoreboardType = "global"
	scoreboardCountry scoreboardType = "country"
	scoreboardFriends scoreboardType = "friends"
	scoreboardMods    scoreboardType = "mods"
	scoreboardRate    scoreboardType = "rate"
	scoreboardAll     scoreboardType = "all"
)

// Returns the redis key for a scoreboard
func scoreboardRedisKey(md5 string, scoreboard scoreboardType) string {
	return fmt.Sprintf("quaver:scoreboard:%v:%v", md5, scoreboard)
}

// Caches a scoreboard to Redis
func cacheScoreboard(scoreboard scoreboardType, md5 string, scores []*Score) error {
	if len(scores) == 0 {
		return nil
	}

	scoresJson, err := json.Marshal(scores)

	if err != nil {
		return err
	}

	return Redis.Set(RedisCtx, scoreboardRedisKey(md5, scoreboard), scoresJson, time.Hour*24*3).Err()
}

// Retrieves a cached scoreboard from redis
func getCachedScoreboard(scoreboard scoreboardType, md5 string) ([]*Score, error) {
	result, err := Redis.Get(RedisCtx, scoreboardRedisKey(md5, scoreboard)).Result()

	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}

		return nil, err
	}

	if result == "" {
		return nil, nil
	}

	var scores []*Score

	if err := json.Unmarshal([]byte(result), &scores); err != nil {
		return nil, err
	}

	// User changed their name, country, or got banned
	for _, score := range scores {
		var user *User

		result := SQL.
			Select("allowed", "username", "country").
			Where("id = ?", score.UserId).
			First(&user)

		if result.Error != nil {
			return nil, result.Error
		}

		if !user.Allowed || user.Username != score.User.Username || user.Country != score.User.Country {
			return nil, nil
		}
	}

	return scores, nil
}
