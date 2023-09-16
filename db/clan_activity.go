package db

type ClanActivity struct {
	Id        int    `db:"id"`
	ClanId    int    `db:"clan_id"`
	Type      uint8  `db:"type"`
	UserId    int    `db:"user_id"`
	MapId     int    `db:"map_id"`
	Message   string `db:"message"`
	Timestamp int64  `db:"timestamp"`
}
