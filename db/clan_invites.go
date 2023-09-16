package db

type ClanInvite struct {
	Id        int   `db:"id"`
	ClanId    int   `db:"clan_id"`
	UserId    int   `db:"user_id"`
	CreatedAt int64 `db:"created_at"`
}
