package commands

import (
	"fmt"
	"time"

	"github.com/Quaver/api2/db"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var scoreModsBackfillBatchSize int
var scoreModsBackfillSleepMs int
var scoreModsBackfillMaxBatches int64

var DatabaseScoresModsBackfill = &cobra.Command{
	Use:   "database:scores:mods:backfill",
	Short: "Backfills score modifier lookup columns in batches",
	RunE: func(cmd *cobra.Command, args []string) error {
		if scoreModsBackfillBatchSize <= 0 {
			return fmt.Errorf("batch-size must be greater than 0")
		}

		if scoreModsBackfillSleepMs < 0 {
			return fmt.Errorf("sleep-ms cannot be negative")
		}

		if scoreModsBackfillMaxBatches < 0 {
			return fmt.Errorf("max-batches cannot be negative")
		}

		var completedBatches int64
		var totalUpdated int64
		lastScoreID := 0

		logrus.Infof(
			"Backfilling score modifier columns for all maps with batches of %d",
			scoreModsBackfillBatchSize,
		)

		for {
			if scoreModsBackfillMaxBatches > 0 && completedBatches >= scoreModsBackfillMaxBatches {
				logrus.Infof(
					"Reached max-batches=%d after updating %d rows. Run the command again to resume.",
					scoreModsBackfillMaxBatches,
					totalUpdated,
				)
				return nil
			}

			var scoreIDs []int

			result := db.SQL.WithContext(cmd.Context()).
				Model(&db.Score{}).
				Where("mods_migrated = 0 AND id > ?", lastScoreID).
				Order("id ASC").
				Limit(scoreModsBackfillBatchSize).
				Pluck("id", &scoreIDs)

			if result.Error != nil {
				return fmt.Errorf("retrieve unmigrated scores after score id %d: %w", lastScoreID, result.Error)
			}

			if len(scoreIDs) == 0 {
				break
			}

			result = db.SQL.WithContext(cmd.Context()).
				Model(&db.Score{}).
				Where("id IN ? AND mods_migrated = 0", scoreIDs).
				UpdateColumn("mods", gorm.Expr("mods"))

			if result.Error != nil {
				return fmt.Errorf("backfill score modifier columns through score id %d: %w", scoreIDs[len(scoreIDs)-1], result.Error)
			}

			lastScoreID = scoreIDs[len(scoreIDs)-1]
			completedBatches++
			totalUpdated += result.RowsAffected

			logrus.Infof(
				"Backfilled score modifier columns through score id %d (%d rows, %d total, %d batches complete)",
				lastScoreID,
				result.RowsAffected,
				totalUpdated,
				completedBatches,
			)

			if scoreModsBackfillSleepMs > 0 {
				time.Sleep(time.Duration(scoreModsBackfillSleepMs) * time.Millisecond)
			}
		}

		logrus.Infof(
			"Score modifier column backfill complete. Processed %d batches and updated %d rows.",
			completedBatches,
			totalUpdated,
		)
		logrus.Info("Set score_mod_columns_ready=true and restart all API instances to enable indexed score modifier lookups.")

		return nil
	},
}

func init() {
	DatabaseScoresModsBackfill.Flags().IntVar(&scoreModsBackfillBatchSize, "batch-size", 5000, "Maximum number of scores to backfill per batch")
	DatabaseScoresModsBackfill.Flags().IntVar(&scoreModsBackfillSleepMs, "sleep-ms", 0, "Milliseconds to sleep between batches")
	DatabaseScoresModsBackfill.Flags().Int64Var(&scoreModsBackfillMaxBatches, "max-batches", 0, "Maximum number of batches to process before exiting")
}
