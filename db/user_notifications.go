package db

import (
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

type UserNotificationType int

const (
	NotificationMapsetRanked UserNotificationType = iota + 1
	NotificationMapsetAction
	UserNotificationMapMod
	NotificationMapModComment
	NotificationClanInvite
	NotificationMuted
	NotificationReceivedOrderItemGift
	NotificationDonatorExpired
)

type UserNotificationCategory int

const (
	NotificationCategoryProfile UserNotificationCategory = iota + 1
	NotificationCategoryClan
	NotificationCategoryRankingQueue
	NotificationCategoryMapModding
)

const (
	QuaverBotId int = 2
)

type UserNotification struct {
	Id            int                      `gorm:"column:id; PRIMARY KEY" json:"id"`
	SenderId      int                      `gorm:"column:sender_id" json:"sender_id"`
	ReceiverId    int                      `gorm:"column:receiver_id" json:"receiver_id"`
	Type          UserNotificationType     `gorm:"column:type" json:"type"`
	Category      UserNotificationCategory `gorm:"column:category" json:"category"`
	RawData       string                   `gorm:"column:data" json:"-"`
	Data          json.RawMessage          `gorm:"-:all" json:"data"`
	ReadAt        int64                    `gorm:"column:read_at" json:"-"`
	ReadAtJSON    time.Time                `gorm:"-:all" json:"read_at"`
	Timestamp     int64                    `gorm:"column:timestamp" json:"-"`
	TimestampJSON time.Time                `gorm:"-:all" json:"timestamp"`
	User          *User                    `gorm:"foreignKey:SenderId; references:Id" json:"user"`
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

// GetNotifications Retrieves a user's notifications
func GetNotifications(userId int, unreadOnly bool, page int, limit int, category UserNotificationCategory) ([]*UserNotification, error) {
	notifications := make([]*UserNotification, 0)

	query := SQL.Where("receiver_id = ?", userId)

	if unreadOnly {
		query = query.Where("read_at = 0")
	}

	if category > 0 {
		query = query.Where("category = ?", category)
	}

	result := query.
		Preload("User").
		Order("timestamp DESC").
		Limit(limit).
		Offset(page * limit).
		Find(&notifications)

	if result.Error != nil {
		return nil, result.Error
	}

	return notifications, nil
}

// GetNotificationCount Gets the total amount of notifications that match a given filter
func GetNotificationCount(userId int, unreadOnly bool, category UserNotificationCategory) (int64, error) {
	var count int64

	query := SQL.Model(&UserNotification{}).Where("receiver_id = ?", userId)

	if unreadOnly {
		query = query.Where("read_at = 0")
	}

	if category > 0 {
		query = query.Where("category = ?", category)
	}

	result := query.Count(&count)

	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}

// GetTotalUnreadNotifications Retrieves the total amount of unread notifications the user has
func GetTotalUnreadNotifications(userId int) (int64, error) {
	var count int64

	result := SQL.
		Model(&UserNotification{}).
		Where("receiver_id = ?", userId).
		Where("read_at = 0").
		Count(&count)

	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}

// GetNotificationById Retrieves a user notification by id
func GetNotificationById(id int) (*UserNotification, error) {
	var notification *UserNotification

	result := SQL.
		Preload("User").
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

// NewMapsetRankedNotification Returns a new ranked mapset notification
func NewMapsetRankedNotification(mapset *Mapset) *UserNotification {
	notif := &UserNotification{
		SenderId:   QuaverBotId,
		ReceiverId: mapset.CreatorID,
		Type:       NotificationMapsetRanked,
		Category:   NotificationCategoryRankingQueue,
	}

	data := map[string]interface{}{
		"mapset_id":    mapset.Id,
		"mapset_title": mapset.String(),
	}

	marshaled, _ := json.Marshal(data)
	notif.RawData = string(marshaled)
	return notif
}

func NewMapsetActionNotification(mapset *Mapset, comment *MapsetRankingQueueComment) *UserNotification {
	notif := &UserNotification{
		SenderId:   comment.UserId,
		ReceiverId: mapset.CreatorID,
		Type:       NotificationMapsetAction,
		Category:   NotificationCategoryRankingQueue,
	}

	action := ""

	switch comment.ActionType {
	case RankingQueueActionComment:
		action = "commented on"
	case RankingQueueActionDeny:
		action = "denied"
	case RankingQueueActionBlacklist:
		action = "blacklisted"
	case RankingQueueActionOnHold:
		action = "put on-hold"
	case RankingQueueActionVote:
		action = "voted for"
	case RankingQueueActionResolved:
		action = "resolved"
	}

	data := map[string]interface{}{
		"mapset_id":    mapset.Id,
		"mapset_title": mapset.String(),
		"action":       action,
	}

	marshaled, _ := json.Marshal(data)
	notif.RawData = string(marshaled)
	return notif
}
