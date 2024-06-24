package handlers

import (
	"errors"
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
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
