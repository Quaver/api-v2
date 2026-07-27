package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/enums"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Score struct {
	Id                          int              `gorm:"column:id; PRIMARY_KEY" json:"id"`
	UserId                      int              `gorm:"column:user_id" json:"user_id"`
	MapMD5                      string           `gorm:"column:map_md5" json:"map_md5"`
	ReplayMD5                   string           `gorm:"column:replay_md5" json:"replay_md5"`
	Mode                        enums.GameMode   `gorm:"column:mode" json:"mode"`
	Timestamp                   int64            `gorm:"column:timestamp" json:"-"`
	TimestampJSON               time.Time        `gorm:"-:all" json:"timestamp"`
	IsPersonalBest              bool             `gorm:"column:personal_best" json:"is_personal_best"`
	PerformanceRating           float64          `gorm:"column:performance_rating" json:"performance_rating"`
	Modifiers                   int64            `gorm:"column:mods" json:"modifiers"`
	SpeedRate                   int              `gorm:"column:speed_rate;->" json:"speed_rate"`
	Mirror                      bool             `gorm:"column:mirror;->" json:"mirror"`
	NoSliderVelocities          bool             `gorm:"column:no_slider_velocities;->" json:"no_slider_velocities"`
	NoLongNotes                 bool             `gorm:"column:no_long_notes;->" json:"no_long_notes"`
	Inverse                     bool             `gorm:"column:inverse;->" json:"inverse"`
	FullLN                      bool             `gorm:"column:full_ln;->" json:"full_ln"`
	NoMiss                      bool             `gorm:"column:no_miss;->" json:"no_miss"`
	ModsMigrated                bool             `gorm:"column:mods_migrated;->" json:"-"`
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
	CountMineHit                int              `gorm:"column:count_minehit" json:"count_minehit"`
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

	if !s.ModsMigrated {
		mods := enums.Mods(s.Modifiers)
		s.SpeedRate = enums.GetSpeedRate(mods)
		s.Mirror = enums.IsModActivated(mods, enums.ModMirror)
		s.NoSliderVelocities = enums.IsModActivated(mods, enums.ModNoSliderVelocities)
		s.NoLongNotes = enums.IsModActivated(mods, enums.ModNoLongNotes)
		s.Inverse = enums.IsModActivated(mods, enums.ModInverse)
		s.FullLN = enums.IsModActivated(mods, enums.ModFullLN)
		s.NoMiss = enums.IsModActivated(mods, enums.ModNoMiss)
	}

	return nil
}

// GetUserBestScoresForMode Retrieves a user's best scores for a given game mode
func GetUserBestScoresForMode(id int, mode enums.GameMode, limit int, page int) ([]*Score, error) {
	var scores = make([]*Score, 0)

	result := SQL.
		Preload("Map").
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
	var scores = make([]*Score, 0)

	donatorScore := " AND scores.is_donator_score = 0"

	if isDonator {
		donatorScore = ""
	}

	result := SQL.
		Preload("Map").
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
	var scores = make([]*Score, 0)

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

	for _, score := range scores {
		_ = score.Map.AfterFind(SQL)
	}

	return scores, nil
}

// GetUserGradeScoresForMode Retrieves a user's scores for a particular grade
func GetUserGradeScoresForMode(id int, mode enums.GameMode, grade string, limit int, page int) ([]*Score, error) {
	var scores = make([]*Score, 0)

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

	for _, score := range scores {
		_ = score.Map.AfterFind(SQL)
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

// GetLastScoreId Retrieves the id of the last score submitted
func GetLastScoreId() (int, error) {
	var lastScoreId int

	if err := SQL.Raw("SELECT id FROM scores ORDER BY id DESC LIMIT 1;").Scan(&lastScoreId).Error; err != nil {
		return -1, err
	}

	return lastScoreId, nil
}

// GetGlobalScoresForMap Retrieves the global scores for a map
func GetGlobalScoresForMap(md5 string, useCache bool) ([]*Score, error) {
	if useCache {
		cached, err := getCachedScoreboard(scoreboardGlobal, md5, 0)

		if err != nil {
			return nil, err
		}

		if cached != nil {
			return cached, nil
		}
	}

	var scores = make([]*Score, 0)

	result := SQL.Raw(fmt.Sprintf(`
		WITH TopScores AS (
			SELECT scores.id
			FROM scores
			JOIN users filter_user ON scores.user_id = filter_user.id
			WHERE scores.map_md5 = ?
				AND scores.personal_best = 1
				AND filter_user.allowed = 1
			ORDER BY scores.performance_rating DESC
			LIMIT 100
		)
		SELECT %v
		FROM TopScores
		JOIN scores s ON s.id = TopScores.id
		JOIN users u ON s.user_id = u.id
		ORDER BY s.performance_rating DESC`, scoreboardScoreAndUserColumns), md5).
		Scan(&scores)

	if result.Error != nil {
		return nil, result.Error
	}

	for _, score := range scores {
		if err := score.AfterFind(SQL); err != nil {
			return nil, err
		}
	}

	if err := hydrateScoreboardUsers(scores); err != nil {
		return nil, err
	}

	if useCache {
		if err := cacheScoreboard(scoreboardGlobal, md5, scores, 0); err != nil {
			return nil, err
		}
	}

	return scores, nil
}

// GetCountryScoresForMap Retrieves the country scores for a map
func GetCountryScoresForMap(md5 string, country string) ([]*Score, error) {
	cached, err := getCachedScoreboard(scoreboardCountry, md5, 0)

	if err != nil {
		return nil, err
	}

	if cached != nil {
		return cached, nil
	}

	var scores = make([]*Score, 0)

	result := SQL.
		Joins("User").
		Where("scores.map_md5 = ? "+
			"AND scores.personal_best = 1 "+
			"AND User.country = ? "+
			"AND User.allowed = 1", md5, country).
		Order("scores.performance_rating DESC").
		Limit(100).
		Find(&scores)

	if result.Error != nil {
		return nil, result.Error
	}

	if err := hydrateScoreboardUsers(scores); err != nil {
		return nil, err
	}

	if err := cacheScoreboard(scoreboardCountry, md5, scores, 0); err != nil {
		return nil, err
	}

	return scores, nil
}

// GetModifierScoresForMap Retrieves the modifier scores for a map
func GetModifierScoresForMap(md5 string, mods int64) ([]*Score, error) {
	cached, err := getCachedScoreboard(scoreboardMods, md5, mods)

	if err != nil {
		return nil, err
	}

	if cached != nil {
		return cached, nil
	}

	var scores = make([]*Score, 0)
	modsQuery, modsArgs, err := getModifierScoreFilter("s", mods)

	if err != nil {
		return nil, err
	}

	args := append([]any{md5}, modsArgs...)

	result := SQL.Raw(fmt.Sprintf(`
		WITH RankedScores AS (
			SELECT 
				s.user_id,
				s.id AS score_id,
				ROW_NUMBER() OVER (PARTITION BY s.user_id ORDER BY s.performance_rating DESC, s.timestamp DESC) AS rnk
			FROM scores s
			WHERE 
				s.map_md5 = ?
			    %v
				AND s.failed = 0
		)
		%v`, modsQuery, getSelectUserScoreboardQuery(100)), args...).
		Scan(&scores)

	if result.Error != nil {
		return nil, result.Error
	}

	for _, score := range scores {
		if err := score.AfterFind(SQL); err != nil {
			return nil, err
		}
	}

	if err := hydrateScoreboardUsers(scores); err != nil {
		return nil, err
	}

	if err := cacheScoreboard(scoreboardMods, md5, scores, mods); err != nil {
		return nil, err
	}

	return scores, nil
}

// GetRateScoresForMap Retrieves the rate scores for a map
func GetRateScoresForMap(md5 string, mods int64) ([]*Score, error) {
	cached, err := getCachedScoreboard(scoreboardRate, md5, mods)

	if err != nil {
		return nil, err
	}

	if cached != nil {
		return cached, nil
	}

	var scores = make([]*Score, 0)
	rateQuery, rateArgs, err := getRateScoreFilter("s", mods)

	if err != nil {
		return nil, err
	}

	args := append([]any{md5}, rateArgs...)

	result := SQL.Raw(fmt.Sprintf(`
		WITH RankedScores AS (
			SELECT 
				s.user_id,
				s.id AS score_id,
				ROW_NUMBER() OVER (PARTITION BY s.user_id ORDER BY s.performance_rating DESC, s.timestamp DESC) AS rnk
			FROM scores s
			WHERE 
				s.map_md5 = ?
				AND s.failed = 0
				%v
		)
		%v`, rateQuery, getSelectUserScoreboardQuery(100)), args...).
		Scan(&scores)

	if result.Error != nil {
		return nil, result.Error
	}

	for _, score := range scores {
		if err := score.AfterFind(SQL); err != nil {
			return nil, err
		}
	}

	if err := hydrateScoreboardUsers(scores); err != nil {
		return nil, err
	}

	if err := cacheScoreboard(scoreboardRate, md5, scores, mods); err != nil {
		return nil, err
	}

	return scores, nil
}

// GetAllScoresForMap Retrieves all scores for a map
func GetAllScoresForMap(md5 string) ([]*Score, error) {
	cached, err := getCachedScoreboard(scoreboardAll, md5, 0)

	if err != nil {
		return nil, err
	}

	if cached != nil {
		return cached, nil
	}

	var scores = make([]*Score, 0)

	result := SQL.Raw(fmt.Sprintf(`
		WITH RankedScores AS (
			SELECT 
				s.user_id,
				s.id AS score_id,
				ROW_NUMBER() OVER (PARTITION BY s.user_id ORDER BY s.performance_rating DESC, s.timestamp DESC) AS rnk
			FROM scores s
			WHERE 
				s.map_md5 = ?
				AND s.failed = 0
		)
		%v`, getSelectUserScoreboardQuery(100)), md5).
		Scan(&scores)

	if result.Error != nil {
		return nil, result.Error
	}

	for _, score := range scores {
		if err := score.AfterFind(SQL); err != nil {
			return nil, err
		}
	}

	if err := hydrateScoreboardUsers(scores); err != nil {
		return nil, err
	}

	if err := cacheScoreboard(scoreboardAll, md5, scores, 0); err != nil {
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

	var scores = make([]*Score, 0)

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

	if err := score.User.AfterFind(nil); err != nil {
		return nil, err
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

	if err := score.User.AfterFind(nil); err != nil {
		return nil, err
	}

	return score, nil
}

// GetUserPersonalBestScoreMods Retrieves a user's personal best modifier score on a given map
func GetUserPersonalBestScoreMods(userId int, md5 string, mods int64) (*Score, error) {
	var score *Score
	modsQuery, modsArgs, err := getModifierScoreFilter("scores", mods)

	if err != nil {
		return nil, err
	}

	args := append([]any{md5}, modsArgs...)
	args = append(args, userId)

	result := SQL.
		Joins("User").
		Where("scores.map_md5 = ? "+
			"AND scores.failed = 0 "+
			modsQuery+
			"AND User.id = ? "+
			"AND User.allowed = 1", args...).
		Order("scores.performance_rating DESC").
		First(&score)

	if result.Error != nil {
		return nil, result.Error
	}

	if err := score.User.AfterFind(nil); err != nil {
		return nil, err
	}

	return score, nil
}

// GetUserPersonalBestScoreRate Retrieves a user's personal best rate score on a given map
func GetUserPersonalBestScoreRate(userId int, md5 string, mods int64) (*Score, error) {
	var score *Score
	rateQuery, rateArgs, err := getRateScoreFilter("scores", mods)

	if err != nil {
		return nil, err
	}

	args := append([]any{md5}, rateArgs...)
	args = append(args, userId)

	result := SQL.
		Joins("User").
		Where("scores.map_md5 = ? "+
			"AND scores.failed = 0 "+
			rateQuery+
			"AND User.id = ? "+
			"AND User.allowed = 1", args...).
		Order("scores.performance_rating DESC").
		First(&score)

	if result.Error != nil {
		return nil, result.Error
	}

	if err := score.User.AfterFind(nil); err != nil {
		return nil, err
	}

	return score, nil
}

// GetClanPlayerScoresOnMap Fetches the top 10 scores from a clan on a given map
func GetClanPlayerScoresOnMap(md5 string, clanId int, callAfterFind bool) ([]*Score, error) {
	scores := make([]*Score, 0)

	result := SQL.Raw(fmt.Sprintf(`
		WITH RankedScores AS (
			SELECT 
				s.user_id,
				s.id AS score_id,
				ROW_NUMBER() OVER (PARTITION BY s.user_id ORDER BY s.performance_rating DESC, s.timestamp DESC) AS rnk
			FROM scores s
			WHERE 
				s.map_md5 = ?
				AND s.clan_id = ?
				AND s.failed = 0
		)
		%v`, getSelectUserScoreboardQuery(10, false)), md5, clanId).
		Scan(&scores)

	if result.Error != nil {
		return nil, result.Error
	}

	if callAfterFind {
		for _, score := range scores {
			_ = score.AfterFind(nil)
			_ = score.User.AfterFind(nil)
		}
	}

	return scores, nil
}

// RemoveUserClanScores Removes all user scores from a clan
func RemoveUserClanScores(clanId int, userId int) error {
	return SQL.Model(&Score{}).
		Where("user_id = ? AND clan_id = ?", userId, clanId).
		Update("clan_id", nil).Error
}

func (s *Score) SoftDelete() error {
	return SQL.Model(&Score{}).
		Where("id = ?", s.Id).
		Update("personal_best", 0).
		Update("is_donator_score", 0).Error
}

// CalculateOverallRating Calculates overall rating from a list of scores
func CalculateOverallRating(scores []*Score) float64 {
	if len(scores) == 0 {
		return 0
	}

	sum := 0.00

	for i, score := range scores {
		sum += score.PerformanceRating * math.Pow(0.95, float64(i))
	}

	return sum
}

// CalculateOverallAccuracy Calculates overall accuracy from a list of scores
func CalculateOverallAccuracy(scores []*Score) float64 {
	if len(scores) == 0 {
		return 0
	}

	var total float64
	var divideTotal float64

	for i, score := range scores {
		add := math.Pow(0.95, float64(i)) * 100
		total += score.Accuracy * add
		divideTotal += add
	}

	if divideTotal == 0 {
		return 0
	}

	return total / divideTotal
}

type scoreboardType string

const (
	scoreboardCacheVersion = "v2"

	scoreboardGlobal  scoreboardType = "global"
	scoreboardCountry scoreboardType = "country"
	scoreboardFriends scoreboardType = "friends"
	scoreboardMods    scoreboardType = "mods"
	scoreboardRate    scoreboardType = "rate"
	scoreboardAll     scoreboardType = "all"
)

// Returns the redis key for a scoreboard
func scoreboardRedisKey(md5 string, scoreboard scoreboardType, mods int64) string {
	switch scoreboard {
	case scoreboardMods, scoreboardRate:
		filterMode := "legacy"

		if scoreModColumnsReady() {
			filterMode = "columns"
		}

		return fmt.Sprintf(
			"quaver:scoreboard:%v:%v:%v:%v:%v",
			scoreboardCacheVersion,
			md5,
			scoreboard,
			mods,
			filterMode,
		)
	default:
		return fmt.Sprintf("quaver:scoreboard:%v:%v:%v", scoreboardCacheVersion, md5, scoreboard)
	}
}

// Caches a scoreboard to Redis
func cacheScoreboard(scoreboard scoreboardType, md5 string, scores []*Score, mods int64) error {
	if len(scores) == 0 {
		return nil
	}

	scoresJson, err := json.Marshal(scores)

	if err != nil {
		return err
	}

	return Redis.Set(RedisCtx, scoreboardRedisKey(md5, scoreboard, mods), scoresJson, time.Hour*24*3).Err()
}

// Retrieves a cached scoreboard from redis
func getCachedScoreboard(scoreboard scoreboardType, md5 string, mods int64) ([]*Score, error) {
	result, err := Redis.Get(RedisCtx, scoreboardRedisKey(md5, scoreboard, mods)).Result()

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

	// Cache invalidation by checking for updated values
	for _, score := range scores {
		var user *User

		result := SQL.
			Select("allowed", "username", "country", "clan_id").
			Where("id = ?", score.UserId).
			First(&user)

		if result.Error != nil {
			return nil, result.Error
		}

		if err := user.SetClanTagAndColor(); err != nil {
			return nil, result.Error
		}

		if !user.Allowed ||
			user.Username != score.User.Username ||
			user.Country != score.User.Country ||
			!comparePointers(user.ClanId, score.User.ClanId) ||
			!comparePointers(user.ClanTag, score.User.ClanTag) ||
			!comparePointers(user.ClanAccentColor, score.User.ClanAccentColor) {
			return nil, nil
		}

	}

	return scores, nil
}

const scoreboardScoreAndUserColumns = `s.user_id,
			   s.*,
			   u.id AS User__id,
			   u.steam_id AS User__steam_id,
			   u.username AS User__username,
			   u.time_registered AS User__time_registered,
			   u.allowed AS User__allowed,
			   u.privileges AS User__privileges,
			   u.usergroups AS User__usergroups,
			   u.mute_endtime AS User__mute_endtime,
			   u.latest_activity AS User__latest_activity,
			   u.country AS User__country,
			   u.avatar_url AS User__avatar_url,
			   u.twitter AS User__twitter,
			   u.title AS User__title,
			   u.userpage AS User__userpage,
			   u.twitch_username AS User__twitch_username,
			   u.donator_end_time AS User__donator_end_time,
			   u.discord_id AS User__discord_id,
			   u.information AS User__information,
			   u.clan_id AS User__clan_id,
			   u.clan_leave_time AS User__clan_leave_time,
			   u.accent_color_customizable AS User__accent_color_customizable,
			   u.accent_color AS User__accent_color`

// Returns a query to select user scores from non personal best scoreboards.
func getSelectUserScoreboardQuery(limit int, donatorOnly ...bool) string {
	query := fmt.Sprintf(`
		SELECT %v
				FROM RankedScores rs
				JOIN scores s ON s.id = rs.score_id
				JOIN users u ON s.user_id = u.id
				JOIN maps m ON s.map_md5 = m.md5
				WHERE rs.rnk = 1 AND u.allowed = 1
		`, scoreboardScoreAndUserColumns)
	if len(donatorOnly) > 0 && donatorOnly[0] == true {
		query += " AND u.donator_end_time > 0"
	}

	query += `
		ORDER BY s.performance_rating DESC
		LIMIT`

	query += fmt.Sprintf(" %v;", limit)
	return query
}

func comparePointers[T comparable](a, b *T) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	return *a == *b
}

func hydrateScoreboardUsers(scores []*Score) error {
	usersByID := make(map[int][]*User, len(scores))
	usersByClanID := make(map[int][]*User)

	for _, score := range scores {
		if score.User == nil {
			continue
		}

		user := score.User

		if err := user.populateAfterFindFields(); err != nil {
			continue
		}

		if user.StatsKeys4 != nil {
			if ranks, err := GetUserRanksForMode(user, enums.GameModeKeys4); err == nil {
				user.StatsKeys4.Ranks = ranks
			}
		}

		if user.StatsKeys7 != nil {
			if ranks, err := GetUserRanksForMode(user, enums.GameModeKeys7); err == nil {
				user.StatsKeys7.Ranks = ranks
			}
		}

		usersByID[user.Id] = append(usersByID[user.Id], user)

		if user.ClanId != nil {
			usersByClanID[*user.ClanId] = append(usersByClanID[*user.ClanId], user)
		}
	}

	if len(usersByID) == 0 {
		return nil
	}

	userIDs := make([]int, 0, len(usersByID))

	for userID := range usersByID {
		userIDs = append(userIDs, userID)
	}

	if statuses, err := getUserClientStatuses(userIDs); err == nil {
		for userID, status := range statuses {
			for _, user := range usersByID[userID] {
				user.ClientStatus = status
			}
		}
	}

	if len(usersByClanID) == 0 {
		return nil
	}

	clanIDs := make([]int, 0, len(usersByClanID))

	for clanID := range usersByClanID {
		clanIDs = append(clanIDs, clanID)
	}

	var clans []Clan
	result := SQL.
		Select("id", "tag", "accent_color").
		Where("id IN ?", clanIDs).
		Find(&clans)

	if result.Error != nil {
		return result.Error
	}

	clansByID := make(map[int]Clan, len(clans))

	for _, clan := range clans {
		clansByID[clan.Id] = clan
	}

	for clanID, users := range usersByClanID {
		clan := clansByID[clanID]

		for _, user := range users {
			tag := clan.Tag
			user.ClanTag = &tag
			user.ClanAccentColor = clan.AccentColor
		}
	}

	return nil
}

var scoreModifierColumns = map[enums.Mods]string{
	enums.ModNoSliderVelocities: "no_slider_velocities",
	enums.ModNoLongNotes:        "no_long_notes",
	enums.ModInverse:            "inverse",
	enums.ModFullLN:             "full_ln",
	enums.ModMirror:             "mirror",
	enums.ModNoMiss:             "no_miss",
}

var ErrUnsupportedScoreModifierFilter = errors.New("modifier filter is not supported by score lookup columns")

func scoreModColumnsReady() bool {
	return config.Instance != nil && config.Instance.ScoreModColumnsReady
}

func getModifierScoreFilter(tableAlias string, mods int64) (string, []any, error) {
	if mods == 0 {
		return fmt.Sprintf("AND %v.mods = ? ", tableAlias), []any{int64(0)}, nil
	}

	if !scoreModColumnsReady() {
		return fmt.Sprintf("AND (%v.mods & ?) != 0 ", tableAlias), []any{mods}, nil
	}

	columns := make([]string, 0)

	for i := 0; (1 << i) < enums.ModEnumMaxValue-1; i++ {
		mod := enums.Mods(1 << i)

		if !enums.IsModActivated(enums.Mods(mods), mod) {
			continue
		}

		column, ok := scoreModifierColumns[mod]

		if !ok {
			return "", nil, ErrUnsupportedScoreModifierFilter
		}

		columns = append(columns, fmt.Sprintf("AND %v.%v = 1 ", tableAlias, column))
	}

	if len(columns) == 0 {
		return "", nil, ErrUnsupportedScoreModifierFilter
	}

	query := ""

	for _, column := range columns {
		query += column
	}

	return query, []any{}, nil
}

func getRateScoreFilter(tableAlias string, mods int64) (string, []any, error) {
	if !scoreModColumnsReady() {
		if mods == 0 {
			return fmt.Sprintf(
				"AND (%v.mods = 0 OR %v.mods = ?) ",
				tableAlias,
				tableAlias,
			), []any{int64(enums.ModMirror)}, nil
		}

		return fmt.Sprintf("AND (%v.mods & ?) != 0 ", tableAlias), []any{mods}, nil
	}

	modCombo := enums.Mods(mods)
	speedRate := enums.GetSpeedRate(modCombo)

	if mods == 0 {
		return fmt.Sprintf("AND %v.speed_rate = ? ", tableAlias), []any{speedRate}, nil
	}

	_, ok := enums.GetSpeedMod(modCombo)

	if !ok {
		return "", nil, ErrUnsupportedScoreModifierFilter
	}

	return fmt.Sprintf("AND %v.speed_rate = ? ", tableAlias), []any{speedRate}, nil
}
