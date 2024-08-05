package db

import (
	"fmt"
	"github.com/Quaver/api2/enums"
	"github.com/redis/go-redis/v9"
	"sort"
	"strconv"
	"strings"
)

func ClanLeaderboardKey(mode enums.GameMode) string {
	return fmt.Sprintf("quaver:clan_leaderboard:%v", mode)
}

// GetClanLeaderboardForMode Retrieves a clan leaderboard for a given game mode
func GetClanLeaderboardForMode(mode enums.GameMode, page int, limit int) ([]*Clan, error) {
	clanIds, err := Redis.ZRevRange(RedisCtx, ClanLeaderboardKey(mode), int64(page*limit), int64(page*limit+limit-1)).Result()

	if err != nil {
		return nil, err
	}

	if len(clanIds) == 0 {
		return []*Clan{}, nil
	}

	var clans = make([]*Clan, 0)

	result := SQL.
		Preload("Stats").
		Where(fmt.Sprintf("clans.id IN (%v)", strings.Join(clanIds, ","))).
		Find(&clans)

	if result.Error != nil {
		return nil, result.Error
	}

	sort.Slice(clans, func(i, j int) bool {
		return clans[i].Stats[mode-1].Rank < clans[j].Stats[mode-1].Rank
	})

	return clans, nil
}

// UpdateClanLeaderboard Adds a clan to a given leaderboard
func UpdateClanLeaderboard(clan *Clan, mode enums.GameMode) error {
	return Redis.ZAdd(RedisCtx, ClanLeaderboardKey(mode), redis.Z{
		Score:  clan.Stats[mode-1].OverallPerformanceRating,
		Member: strconv.Itoa(clan.Id),
	}).Err()
}

// RemoveClanFromLeaderboards Removes a clan from a given leaderboard
func RemoveClanFromLeaderboards(clanId int) error {
	for i := 1; i <= 2; i++ {
		err := Redis.ZRem(RedisCtx, ClanLeaderboardKey(enums.GameMode(i)), strconv.Itoa(clanId)).Err()

		if err != nil {
			return err
		}
	}

	return nil
}
