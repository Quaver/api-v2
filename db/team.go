package db

import (
	"github.com/Quaver/api2/enums"
	"time"
)

type Team struct {
	Developers         []*User `json:"developers"`
	Administrators     []*User `json:"administrators"`
	Moderators         []*User `json:"moderators"`
	RankingSupervisors []*User `json:"ranking_supervisors"`
	Contributors       []*User `json:"contributors"`
}

// GetTeamMembers Returns users in the Quaver team
func GetTeamMembers() (*Team, error) {
	var users = make([]*User, 0)
	var team = &Team{}

	err := CacheJsonInRedis("quaver:team", &users, time.Hour*24, false, func() error {
		result := SQL.
			Joins("StatsKeys4").
			Joins("StatsKeys7").
			Where("users.usergroups > 1 AND users.allowed = 1").
			Order("users.id ASC").
			Find(&users)

		if result.Error != nil {
			return result.Error
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if enums.HasUserGroup(user.UserGroups, enums.UserGroupDeveloper) {
			team.Developers = append(team.Developers, user)
		} else if enums.HasUserGroup(user.UserGroups, enums.UserGroupAdmin) {
			team.Administrators = append(team.Administrators, user)
		} else if enums.HasUserGroup(user.UserGroups, enums.UserGroupModerator) {
			team.Moderators = append(team.Moderators, user)
		} else if enums.HasUserGroup(user.UserGroups, enums.UserGroupRankingSupervisor) {
			team.RankingSupervisors = append(team.RankingSupervisors, user)
		} else if enums.HasUserGroup(user.UserGroups, enums.UserGroupContributor) {
			team.Contributors = append(team.Contributors, user)
		}
	}

	return team, nil
}

// GetRankingSupervisors Returns users who are Ranking Supervisors
func GetRankingSupervisors(ignoreCache bool) ([]*User, error) {
	var temp = make([]*User, 0)

	err := CacheJsonInRedis("quaver:supervisors", &temp, time.Hour*1, ignoreCache, func() error {
		result := SQL.
			Where("(users.usergroups & ? != 0) AND users.allowed = 1", enums.UserGroupRankingSupervisor).
			Order("users.id ASC").
			Find(&temp)

		if result.Error != nil {
			return result.Error
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	var users = make([]*User, 0)

	for _, user := range temp {
		if enums.HasUserGroup(user.UserGroups, enums.UserGroupDeveloper) ||
			enums.HasUserGroup(user.UserGroups, enums.UserGroupAdmin) {
			continue
		}

		users = append(users, user)
	}

	return users, nil
}
