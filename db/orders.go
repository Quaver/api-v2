package db

import (
	"errors"
	"github.com/Quaver/api2/enums"
	"gorm.io/gorm"
	"time"
)

type OrderItemId int

const (
	OrderItemDonator OrderItemId = 1
)

type OrderStatus string

const (
	OrderStatusWaiting   OrderStatus = "Waiting"
	OrderStatusCompleted OrderStatus = "Completed"
)

type Order struct {
	Id                 int         `gorm:"column:id; PRIMARY_KEY" json:"id"`
	UserId             int         `gorm:"column:user_id" json:"user_id"`
	SteamOrderId       int         `gorm:"column:order_id" json:"-"`
	SteamTransactionId string      `gorm:"column:transaction_id" json:"-"`
	IPAddress          string      `gorm:"column:ip_address" json:"-"`
	ItemId             OrderItemId `gorm:"column:item_id" json:"item_id"`
	Quantity           int         `gorm:"column:quantity" json:"quantity"`
	Amount             float32     `gorm:"column:amount" json:"amount"`
	Description        string      `gorm:"column:description" json:"description"`
	ReceiverUserId     int         `gorm:"column:gifted_to" json:"-"`
	Receiver           *User       `gorm:"foreignKey:ReceiverUserId" json:"receiver"`
	Timestamp          int64       `gorm:"column:timestamp" json:"-"`
	TimestampJSON      time.Time   `gorm:"-:all" json:"timestamp"`
	Status             OrderStatus `gorm:"column:status" json:"status"`
}

func (*Order) TableName() string {
	return "orders"
}

func (order *Order) AfterFind(*gorm.DB) (err error) {
	order.TimestampJSON = time.UnixMilli(order.Timestamp)
	return nil
}

// Insert Inserts a new order into the database
func (order *Order) Insert() error {
	order.Timestamp = time.Now().UnixMilli()

	if err := SQL.Create(&order).Error; err != nil {
		return err
	}

	return nil
}

// Finalize Finalizes an order and grants the user their purchase items
func (order *Order) Finalize() error {
	if order.ItemId == OrderItemDonator {
		if err := order.FinalizeDonator(); err != nil {
			return err
		}
	}

	order.Status = "Completed"
	return SQL.Save(&order).Error
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
	timeAdded := int64(order.Quantity * 30 * 24 * 60 * 60 * 1000)

	if order.Receiver.DonatorEndTime == 0 {
		endTime = time.Now().UnixMilli() + timeAdded
	} else {
		endTime = order.Receiver.DonatorEndTime + timeAdded
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

// GetUserOrders Gets a user's orders
func GetUserOrders(userId int) ([]*Order, error) {
	var orders []*Order

	result := SQL.
		Preload("Receiver").
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
	var orders []*Order

	result := SQL.
		Preload("Receiver").
		Where("orders.order_id = ? AND orders.transaction_id = ?", steamOrderId, transactionId).
		Find(&orders)

	if result.Error != nil {
		return nil, result.Error
	}

	return orders, nil
}
