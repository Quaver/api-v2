package db

import "github.com/Quaver/api2/enums"

type RedisScore struct {
	Map struct {
		Id                   int                `json:"id"`
		MD5                  string             `json:"md5"`
		GameMode             enums.GameMode     `json:"game_mode"`
		RankedStatus         enums.RankedStatus `json:"ranked_status"`
		DifficultyRating     float64            `json:"difficulty_rating"`
		ClanRanked           bool               `json:"clan_ranked"`
		CountHitobjectNormal int                `json:"count_hitobject_normal"`
		CountHitobjectLong   int                `json:"count_hitobject_long"`
	} `json:"map"`
	Score struct {
		Id                int     `json:"id"`
		PerformanceRating float64 `json:"performance_rating"`
		PersonalBest      bool    `json:"personal_best"`
		Failed            bool    `json:"failed"`
		Accuracy          float64 `json:"accuracy"`
		CountMarvelous    int     `json:"count_marvelous"`
		CountPerfect      int     `json:"count_perfect"`
		CountGreat        int     `json:"count_great"`
		CountGood         int     `json:"count_good"`
		CountOkay         int     `json:"count_okay"`
		CountMiss         int     `json:"count_miss"`
	} `json:"score"`
	User struct {
		Id           int    `json:"id"`
		Username     string `json:"username"`
		Country      string `json:"country"`
		ShadowBanned bool   `json:"shadow_banned"`
		ClanId       int    `json:"clan_id"`
	} `json:"user"`
}
