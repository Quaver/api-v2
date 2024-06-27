package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/webhooks"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/checkout/session"
	"github.com/stripe/stripe-go/v79/webhook"
	"gorm.io/gorm"
	"io"
	"net/http"
)

// InitiateStripeDonatorCheckoutSession Initiates a stripe checkout session for donator
// Endpoint: POST /v2/orders/stripe/initiate/donation
func InitiateStripeDonatorCheckoutSession(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	body := donationRequestBody{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	if body.Recurring && body.GiftUserId != user.Id {
		return APIErrorBadRequest("You cannot do recurring payments for gifted donator.")
	}

	price, err := getDonatorPrice(body.Months, false)

	if err != nil {
		return APIErrorBadRequest("You have provided an invalid amount of months.")
	}

	params := &stripe.CheckoutSessionParams{
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(getStripeDonatorPriceId(body.Months, body.Recurring)),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:         stripe.String(string(getStripeCheckoutMode(body.Recurring))),
		SuccessURL:   stripe.String(config.Instance.Stripe.DonateRedirectUrl),
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{Enabled: stripe.Bool(true)},
	}

	stripe.Key = config.Instance.Stripe.APIKey
	s, err := session.New(params)

	if err != nil {
		return APIErrorServerError("Error creating stripe checkout session", err)
	}

	order := &db.Order{
		UserId:        user.Id,
		OrderId:       -1,
		TransactionId: s.ID,
		IPAddress:     getIpFromRequest(c),
		ItemId:        db.OrderItemDonator,
		Quantity:      body.Months,
		Amount:        price,
		Description:   fmt.Sprintf("%v month(s) of Quaver Donator Perks (Stripe)", body.Months),
	}

	isSet, err := order.SetReceiver(user, body.GiftUserId)

	if err != nil {
		return APIErrorServerError("Error setting order receiver in db", err)
	}

	if !isSet {
		return APIErrorBadRequest("You are gifting donator to a user who doesn't exist.")
	}

	if err := order.Insert(); err != nil {
		return APIErrorServerError("Error inserting order into db", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Stripe checkout session successfully created.",
		"url":     s.URL,
	})

	return nil
}

// HandleStripeWebhook Handles an incoming stripe webhook
// POST: /v2/orders/stripe/webhook
func HandleStripeWebhook(c *gin.Context) *APIError {
	body, _ := io.ReadAll(c.Request.Body)

	event, err := webhook.ConstructEvent(body,
		c.Request.Header.Get("Stripe-Signature"), config.Instance.Stripe.WebhookSigningSecret)

	if err != nil {
		logrus.Error("Error verifying stripe webhook signature: ", err)
		return APIErrorUnauthorized("Error verifying webhook signature.")
	}

	switch event.Type {
	case "checkout.session.completed":
		if apiErr := FinalizeStripeOrder(&event); apiErr != nil {
			return apiErr
		}
		break
	case "invoice.paid":
		break
	default:
		break
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
	return nil
}

// FinalizeStripeOrder Handles the finalization of a Stripe order
func FinalizeStripeOrder(event *stripe.Event) *APIError {
	var stripeSession stripe.CheckoutSession

	if err := json.Unmarshal(event.Data.Raw, &stripeSession); err != nil {
		logrus.Error("Error parsing stripe webhook JSON", err)
		return APIErrorBadRequest("Error parsing stripe webhook JSON")
	}

	orders, err := db.GetStripeOrderById(stripeSession.ID)

	if err != nil {
		return APIErrorServerError("Error retrieving stripe order by id", err)
	}

	// Handle new incoming subscription
	var subscription *db.OrderSubscriptionStripe

	if stripeSession.Subscription != nil {
		existingSubscription, err := db.GetOrderSubscriptionById(stripeSession.Subscription.ID)

		if err != nil && err != gorm.ErrRecordNotFound {
			return APIErrorServerError("Error retrieving existing user stripe subscription", err)
		}

		if existingSubscription == nil {
			subscription = &db.OrderSubscriptionStripe{
				UserId:               orders[0].UserId,
				StripeCustomerId:     stripeSession.Customer.ID,
				StripeSubscriptionId: stripeSession.Subscription.ID,
			}

			if err := subscription.Insert(); err != nil {
				return APIErrorServerError("Error inserting new subscription", err)
			}
		}
	}

	for _, order := range orders {
		if subscription != nil {
			order.SubscriptionId = &subscription.Id
		}

		if err := order.Finalize(); err != nil {
			return APIErrorServerError("Error finalizing order", err)
		}
	}

	if err := webhooks.SendOrderWebhook(orders); err != nil {
		logrus.Error("Error sending order webhook: ", err)
	}

	return nil
}

// Creates a new stripe checkout session with a list of orders
func createStripeCheckoutSession(c *gin.Context, orders []*db.Order) *APIError {
	var lineItems []*stripe.CheckoutSessionLineItemParams

	for _, order := range orders {
		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			Price:    stripe.String(order.Item.StripePriceId),
			Quantity: stripe.Int64(int64(order.Quantity)),
		})
	}

	params := &stripe.CheckoutSessionParams{
		LineItems:    lineItems,
		Mode:         stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:   stripe.String(config.Instance.Stripe.StorePaymentRedirectUrl),
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{Enabled: stripe.Bool(true)},
	}

	stripe.Key = config.Instance.Stripe.APIKey
	s, err := session.New(params)

	if err != nil {
		return APIErrorServerError("Error creating stripe checkout session", err)
	}

	// Make sure all properties are set and insert to db
	for _, order := range orders {
		order.OrderId = -1
		order.TransactionId = s.ID
		order.Amount = float32(order.Item.PriceStripe) / 100

		if err := order.Insert(); err != nil {
			return APIErrorServerError("Error inserting order into db", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Stripe checkout session successfully created.",
		"url":     s.URL,
	})

	return nil
}

// Gets the donator price id for Stripe
func getStripeDonatorPriceId(months int, isRecurring bool) string {
	if config.Instance.IsProduction {
		if isRecurring {
			switch months {
			case 1:
				return "price_1PVzPCRw21C92BKs2QSHNO7r"
			case 3:
				return "price_1PVzPCRw21C92BKsKAP3P5q1"
			case 6:
				return "price_1PVzPCRw21C92BKsTBr4KH42"
			case 12:
				return "price_1PVzPCRw21C92BKsHmJrMTzV"
			}
		} else {
			switch months {
			case 1:
				return "price_1PVzPCRw21C92BKs2QSHNO7r"
			case 3:
				return "price_1PVzPCRw21C92BKshby0SSMt"
			case 6:
				return "price_1PVzPCRw21C92BKsQPIwwwXE"
			case 12:
				return "price_1PVzPCRw21C92BKsk16cVikC"
			}
		}
		// Debug
	} else {
		if isRecurring {
			switch months {
			case 1:
				return "price_1PVwIaRw21C92BKsG9v8AbPs"
			case 3:
				return "price_1PVwJfRw21C92BKsYDQVjBAD"
			case 6:
				return "price_1PVwKwRw21C92BKsIiU4zoDf"
			case 12:
				return "price_1PVwLRRw21C92BKskJdQafnA"
			}
		} else {
			switch months {
			case 1:
				return "price_1PVw9yRw21C92BKstVWQy0zJ"
			case 3:
				return "price_1PVwJSRw21C92BKs6DbA6L1n"
			case 6:
				return "price_1PVwKkRw21C92BKs3VhkLSMJ"
			case 12:
				return "price_1PVwLGRw21C92BKsNHpvLuTb"
			}
		}
	}

	return ""
}

// Returns the stripe checkout session mode depending on if the payment is recurring or not.
func getStripeCheckoutMode(isRecurring bool) stripe.CheckoutSessionMode {
	if isRecurring {
		return stripe.CheckoutSessionModeSubscription
	}

	return stripe.CheckoutSessionModePayment
}
