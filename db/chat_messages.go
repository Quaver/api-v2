package db

import (
	"fmt"
	"gorm.io/gorm"
	"slices"
	"time"
)

type ChatMessage struct {
	Id            int       `gorm:"column:id; PRIMARY_KEY" json:"id"`
	SenderId      int       `gorm:"column:sender_id" json:"sender_id"`
	Type          int8      `gorm:"type" json:"type"`
	ReceiverId    *int      `gorm:"receiver_id" json:"receiver_id"`
	Channel       string    `gorm:"channel" json:"channel"`
	Message       string    `gorm:"message" json:"message"`
	Timestamp     int64     `gorm:"timestamp" json:"-"`
	TimestampJSON time.Time `gorm:"-:all" json:"timestamp"`
	IsHidden      bool      `gorm:"column:hidden" json:"-"`
	User          *User     `gorm:"foreignKey:SenderId" json:"user"`
}

func (*ChatMessage) TableName() string {
	return "chat_messages"
}

func (m *ChatMessage) BeforeCreate(*gorm.DB) (err error) {
	t := time.Now()
	m.TimestampJSON = t
	return nil
}

func (m *ChatMessage) AfterFind(*gorm.DB) (err error) {
	m.TimestampJSON = time.UnixMilli(m.Timestamp)
	return nil
}

// GetPublicChatMessageHistory Gets the last 50 messages of a public chat
func GetPublicChatMessageHistory(channel string) ([]*ChatMessage, error) {
	var messages = make([]*ChatMessage, 0)

	result := SQL.
		Joins("User").
		Where("chat_messages.channel = ? AND chat_messages.hidden = 0", fmt.Sprintf("#%v", channel)).
		Limit(50).
		Order("chat_messages.id DESC").
		Find(&messages)

	if result.Error != nil {
		return nil, result.Error
	}

	slices.SortFunc(messages, func(a, b *ChatMessage) int {
		return a.Id - b.Id
	})

	return messages, nil
}

// GetPrivateChatMessageHistory Gets the last 50 messages of a private chat
func GetPrivateChatMessageHistory(userId int, otherUser int) ([]*ChatMessage, error) {
	var messages = make([]*ChatMessage, 0)

	result := SQL.
		Joins("User").
		Where("chat_messages.hidden = 0 AND "+
			"(chat_messages.sender_id = ? AND chat_messages.receiver_id = ?) OR "+
			"(chat_messages.sender_id = ? AND chat_messages.receiver_id = ?) ",
			userId, otherUser,
			otherUser, userId).
		Limit(50).
		Order("chat_messages.id DESC").
		Find(&messages)

	if result.Error != nil {
		return nil, result.Error
	}

	slices.SortFunc(messages, func(a, b *ChatMessage) int {
		return a.Id - b.Id
	})

	return messages, nil
}
