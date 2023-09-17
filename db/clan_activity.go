package db

type ClanActivity struct {
	Id        int    `gorm:"column:id; PRIMARY_KEY"`
	ClanId    int    `gorm:"column:clan_id"`
	Type      uint8  `gorm:"column:type"`
	UserId    int    `gorm:"column:user_id"`
	MapId     int    `gorm:"column:map_id"`
	Message   string `gorm:"column:message"`
	Timestamp int64  `gorm:"column:timestamp"`
}
