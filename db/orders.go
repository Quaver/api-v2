package db

import (
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

func (o *Order) AfterFind(*gorm.DB) (err error) {
	o.TimestampJSON = time.UnixMilli(o.Timestamp)
	return nil
}

// Insert Inserts a new order into the database
func (o *Order) Insert() error {
	o.Timestamp = time.Now().UnixMilli()

	if err := SQL.Create(&o).Error; err != nil {
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
