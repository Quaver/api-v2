package db

import (
	"fmt"
	"slices"
	"time"

	"gorm.io/gorm"
)

type ChatMessage struct {
	Id       int `gorm:"column:id; PRIMARY_KEY" json:"id"`
	SenderId int `gorm:"column:sender_id" json:"sender_id"`
	// In case you are wondering why these are here instead of using User.ClanXXX,
	// In other places retrieving User, you need an additional user.SetClanTagAndColor()
	// to fill in those fields, which would be less efficient than just joining the tables
	ClanId          *int      `gorm:"column:clan_id" json:"clan_id"`
	ClanTag         *string   `gorm:"column:clan_tag" json:"clan_tag"`
	ClanAccentColor *string   `gorm:"column:clan_accent_color" json:"clan_accent_color"`
	Type            int8      `gorm:"column:type" json:"type"`
	ReceiverId      *int      `gorm:"column:receiver_id" json:"receiver_id"`
	Channel         string    `gorm:"column:channel" json:"channel"`
	Message         string    `gorm:"column:message" json:"message"`
	Timestamp       int64     `gorm:"column:timestamp" json:"-"`
	TimestampJSON   time.Time `gorm:"-:all" json:"timestamp"`
	IsHidden        bool      `gorm:"column:hidden" json:"-"`
	User            *User     `gorm:"foreignKey:SenderId" json:"user"`
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
		Select("chat_messages.*, `User`.clan_id AS clan_id, clans.tag AS clan_tag, clans.accent_color AS clan_accent_color").
		Joins("User").
		Joins("LEFT JOIN clans ON `User`.clan_id = clans.id").
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
		Select("chat_messages.*, `User`.clan_id AS clan_id, clans.tag AS clan_tag, clans.accent_color AS clan_accent_color").
		Joins("User").
		Joins("LEFT JOIN clans ON `User`.clan_id = clans.id").
		Where("chat_messages.hidden = 0 AND ("+
			"(chat_messages.sender_id = ? AND chat_messages.receiver_id = ?) OR "+
			"(chat_messages.sender_id = ? AND chat_messages.receiver_id = ?))",
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
