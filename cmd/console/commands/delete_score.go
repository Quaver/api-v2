package commands

import (
	"github.com/Quaver/api2/db"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
	"strconv"
)

var DeleteScoreCmd = &cobra.Command{
	Use:   "score:delete",
	Short: "Soft deletes a score from the db",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			logrus.Error("You must provide a score to delete")
		}

		id, err := strconv.Atoi(args[0])

		if err != nil {
			logrus.Error(err)
			return
		}

		score, err := db.GetScoreById(id)

		if err != nil && err != gorm.ErrRecordNotFound {
			logrus.Error(err)
			return
		}

		if score == nil {
			logrus.Error("Score not found")
			return
		}

		if err := score.SoftDelete(); err != nil {
			logrus.Error(err)
			return
		}

		logrus.Info("The provided score has been soft deleted.")
	},
}
