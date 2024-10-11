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
	NotificationMapMod
	NotificationMapModComment
	NotificationClanInvite
	NotificationMuted
	NotificationReceivedOrderItemGift
	NotificationDonatorExpired
	NotificationClanKicked
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
	ReadAtJSON    *time.Time               `gorm:"-:all" json:"read_at"`
	Timestamp     int64                    `gorm:"column:timestamp" json:"-"`
	TimestampJSON time.Time                `gorm:"-:all" json:"timestamp"`
	User          *User                    `gorm:"foreignKey:SenderId; references:Id" json:"user"`
}

func (*UserNotification) TableName() string {
	return "user_notifications"
}

func (n *UserNotification) AfterFind(*gorm.DB) error {
	if n.ReadAt > 0 {
		t := time.UnixMilli(n.ReadAt)
		n.ReadAtJSON = &t
	}

	n.TimestampJSON = time.UnixMilli(n.Timestamp)
	_ = json.Unmarshal([]byte(n.RawData), &n.Data)

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

	t := time.UnixMilli(n.ReadAt)
	n.ReadAtJSON = &t

	result := SQL.Model(&UserNotification{}).
		Where("id = ?", n.Id).
		Update("read_at", n.ReadAt)

	return result.Error
}

// MarkUserNotificationsAsRead Marks all of a user's notifications as read
func MarkUserNotificationsAsRead(userId int) error {
	return SQL.Model(&UserNotification{}).
		Where("receiver_id = ? AND read_at = 0", userId).
		Update("read_at", time.Now().UnixMilli()).Error
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

// Returns a new mapset ranking queue action notification
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

// NewMapModNotification Returns a new map mod notification
func NewMapModNotification(mapQua *MapQua, mod *MapMod) *UserNotification {
	notif := &UserNotification{
		SenderId:   mod.AuthorId,
		ReceiverId: mapQua.CreatorId,
		Type:       NotificationMapMod,
		Category:   NotificationCategoryMapModding,
	}

	data := map[string]interface{}{
		"mod_id":         mod.Id,
		"map_id":         mapQua.Id,
		"map_title":      mapQua.String(),
		"truncated_text": mod.Comment,
	}

	marshaled, _ := json.Marshal(data)
	notif.RawData = string(marshaled)
	return notif
}

// NewMapModCommentNotification Returns a new map mod comment notification
func NewMapModCommentNotification(mapQua *MapQua, mod *MapMod, comment *MapModComment) *UserNotification {
	notif := &UserNotification{
		SenderId:   comment.AuthorId,
		ReceiverId: mod.AuthorId,
		Type:       NotificationMapModComment,
		Category:   NotificationCategoryMapModding,
	}

	data := map[string]interface{}{
		"mod_id":         mod.Id,
		"mod_comment_id": comment.Id,
		"map_id":         mapQua.Id,
		"map_title":      mapQua.String(),
		"truncated_text": comment.Comment,
	}

	marshaled, _ := json.Marshal(data)
	notif.RawData = string(marshaled)
	return notif
}

// NewClanInviteNotification Returns a new clan invite notification
func NewClanInviteNotification(clan *Clan, invite *ClanInvite) *UserNotification {
	notif := &UserNotification{
		SenderId:   clan.OwnerId,
		ReceiverId: invite.UserId,
		Type:       NotificationClanInvite,
		Category:   NotificationCategoryClan,
	}

	data := map[string]interface{}{
		"clan_id":        clan.Id,
		"clan_name":      clan.Name,
		"clan_invite_id": invite.Id,
	}

	marshaled, _ := json.Marshal(data)
	notif.RawData = string(marshaled)
	return notif
}

// NewOrderItemGiftNotification Returns a new order gift notification
func NewOrderItemGiftNotification(order *Order) *UserNotification {
	notif := &UserNotification{
		SenderId:   order.UserId,
		ReceiverId: order.ReceiverUserId,
		Type:       NotificationReceivedOrderItemGift,
		Category:   NotificationCategoryProfile,
	}

	data := map[string]interface{}{
		"order_item_id":   order.ItemId,
		"order_item_name": order.Description,
		"anonymize_gift":  order.AnonymizeGift,
	}

	marshaled, _ := json.Marshal(data)
	notif.RawData = string(marshaled)
	return notif
}

// NewDonatorExpiredNotification Returns a new donator expired notification
func NewDonatorExpiredNotification(userId int) *UserNotification {
	notif := &UserNotification{
		SenderId:   QuaverBotId,
		ReceiverId: userId,
		Type:       NotificationDonatorExpired,
		Category:   NotificationCategoryProfile,
	}

	notif.RawData = ""
	return notif
}

// NewClanKickedNotification Returns a new clan kicked notification
func NewClanKickedNotification(clan *Clan, userId int) *UserNotification {
	notif := &UserNotification{
		SenderId:   clan.OwnerId,
		ReceiverId: userId,
		Type:       NotificationClanKicked,
		Category:   NotificationCategoryClan,
	}

	data := map[string]interface{}{
		"clan_id":   clan.Id,
		"clan_name": clan.Name,
	}

	marshaled, _ := json.Marshal(data)
	notif.RawData = string(marshaled)
	return notif
}
