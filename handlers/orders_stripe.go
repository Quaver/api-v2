package handlers

import (
	"fmt"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/checkout/session"
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

	if body.Recurring {
		return APIErrorBadRequest("Recurring payments are not available.")
	}

	if body.Recurring && body.GiftUserId == user.Id {
		return APIErrorBadRequest("You cannot do recurring payments for gifted donator.")
	}

	orderReceiver, apiErr := body.getOrderReceiver(user)

	if apiErr != nil {
		return apiErr
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
		SuccessURL:   stripe.String("https://quavergame.com/donate?status=success"),
		CancelURL:    stripe.String("https://quavergame.com/donate?status=cancelled"),
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{Enabled: stripe.Bool(true)},
	}

	stripe.Key = config.Instance.Stripe.APIKey
	s, err := session.New(params)

	if err != nil {
		return APIErrorServerError("Error creating stripe checkout session", err)
	}

	orders := []*db.Order{
		{
			UserId:         user.Id,
			OrderId:        -1,
			TransactionId:  s.ID,
			IPAddress:      getIpFromRequest(c),
			ItemId:         db.OrderItemDonator,
			Quantity:       body.Months,
			Amount:         price,
			Description:    fmt.Sprintf("%v month(s) of Quaver Donator Perks for %v (Stripe)", body.Months, orderReceiver.Username),
			ReceiverUserId: body.GiftUserId,
			Receiver:       orderReceiver,
		},
	}

	for _, order := range orders {
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

	return ""
}

// Returns the stripe checkout session mode depending on if the payment is recurring or not.
func getStripeCheckoutMode(isRecurring bool) stripe.CheckoutSessionMode {
	if isRecurring {
		return stripe.CheckoutSessionModeSubscription
	}

	return stripe.CheckoutSessionModePayment
}
