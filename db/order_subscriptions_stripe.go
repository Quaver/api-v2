package db

import (
	"github.com/Quaver/api2/config"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/subscription"
	"time"
)

type OrderSubscriptionStripe struct {
	Id                   int    `gorm:"column:id; PRIMARY_KEY" json:"id"`
	UserId               int    `gorm:"column:user_id" json:"user_id" json:"-"`
	StripeCustomerId     string `gorm:"column:customer_id" json:"stripe_customer_id"`
	StripeSubscriptionId string `gorm:"column:subscription_id" json:"stripe_subscription_id"`
	TimeCreated          int64  `gorm:"time_created" json:"-"`
	TimeLastUpdated      int64  `gorm:"time_last_updated" json:"-"`
	IsActive             bool   `gorm:"column:is_active" json:"-"`
}

func (*OrderSubscriptionStripe) TableName() string {
	return "order_subscriptions_stripe"
}

func (sub *OrderSubscriptionStripe) Insert() error {
	sub.IsActive = true
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

// GetUserActiveSubscriptions Retrieves a user's active stripe subscriptions
func GetUserActiveSubscriptions(userId int) ([]*OrderSubscriptionStripe, error) {
	var subscriptions = make([]*OrderSubscriptionStripe, 0)

	result := SQL.
		Where("user_id = ? AND is_active = 1", userId).
		Find(&subscriptions)

	if result.Error != nil {
		return nil, result.Error
	}

	return subscriptions, nil
}

// GetUserStripeSubscriptions Gets a user's active stripe subscriptions
func GetUserStripeSubscriptions(userId int) ([]*stripe.Subscription, error) {
	var activeSubs = make([]*stripe.Subscription, 0)

	subscriptions, err := GetUserActiveSubscriptions(userId)

	if err != nil {
		return nil, err
	}

	if len(subscriptions) == 0 {
		return []*stripe.Subscription{}, nil
	}

	for _, sub := range subscriptions {
		stripe.Key = config.Instance.Stripe.APIKey

		params := &stripe.SubscriptionListParams{Customer: stripe.String(sub.StripeCustomerId)}
		result := subscription.List(params)

		if result.Err() != nil {
			return nil, result.Err()
		}

		exists := result.Next()

		sub.IsActive = exists
		sub.TimeLastUpdated = time.Now().UnixMilli()

		if err := SQL.Save(sub).Error; err != nil {
			return nil, err
		}

		if !exists {
			continue
		}

		activeSub := result.Subscription()

		if activeSub != nil && activeSub.ID == sub.StripeSubscriptionId {
			activeSubs = append(activeSubs, activeSub)
		}
	}

	return activeSubs, nil
}
