package db

import (
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

type UserNotification struct {
	Id            int             `gorm:"column:id; PRIMARY KEY" json:"id"`
	SenderId      int             `gorm:"column:sender_id" json:"sender_id"`
	ReceiverId    int             `gorm:"column:receiver_id" json:"receiver_id"`
	Type          int8            `gorm:"column:type" json:"type"`
	RawData       string          `gorm:"column:data" json:"-"`
	Data          json.RawMessage `gorm:"-:all" json:"data"`
	ReadAt        int64           `gorm:"column:read_at" json:"-"`
	ReadAtJSON    time.Time       `gorm:"-:all" json:"-"`
	Timestamp     int64           `gorm:"column:timestamp" json:"-"`
	TimestampJSON time.Time       `gorm:"-:all" json:"timestamp"`
}

func (*UserNotification) TableName() string {
	return "user_notifications"
}

func (n *UserNotification) AfterFind(*gorm.DB) error {
	n.ReadAtJSON = time.UnixMilli(n.ReadAt)
	n.TimestampJSON = time.UnixMilli(n.Timestamp)

	if err := json.Unmarshal([]byte(n.RawData), &n.Data); err != nil {
		return err
	}

	return nil
}

func (n *UserNotification) Insert() error {
	n.Timestamp = time.Now().UnixMilli()
	return SQL.Create(&n).Error
}

// GetNotificationById Retrieves a user notification by id
func GetNotificationById(id int) (*UserNotification, error) {
	var notification *UserNotification

	result := SQL.
		Where("id = ?", id).
		First(&notification)

	if result.Error != nil {
		return nil, result.Error
	}

	return notification, nil
}

// UpdateReadStatus Updates the read status of a notification
func (n *UserNotification) UpdateReadStatus(isRead bool) error {
	if isRead {
		n.ReadAt = time.Now().UnixMilli()
	} else {
		n.ReadAt = 0
	}

	n.ReadAtJSON = time.UnixMilli(n.ReadAt)

	result := SQL.Model(&UserNotification{}).
		Where("id = ?", n.Id).
		Update("read_at", n.ReadAt)

	return result.Error
}
