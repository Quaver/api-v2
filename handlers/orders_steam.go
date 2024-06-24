package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"math/rand"
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

// InitiateSteamDonatorTransaction Initiates a transaction for Steam donator
// Endpoint: POST /v2/orders/donations/steam/initiate
func InitiateSteamDonatorTransaction(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	body := struct {
		Months     int `form:"months" json:"months" binding:"required"`
		GiftUserId int `form:"gift_user_id" json:"gift_user_id"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	// Check to see if the gift user exists
	if body.GiftUserId != 0 {
		if _, err := db.GetUserById(body.GiftUserId); err != nil {
			if err == gorm.ErrRecordNotFound {
				return APIErrorBadRequest("The user you are trying to gift to doesn't exist.")
			}

			return APIErrorServerError("Error retrieving donator gift user id", err)
		}
	} else {
		body.GiftUserId = user.Id
	}

	price, err := getDonatorPrice(body.Months, true)

	if err != nil {
		return APIErrorBadRequest("You have provided an invalid amount of months.")
	}

	order := &db.Order{
		UserId:         user.Id,
		SteamOrderId:   generateSteamOrderId(),
		IPAddress:      getSteamTransactionIp(c),
		ItemId:         1,
		Quantity:       1,
		Amount:         price,
		Description:    fmt.Sprintf("%v month(s) of Quaver Donator Perks", body.Months),
		ReceiverUserId: body.GiftUserId,
		Status:         db.OrderStatusWaiting,
	}

	resp, err := resty.New().R().
		SetFormData(map[string]string{
			"appid":          fmt.Sprintf("%v", config.Instance.Steam.AppId),
			"key":            config.Instance.Steam.PublisherKey,
			"steamid":        user.SteamId,
			"orderid":        fmt.Sprintf("%v", generateSteamOrderId()),
			"usersession":    "web",
			"ipaddress":      order.IPAddress,
			"language":       "en",
			"currency":       "USD",
			"itemcount":      "1",
			"itemid[0]":      "1",
			"qty[0]":         "1",
			"amount[0]":      fmt.Sprintf("%v", order.Amount*100),
			"description[0]": order.Description,
		}).
		Post(getSteamTransactionUrl())

	if err != nil || resp.IsError() {
		logrus.Errorf("Steam InitTxn failed w/ error: %v - %v", resp.StatusCode(), string(resp.Body()))
		return APIErrorServerError("Steam InitTxn Failed", err)
	}

	var parsed steamInitTxnResponse

	if err := json.Unmarshal(resp.Body(), &parsed); err != nil {
		return APIErrorServerError("Error parsing Steam InitTxn response", err)
	}

	if parsed.Response.Result != "OK" {
		logrus.Errorf("Steam InitTxn failed w/ error: %v - %v", resp.StatusCode(), string(resp.Body()))
		return APIErrorServerError("Steam InitTxn Failed", errors.New("result not OK"))
	}

	order.SteamTransactionId = parsed.Response.Params.TransactionId

	if err := order.Insert(); err != nil {
		return APIErrorServerError("Error saving order to db", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "The transaction has been successfully initiated.",
		"steam_url": parsed.Response.Params.SteamURL,
	})

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

// Generates a random 8 digit steam order id
func generateSteamOrderId() int {
	minimum := 10000000
	maximum := 99999999

	return rand.Intn(maximum-minimum+1) + minimum
}

func getSteamTransactionUrl() string {
	if config.Instance.IsProduction {
		return "https://partner.steam-api.com/ISteamMicroTxn/InitTxn/v3/"
	} else {
		return "https://partner.steam-api.com/ISteamMicroTxnSandbox/InitTxn/v3"
	}
}

// Returns the transaction ip address for steam
func getSteamTransactionIp(c *gin.Context) string {
	if config.Instance.IsProduction {
		return getIpFromRequest(c)
	}

	return "1.1.1.1"
}

// Response from Steam InitTxn API Call
// Endpoint: https://partner.steam-api.com/ISteamMicroTxn/InitTxn/v3/
type steamInitTxnResponse struct {
	Response struct {
		Result string `json:"result,omitempty"`

		Params struct {
			OrderId       string `json:"orderid,omitempty"`
			TransactionId string `json:"transid,omitempty"`
			SteamURL      string `json:"steamurl,omitempty"`
		} `json:"params"`

		Error interface{} `json:"error,omitempty"`
	} `json:"response"`
}
