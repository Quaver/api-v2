package enums

type UserGroups int64

const (
	UserGroupNormal = 1 << iota
	UserGroupAdmin
	UserGroupBot
	UserGroupDeveloper
	UserGroupModerator
	UserGroupRankingSupervisor
	UserGroupSwan
	UserGroupContributor
	UserGroupDonator
)

// HasUserGroup Returns if a combination of user groups contains a single group
func HasUserGroup(groupsCombo UserGroups, group UserGroups) bool {
	return groupsCombo&group != 0
}
