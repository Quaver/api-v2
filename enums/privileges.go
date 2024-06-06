package enums

type Privileges int64

const (
	PrivilegeNormal Privileges = 1 << iota
	PrivilegeKickUsers
	PrivilegeBanUsers
	PrivilegeNotifyUsers
	PrivilegeMuteUsers
	PrivilegeRankMapsets
	PrivilegeViewAdminLogs
	PrivilegeEditUsers
	PrivilegeManageBuilds
	PrivilegeManageAlphaKeys
	PrivilegeManageMapsets
	PrivilegeEnableTournamentMode
	PrivilegeWipeUsers
	PrivilegeEditUsername
	PrivilegeEditFlag
	PrivilegeEditPrivileges
	PrivilegeEditGroups
	PrivilegeEditNotes
	PrivilegeEditAvatar
	PrivilegeViewCrashes
	PrivilegeEditDonate
)

// HasPrivilege Returns if a combination of user groups contains a single group
func HasPrivilege(privilegesCombo Privileges, privilege Privileges) bool {
	return privilegesCombo&privilege != 0
}
