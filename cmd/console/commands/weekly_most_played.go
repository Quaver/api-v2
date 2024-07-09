package commands

import (
	"github.com/Quaver/api2/db"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var WeeklyMostPlayedMapsetsCmd = &cobra.Command{
	Use:   "mapsets:weekly:cache",
	Short: "Caches the most played mapsets of the week in redis",
	Run: func(cmd *cobra.Command, args []string) {
		mapsets, err := db.GetWeeklyMostPlayedMapsets(true)

		if err != nil {
			logrus.Error("Error caching weekly most played mapsets", err)
		}

		logrus.Infof("Successfully cached: %v weekly most played mapsets", len(mapsets))
	},
}
