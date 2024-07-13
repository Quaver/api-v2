package commands

import (
	"github.com/Quaver/api2/db"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ElasticIndexMapsets = &cobra.Command{
	Use:   "elastic:index:mapsets",
	Short: "Indexes all mapsets in Elastic Search",
	Run: func(cmd *cobra.Command, args []string) {
		if err := db.IndexAllElasticSearchMapsets(true); err != nil {
			logrus.Error(err)
			return
		}

		logrus.Info("Complete!")
	},
}
