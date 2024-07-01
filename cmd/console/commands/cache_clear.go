package commands

import (
	"fmt"
	"github.com/Quaver/api2/db"
	"github.com/redis/go-redis/v9"
	"log"

	"github.com/spf13/cobra"
)

var CacheClearCmd = &cobra.Command{
	Use:   "cache:clear",
	Short: "Clears the cache",
	Long:  `Clears the application cache.`,
	Run: func(cmd *cobra.Command, args []string) {
		var cursor uint64

		keys, cursor, err := db.Redis.Scan(db.RedisCtx, cursor, fmt.Sprintf("quaver:*"), 0).Result()

		if err != nil && err != redis.Nil {
			fmt.Println(err)
			return
		}

		if len(keys) > 0 {
			delCount, err := db.Redis.Del(db.RedisCtx, keys...).Result()

			if err != nil {
				log.Fatalf("Failed to DELETE keys: %v", err)
			}

			fmt.Printf("Deleted %d keys\n", delCount)
		}

		if cursor == 0 {
			return
		}

		fmt.Println("Cache has been cleared.")
	},
}
