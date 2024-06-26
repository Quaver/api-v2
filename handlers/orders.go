package handlers

import (
	"errors"
	"fmt"
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

type checkoutPaymentMethod int

const (
	paymentMethodSteam checkoutPaymentMethod = iota
	paymentMethodStripe
)

// Common request body when initiating a donator transaction
type donationRequestBody struct {
	Months     int  `form:"months" json:"months" binding:"required"`
	GiftUserId int  `form:"gift_user_id" json:"gift_user_id"`
	Recurring  bool `form:"recurring" json:"recurring"`
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
			Id         db.OrderItemId `json:"id"`
			Quantity   int            `json:"quantity"`
			GiftUserId int            `json:"gift_user_id"`
		} `json:"line_items"`
	}{}

	if err := c.ShouldBindJSON(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	var orders []*db.Order

	for _, lineItem := range body.LineItems {
		if lineItem.Id == db.OrderItemDonator {
			return APIErrorBadRequest("You cannot purchase donator through this endpoint.")
		}

		orderItem, err := db.GetOrderItemById(int(lineItem.Id))

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
			UserId:      user.Id,
			IPAddress:   getSteamTransactionIp(c),
			ItemId:      lineItem.Id,
			Quantity:    lineItem.Quantity,
			Description: orderItem.Name,
			Item:        orderItem,
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
	}

	switch body.PaymentMethod {
	case paymentMethodSteam:
		return createSteamCheckoutSession(c, user, orders)
	case paymentMethodStripe:
		return createStripeCheckoutSession(c, orders)
	}

	return APIErrorBadRequest("Invalid payment method provided")
}

// Gets the donator price.
// Steam price = OG Price + 30%
func getDonatorPrice(months int, isSteam bool) (float32, error) {
	switch months {
	case 1:
		if isSteam {
			return 6.49, nil
		}

		return 4.99, nil
	case 3:
		if isSteam {
			return 17.99, nil
		}

		return 13.99, nil
	case 6:
		if isSteam {
			return 34.99, nil
		}

		return 26.99, nil
	case 12:
		if isSteam {
			return 64.99, nil
		}

		return 49.99, nil
	}

	return 0, errors.New("invalid donator months provided")
}
