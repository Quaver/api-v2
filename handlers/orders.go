package handlers

import (
	"errors"
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

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

// Common request body when initiating a donator
type donationRequestBody struct {
	Months     int  `form:"months" json:"months" binding:"required"`
	GiftUserId int  `form:"gift_user_id" json:"gift_user_id"`
	Recurring  bool `form:"recurring" json:"recurring"`
}

// Gets the receiver of the order and makes sure donationRequestBody.GiftUserId is set properly
func (body *donationRequestBody) getOrderReceiver(user *db.User) (*db.User, *APIError) {
	// User is gifting to someone else
	if body.GiftUserId != 0 {
		receiver, err := db.GetUserById(body.GiftUserId)

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, APIErrorBadRequest("The user you are trying to gift to doesn't exist.")
			}

			return nil, APIErrorServerError("Error retrieving donator gift user id", err)
		}

		return receiver, nil
	}

	// User is purchasing for themselves.
	body.GiftUserId = user.Id
	return user, nil
}
