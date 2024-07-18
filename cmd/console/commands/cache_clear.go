package commands

import (
	"github.com/Quaver/api2/db"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var CacheClearCmd = &cobra.Command{
	Use:   "cache:clear",
	Short: "Clears the cache",
	Long:  `Clears the application cache.`,
	Run: func(cmd *cobra.Command, args []string) {
		keys, err := db.Redis.Keys(db.RedisCtx, "quaver:*").Result()

		if err != nil && err != redis.Nil {
			logrus.Println(err)
			return
		}

		if len(keys) > 0 {
			delCount, err := db.Redis.Del(db.RedisCtx, keys...).Result()

			if err != nil {
				logrus.Errorf("Failed to DELETE keys: %v", err)
				return
			}

			logrus.Printf("Deleted %d keys\n", delCount)
		}

		logrus.Println("Cache has been cleared.")
	},
}
