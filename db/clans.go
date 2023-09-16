package db

type Clan struct {
	Id                 int    `db:"id"`
	OwnerId            int    `db:"owner_id"`
	Name               string `db:"name"`
	Tag                string `db:"tag"`
	CreatedAt          int64  `db:"created_at"`
	AboutMe            string `db:"about_me"`
	FavoriteMode       uint8  `db:"favorite_mode"`
	LastNameChangeTime int64  `db:"last_name_change_time"`
}
