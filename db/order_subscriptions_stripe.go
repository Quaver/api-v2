package db

import "time"

type OrderSubscriptionStripe struct {
	Id                   int    `gorm:"column:id; PRIMARY_KEY"`
	UserId               int    `gorm:"column:user_id"`
	StripeCustomerId     string `gorm:"column:customer_id"`
	StripeSubscriptionId string `gorm:"column:subscription_id"`
	TimeCreated          int64  `gorm:"time_created"`
	TimeLastUpdated      int64  `gorm:"time_last_updated"`
}

func (*OrderSubscriptionStripe) TableName() string {
	return "order_subscriptions_stripe"
}

func (sub *OrderSubscriptionStripe) Insert() error {
	sub.TimeCreated = time.Now().UnixMilli()
	sub.TimeLastUpdated = time.Now().UnixMilli()

	return SQL.Create(&sub).Error
}

// GetOrderSubscriptionById Retrieves a subscription by id
func GetOrderSubscriptionById(subscriptionId string) (*OrderSubscriptionStripe, error) {
	var sub *OrderSubscriptionStripe

	result := SQL.
		Where("subscription_id = ?", subscriptionId).
		First(&sub)

	if result.Error != nil {
		return nil, result.Error
	}

	return sub, nil
}
