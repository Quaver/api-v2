package commands

import (
	"github.com/Quaver/api2/db"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strconv"
)

var UpdateStripePriceId = &cobra.Command{
	Use:   "update:stripe:id",
	Short: "Updates the stripe price id for an order item",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			logrus.Error("You must provide an order item id and a stripe price id.")
			return
		}

		id, err := strconv.Atoi(args[0])

		if err != nil {
			logrus.Error(err)
			return
		}

		item, err := db.GetOrderItemById(db.OrderItemId(id))

		if err != nil {
			logrus.Error(err)
			return
		}

		if item.Id == db.OrderItemDonator {
			logrus.Error("Cannot update donator price id")
			return
		}

		if err := item.UpdateStripePriceId(args[1]); err != nil {
			logrus.Error(err)
			return
		}

		logrus.Info("Done!")
	},
}
