package db

type ClanInvite struct {
	Id        int   `gorm:"column:id"`
	ClanId    int   `gorm:"column:clan_id"`
	UserId    int   `gorm:"column:user_id"`
	CreatedAt int64 `gorm:"column:created_at"`
}
