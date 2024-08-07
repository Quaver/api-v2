package commands

import (
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/Quaver/api2/webhooks"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"time"
)

var SupervisorActivityCmd = &cobra.Command{
	Use:   "supervisor:activity",
	Short: "Handles providing donator for supervisor activity",
	Run: func(cmd *cobra.Command, args []string) {
		supervisors, err := db.GetRankingSupervisors(true)

		if err != nil {
			logrus.Error("Error retrieving supervisors from DB: ", err)
			return
		}

		if len(supervisors) == 0 {
			return
		}

		timeStart := time.Now().AddDate(0, 0, -7).UnixMilli()
		timeEnd := time.Now().UnixMilli()

		var userActions = map[*db.User]int{}

		for _, supervisor := range supervisors {
			actions, err := db.GetUserRankingQueueComments(supervisor.Id, timeStart, timeEnd)

			if err != nil {
				logrus.Error("Error retrieving ranking queue comments: ", err)
				return
			}

			userActions[supervisor] = len(actions)

			if len(actions) < config.Instance.RankingQueue.WeeklyRequiredSupervisorActions {
				continue
			}

			var endTime int64

			if supervisor.DonatorEndTime == 0 {
				endTime = time.Now().AddDate(0, 0, 7).UnixMilli()
			} else {
				endTime = time.UnixMilli(supervisor.DonatorEndTime).AddDate(0, 0, 7).UnixMilli()
			}

			if err := supervisor.UpdateDonatorEndTime(endTime); err != nil {
				logrus.Error("Error updating supervisor donator end time: ", err)
				return
			}

			logrus.Infof("[Supervisor Activity] Gave 1 week donator to: %v (#%v)", supervisor.Username, supervisor.Id)

			if enums.HasUserGroup(supervisor.UserGroups, enums.UserGroupDonator) {
				continue
			}

			if err := supervisor.UpdateUserGroups(supervisor.UserGroups | enums.UserGroupDonator); err != nil {
				logrus.Error("Error updating supervisor donator usergroup: ", err)
				return
			}

			logrus.Infof("[Supervisor Activity] Gave dono group to: %v (#%v)", supervisor.Username, supervisor.Id)
		}

		_ = webhooks.SendSupervisorActivityWebhook(userActions, timeStart, timeEnd)
	},
}
