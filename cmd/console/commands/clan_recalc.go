package commands

import (
	"github.com/Quaver/api2/db"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ClanRecalculateCommand = &cobra.Command{
	Use:   "clan:recalc",
	Short: "Recalculates all clans",
	Run: func(cmd *cobra.Command, args []string) {
		clans, err := db.GetClans()

		if err != nil {
			logrus.Error(err)
			return
		}

		for _, clan := range clans {
			if err := db.PerformFullClanRecalculation(clan); err != nil {
				logrus.Error(err)
				return
			}
		}
	},
}
