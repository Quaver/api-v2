package db

import (
	"errors"
	"fmt"
	"github.com/Quaver/api2/enums"
	"gorm.io/gorm"
	"time"
)

type OrderStatus string

const (
	OrderStatusWaiting   OrderStatus = "Waiting"
	OrderStatusCompleted OrderStatus = "Completed"
)

type Order struct {
	Id             int                      `gorm:"column:id; PRIMARY_KEY" json:"id"`
	UserId         int                      `gorm:"column:user_id" json:"user_id"`
	OrderId        int                      `gorm:"column:order_id" json:"-"`
	TransactionId  string                   `gorm:"column:transaction_id" json:"-"`
	IPAddress      string                   `gorm:"column:ip_address" json:"-"`
	ItemId         OrderItemId              `gorm:"column:item_id" json:"item_id"`
	Quantity       int                      `gorm:"column:quantity" json:"quantity"`
	Amount         float32                  `gorm:"column:amount" json:"amount"`
	Description    string                   `gorm:"column:description" json:"description"`
	ReceiverUserId int                      `gorm:"column:gifted_to" json:"-"`
	Receiver       *User                    `gorm:"foreignKey:ReceiverUserId" json:"receiver"`
	Timestamp      int64                    `gorm:"column:timestamp" json:"-"`
	TimestampJSON  time.Time                `gorm:"-:all" json:"timestamp"`
	Status         OrderStatus              `gorm:"column:status" json:"status"`
	SubscriptionId *int                     `gorm:"column:subscription_id" json:"-"`
	AnonymizeGift  bool                     `gorm:"column:anonymize_gift" json:"anonymize_gift"`
	Item           *OrderItem               `gorm:"foreignKey:ItemId" json:"item"`
	Subscription   *OrderSubscriptionStripe `gorm:"foreignKey:SubscriptionId" json:"subscription,omitempty"`
}

func (*Order) TableName() string {
	return "orders"
}

func (order *Order) AfterFind(*gorm.DB) (err error) {
	order.TimestampJSON = time.UnixMilli(order.Timestamp)
	return nil
}

// SetReceiver Sets the receiver of the order. Checks if the user they are attempting to gift to exist.
// Returns if the receiver was successfully set and if there was a db error
func (order *Order) SetReceiver(payer *User, giftUserId int) (bool, error) {
	if giftUserId != 0 {
		receiver, err := GetUserById(giftUserId)

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return false, nil
			}

			return false, err
		}

		order.ReceiverUserId = giftUserId
		order.Receiver = receiver
		return true, nil
	}

	order.ReceiverUserId = payer.Id
	order.Receiver = payer
	return true, nil
}

// Insert Inserts a new order into the database
func (order *Order) Insert() error {
	order.Status = OrderStatusWaiting
	order.Timestamp = time.Now().UnixMilli()

	if err := SQL.Create(&order).Error; err != nil {
		return err
	}

	return nil
}

// Finalize Finalizes an order and grants the user their purchase items
func (order *Order) Finalize() error {
	var err error

	// Finalize items that have multiple products in its category
	switch order.Item.Category {
	case OrderItemCategoryDonator:
		err = order.FinalizeDonator()
	case OrderItemCategoryBadge:
		err = order.FinalizeBadge()
	}

	// Finalize Specific items
	switch order.Item.Id {
	case OrderItemClanCustomizable:
		err = order.FinalizeClanCustomizable()
	case OrderItemUserAccentColor:
		err = order.FinalizeUserAccentColor()
	}

	if err != nil {
		return err
	}

	order.Status = OrderStatusCompleted

	result := SQL.Model(&Order{}).
		Where("id = ?", order.Id).
		Update("status", order.Status)

	if result.Error != nil {
		return result.Error
	}

	if order.UserId != order.ReceiverUserId {
		if err := NewOrderItemGiftNotification(order).Insert(); err != nil {
			return err
		}
	}

	return nil
}

// FinalizeDonator Finalizes a donator item
func (order *Order) FinalizeDonator() error {
	if order.ItemId != OrderItemDonator {
		return errors.New("calling FinalizeDonator() on a non-donator item")
	}

	// Give usergroup if they don't already have it
	if !enums.HasUserGroup(order.Receiver.UserGroups, enums.UserGroupDonator) {
		if err := order.Receiver.UpdateUserGroups(order.Receiver.UserGroups | enums.UserGroupDonator); err != nil {
			return err
		}
	}

	// Extend Donator Time
	var endTime int64

	if order.Receiver.DonatorEndTime == 0 {
		endTime = time.Now().AddDate(0, order.Quantity, 0).UnixMilli()
	} else {
		endTime = time.UnixMilli(order.Receiver.DonatorEndTime).AddDate(0, order.Quantity, 0).UnixMilli()
	}

	if err := order.Receiver.UpdateDonatorEndTime(endTime); err != nil {
		return err
	}

	// Add Activity Log
	activity := &UserActivity{
		UserId:   order.ReceiverUserId,
		MapsetId: -1,
	}

	if order.UserId == order.ReceiverUserId {
		activity.Type = UserActivityDonated
	} else {
		activity.Type = UserActivityReceivedDonatorGift
	}

	if err := activity.Insert(); err != nil {
		return err
	}

	return nil
}

// FinalizeBadge Finalizes a badge order
func (order *Order) FinalizeBadge() error {
	if order.Item.Category != OrderItemCategoryBadge {
		return errors.New("cannot call FinalizeBadge() on a non badge order")
	}

	if order.Item.BadgeId == nil {
		return errors.New("badgeId in database is NULL")
	}

	userHasBadge, err := UserHasBadge(order.ReceiverUserId, *order.Item.BadgeId)

	if err != nil {
		return err
	}

	if userHasBadge {
		return nil
	}

	badge := &UserBadge{
		UserId:  order.ReceiverUserId,
		BadgeId: *order.Item.BadgeId,
	}

	if err := badge.Insert(); err != nil {
		return err
	}

	return nil
}

// FinalizeClanCustomizable Finalizes a clan customizable access order
func (order *Order) FinalizeClanCustomizable() error {
	if order.Item.Category != OrderItemCategoryClan || order.ItemId != OrderItemClanCustomizable {
		return errors.New("cannot call FinalizeClanCustomizable() on a mismatching order")
	}

	user, err := GetUserById(order.ReceiverUserId)

	if err != nil {
		return err
	}

	if user.ClanId == nil {
		return fmt.Errorf("couldnt finalize clan order for user: %v, as they are not in a clan", user.Id)
	}

	clan, err := GetClanById(*user.ClanId)

	if err != nil {
		return err
	}

	return clan.UpdateCustomizable(true)
}

// FinalizeUserAccentColor Finalizes a user accent color order
func (order *Order) FinalizeUserAccentColor() error {
	if order.Item.Category != OrderItemCategoryUserProfile || order.ItemId != OrderItemUserAccentColor {
		return errors.New("cannot call FinalizeUserAccentColor() on a mismatching order")
	}

	user, err := GetUserById(order.ReceiverUserId)

	if err != nil {
		return err
	}

	return UpdateUserAccentColorCustomizable(user.Id, true)
}

// GetUserOrders Gets a user's orders
func GetUserOrders(userId int) ([]*Order, error) {
	var orders = make([]*Order, 0)

	result := SQL.
		Preload("Receiver").
		Preload("Item").
		Preload("Subscription").
		Where("orders.user_id = ? AND orders.status = ?", userId, "Completed").
		Find(&orders)

	if result.Error != nil {
		return nil, result.Error
	}

	return orders, nil
}

// GetSteamOrdersByIds Retrieves orders by their steam order id & transaction id.
// Multiple orders in the database can have them if a user has multiple items in their cart.
func GetSteamOrdersByIds(steamOrderId string, transactionId string) ([]*Order, error) {
	var orders = make([]*Order, 0)

	result := SQL.
		Preload("Receiver").
		Preload("Item").
		Preload("Subscription").
		Where("orders.order_id = ? AND orders.transaction_id = ?", steamOrderId, transactionId).
		Find(&orders)

	if result.Error != nil {
		return nil, result.Error
	}

	return orders, nil
}

// GetStripeOrderById Gets Stripe orders by id
func GetStripeOrderById(transactionId string) ([]*Order, error) {
	var orders = make([]*Order, 0)

	result := SQL.
		Preload("Receiver").
		Preload("Item").
		Preload("Subscription").
		Where("orders.transaction_id = ?", transactionId).
		Find(&orders)

	if result.Error != nil {
		return nil, result.Error
	}

	return orders, nil
}
