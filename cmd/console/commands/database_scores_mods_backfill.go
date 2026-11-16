package commands

import (
	"fmt"
	"strings"

	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var DatabaseScoresModsBackfillCmd = &cobra.Command{
	Use:   "database:scores:mods:backfill",
	Short: "Backfill mods for old scores in ID order",
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		const batchSize = 5000

		nextID, err := cmd.Flags().GetInt64("start-id")
		if err != nil {
			return err
		}
		if nextID <= 0 {
			return fmt.Errorf("start-id must be positive")
		}

		var speedCase strings.Builder
		speedCase.WriteString("CASE")
		var speedArgs []interface{}
		for _, speed := range enums.ScoreSpeedRates {
			speedCase.WriteString(" WHEN (mods & ?) <> 0 THEN ?")
			speedArgs = append(speedArgs, int64(speed.Mod), speed.Rate)
		}
		speedCase.WriteString(" ELSE 100 END")

		updates := map[string]interface{}{
			"no_miss":              gorm.Expr("(mods & ?) <> 0", int64(enums.ModNoMiss)),
			"full_ln":              gorm.Expr("(mods & ?) <> 0", int64(enums.ModFullLN)),
			"inverse":              gorm.Expr("(mods & ?) <> 0", int64(enums.ModInverse)),
			"no_long_notes":        gorm.Expr("(mods & ?) <> 0", int64(enums.ModNoLongNotes)),
			"no_slider_velocities": gorm.Expr("(mods & ?) <> 0", int64(enums.ModNoSliderVelocities)),
			"mirror":               gorm.Expr("(mods & ?) <> 0", int64(enums.ModMirror)),
			"speed_rate":           gorm.Expr(speedCase.String(), speedArgs...),
			"mods_migrated":        true,
		}

		sql := db.SQL.WithContext(cmd.Context())
		var total int64
		logrus.Infof("Starting score mods backfill at ID %d (batch size %d)", nextID, batchSize)

		for {
			var scores []struct {
				ID           int64 `gorm:"column:id"`
				ModsMigrated bool  `gorm:"column:mods_migrated"`
			}
			if err := sql.Table("scores").
				Select("id, mods_migrated").
				Where("id >= ?", nextID).
				Order("id ASC").
				Limit(batchSize).
				Scan(&scores).Error; err != nil {
				return fmt.Errorf("read score mods batch (resume with --start-id %d): %w", nextID, err)
			}
			if len(scores) == 0 {
				logrus.Infof("Score mods backfill complete: no scores remain from ID %d; migrated %d scores", nextID, total)
				return nil
			}

			end := len(scores)
			for i, score := range scores {
				if score.ModsMigrated {
					end = i
					break
				}
			}

			if end > 0 {
				firstID, lastID := scores[0].ID, scores[end-1].ID
				result := sql.Table("scores").
					Where("id >= ? AND id <= ? AND mods_migrated = ?", firstID, lastID, false).
					Updates(updates)
				if result.Error != nil {
					return fmt.Errorf("backfill score mods IDs %d-%d (resume with --start-id %d): %w", firstID, lastID, nextID, result.Error)
				}

				total += result.RowsAffected
				nextID = lastID + 1
				logrus.Infof("Backfilled score mods IDs %d-%d: %d rows; total %d; next resume ID %d", firstID, lastID, result.RowsAffected, total, nextID)
			}

			if end < len(scores) {
				logrus.Infof("Score mods backfill stopped at migrated score ID %d; migrated %d scores", scores[end].ID, total)
				return nil
			}
		}
	},
}

func init() {
	DatabaseScoresModsBackfillCmd.Flags().Int64("start-id", 1, "First score ID to process (inclusive); use the logged next resume ID after interruption")
}
