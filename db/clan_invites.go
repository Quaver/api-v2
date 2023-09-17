package db

type ClanInvite struct {
	Id        uint  `gorm:"column:id; PRIMARY_KEY"`
	ClanId    int   `gorm:"column:clan_id"`
	UserId    int   `gorm:"column:user_id"`
	CreatedAt int64 `gorm:"column:created_at"`
}
