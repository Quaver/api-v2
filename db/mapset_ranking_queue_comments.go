package db

import "time"

type MapsetRankingQueueComment struct {
	Id            int       `gorm:"column:id; PRIMARY_KEY" json:"id"`
	UserId        int       `gorm:"column:user_id" json:"user_id"`
	MapsetId      int       `gorm:"column:mapset_id" json:"mapset_id"`
	Timestamp     int64     `gorm:"column:timestamp" json:"-"`
	TimestampJSON time.Time `gorm:"-:all" json:"timestamp"`
	Comment       string    `gorm:"comment" json:"comment"`
	User          *User     `gorm:"foreignKey:UserId; references:Id" json:"user"`
}

func (*MapsetRankingQueueComment) TableName() string {
	return "mapset_ranking_queue_comments"
}

// Insert Inserts a ranking queue comment into the database
func (c *MapsetRankingQueueComment) Insert() error {
	if err := SQL.Create(&c).Error; err != nil {
		return err
	}

	return nil
}
