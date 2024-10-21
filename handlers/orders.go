package handlers

import (
	"errors"
	"fmt"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/billingportal/session"
	"gorm.io/gorm"
	"net/http"
	"slices"
)

type DonatorPricing struct {
	Steam  DonatorPaymentMethod `json:"steam"`
	Stripe DonatorPaymentMethod `json:"stripe"`
}

type DonatorPaymentMethod struct {
	Months1  float32 `json:"months_1"`
	Months3  float32 `json:"months_3"`
	Months6  float32 `json:"months_6"`
	Months12 float32 `json:"months_12"`
}

type checkoutPaymentMethod int

const (
	paymentMethodSteam checkoutPaymentMethod = iota
	paymentMethodStripe
)

var (
	donatorPrices = DonatorPricing{
		Steam: DonatorPaymentMethod{
			Months1:  6.49,
			Months3:  17.99,
			Months6:  34.99,
			Months12: 64.99,
		},
		Stripe: DonatorPaymentMethod{
			Months1:  4.99,
			Months3:  13.99,
			Months6:  26.99,
			Months12: 49.99,
		},
	}
)

// Common request body when initiating a donator transaction
type donationRequestBody struct {
	Months        int    `form:"months" json:"months" binding:"required"`
	GiftUserId    int    `form:"gift_user_id" json:"gift_user_id"`
	Recurring     bool   `form:"recurring" json:"recurring"`
	Ip            string `form:"ip" json:"ip"`
	AnonymizeGift bool   `form:"anonymize_gift" json:"anonymize_gift"`
}

// GetDonatorPrices Returns the current donator prices
// Endpoint: GET /v2/orders/donations/prices
func GetDonatorPrices(c *gin.Context) *APIError {
	c.JSON(http.StatusOK, gin.H{"pricing": donatorPrices})
	return nil
}

// GetUserOrders Gets a user's completed orders.
// Endpoint: GET /v2/orders
func GetUserOrders(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	orders, err := db.GetUserOrders(user.Id)

	if err != nil {
		return APIErrorServerError("Error retrieving orders from db", err)
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
	return nil
}

// GetActiveSubscriptions Returns a user's active subscriptions
// GET /v2/orders/stripe/subscriptions
func GetActiveSubscriptions(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	activeSubs, err := db.GetUserStripeSubscriptions(user.Id)

	if err != nil {
		return APIErrorServerError("Error getting user stripe subscriptions", err)
	}

	c.JSON(http.StatusOK, gin.H{"subscriptions": activeSubs})
	return nil
}

// CreateOrderCheckoutSession Creates a checkout session for a given order (store items)
// Endpoint: POST /v2/orders/checkout
func CreateOrderCheckoutSession(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	body := struct {
		PaymentMethod checkoutPaymentMethod `json:"payment_method"`
		LineItems     []struct {
			Id            db.OrderItemId `json:"id"`
			Quantity      int            `json:"quantity"`
			GiftUserId    int            `json:"gift_user_id"`
			AnonymizeGift bool           `json:"anonymize_gift"`
		} `json:"line_items"`
		Ip string `json:"ip"`
	}{}

	if err := c.ShouldBindJSON(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	var orders []*db.Order
	var itemIds []db.OrderItemId

	for _, lineItem := range body.LineItems {
		if lineItem.Id == db.OrderItemDonator {
			return APIErrorBadRequest("You cannot purchase donator through this endpoint.")
		}

		if slices.Contains(itemIds, lineItem.Id) {
			return APIErrorBadRequest("You must submit 1 line_item per item id.")
		}

		orderItem, err := db.GetOrderItemById(lineItem.Id)

		if err != nil && err != gorm.ErrRecordNotFound {
			return APIErrorServerError("Error retrieving order item from db", err)
		}

		if orderItem == nil {
			return APIErrorBadRequest(fmt.Sprintf("Item with id %v does not exist", lineItem.Id))
		}

		if !orderItem.InStock {
			return APIErrorBadRequest(fmt.Sprintf("Item with id %v is no longer in stock", lineItem.Id))
		}

		if lineItem.Quantity == 0 || lineItem.Quantity > orderItem.MaxQuantityAllowed {
			return APIErrorBadRequest(fmt.Sprintf("Invalid item quantity for item %v", lineItem.Id))
		}

		if lineItem.GiftUserId != 0 && !orderItem.CanGift {
			return APIErrorBadRequest(fmt.Sprintf("You cannot gift item %v", lineItem.Id))
		}

		order := &db.Order{
			UserId:        user.Id,
			IPAddress:     getOrderIp(body.Ip),
			ItemId:        lineItem.Id,
			Quantity:      lineItem.Quantity,
			Description:   orderItem.Name,
			Item:          orderItem,
			AnonymizeGift: lineItem.AnonymizeGift,
		}

		isSet, err := order.SetReceiver(user, lineItem.GiftUserId)

		if err != nil {
			return APIErrorServerError("Error setting order receiver in db", err)
		}

		if !isSet {
			return APIErrorBadRequest(fmt.Sprintf("Gifting item %v to a user who doesn't exist", lineItem.Id))
		}

		if body.PaymentMethod == paymentMethodSteam {
			order.Amount = float32(orderItem.PriceSteam) / 100
		} else if body.PaymentMethod == paymentMethodStripe {
			order.Amount = float32(orderItem.PriceStripe) / 100
		}

		orders = append(orders, order)
		itemIds = append(itemIds, lineItem.Id)
	}

	switch body.PaymentMethod {
	case paymentMethodSteam:
		return createSteamCheckoutSession(c, user, orders)
	case paymentMethodStripe:
		return createStripeCheckoutSession(c, orders)
	}

	return APIErrorBadRequest("Invalid payment method provided")
}

// ModifyStripeSubscription Returns a link so the user can modify their subscription
// Endpoint: GET /v2/orders/stripe/subscriptions/modify
func ModifyStripeSubscription(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	activeSubs, err := db.GetUserStripeSubscriptions(user.Id)

	if err != nil {
		return APIErrorServerError("Error getting user active stripe subscriptions", err)
	}

	if len(activeSubs) == 0 {
		return APIErrorNotFound("Subscription")
	}

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(activeSubs[0].Customer.ID),
		ReturnURL: stripe.String(config.Instance.Stripe.DonateRedirectUrl),
	}

	result, err := session.New(params)

	if err != nil {
		return APIErrorServerError("Error creating stripe billing portal session", err)
	}

	c.JSON(http.StatusOK, gin.H{"url": result.URL})
	return nil
}

// Gets the donator price.
// Steam price = OG Price + 30%
func getDonatorPrice(months int, isSteam bool) (float32, error) {
	var paymentMethodPricing DonatorPaymentMethod

	if isSteam {
		paymentMethodPricing = donatorPrices.Steam
	} else {
		paymentMethodPricing = donatorPrices.Stripe
	}

	switch months {
	case 1:
		return paymentMethodPricing.Months1, nil
	case 3:
		return paymentMethodPricing.Months3, nil
	case 6:
		return paymentMethodPricing.Months6, nil
	case 12:
		return paymentMethodPricing.Months12, nil
	}

	return 0, errors.New("invalid donator months provided")
}

func getOrderIp(bodyIp string) string {
	if bodyIp != "" {
		return bodyIp
	}

	return "1.1.1.1"
}
