package db

import (
	"gorm.io/gorm"
	"time"
)

type MapModComment struct {
	Id            int       `gorm:"column:id; PRIMARY_KEY" json:"id"`
	MapModId      int       `gorm:"column:map_mod_id" json:"map_mod_id"`
	AuthorId      int       `gorm:"column:author_id" json:"author_id"`
	Timestamp     int64     `gorm:"column:timestamp" json:"-"`
	TimestampJSON time.Time `gorm:"-:all" json:"timestamp"`
	Comment       string    `gorm:"column:comment" json:"comments"`
	Spam          bool      `gorm:"column:spam" json:"spam"`
	Author        *User     `gorm:"foreignKey:AuthorId; references:Id" json:"author"`
}

func (*MapModComment) TableName() string {
	return "map_mods_comments"
}

func (comment *MapModComment) AfterFind(*gorm.DB) (err error) {
	comment.TimestampJSON = time.UnixMilli(comment.Timestamp)
	return nil
}

// Insert Inserts a new mod comment into the database
func (comment *MapModComment) Insert() error {
	comment.Timestamp = time.Now().UnixMilli()

	if err := SQL.Create(&comment).Error; err != nil {
		return err
	}

	return nil
}
