package commands

import (
	"github.com/Quaver/api2/db"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strconv"
)

var BadgePlayerGiveCmd = &cobra.Command{
	Use:   "badge:player:give",
	Short: "Gives a player a badge",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			logrus.Error("You must provide a badge id and a player id")
			return
		}

		badgeId, err := strconv.Atoi(args[0])

		if err != nil {
			logrus.Error(err)
			return
		}

		playerId, err := strconv.Atoi(args[1])

		if err != nil {
			logrus.Error(err)
			return
		}

		_, err = db.GetUserById(playerId)

		if err != nil {
			logrus.Error("Error retrieving player: ", err)
			return
		}

		hasBadge, err := db.UserHasBadge(playerId, badgeId)

		if err != nil {
			logrus.Error("Error checking user has badge: ", err)
			return
		}

		if hasBadge {
			logrus.Info("User already has badge")
			return
		}

		badge := &db.UserBadge{
			UserId:  playerId,
			BadgeId: badgeId,
		}

		if err := badge.Insert(); err != nil {
			logrus.Error("Error inserting badge: ", err)
			return
		}

		logrus.Info("Done!")
	},
}