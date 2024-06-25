package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/webhooks"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"math/rand"
	"net/http"
	"slices"
	"strconv"
	"time"
)

// InitiateSteamDonatorTransaction Initiates a transaction for Steam donator
// Endpoint: POST /v2/orders/steam/initiate/donation
func InitiateSteamDonatorTransaction(c *gin.Context) *APIError {
	user := getAuthedUser(c)

	if user == nil {
		return nil
	}

	body := struct {
		Months       int            `form:"months" json:"months" binding:"required"`
		GiftUserId   int            `form:"gift_user_id" json:"gift_user_id"`
		BundleItemId db.OrderItemId `form:"bundle_item_id" json:"bundle_item_id"`
	}{}

	if err := c.ShouldBind(&body); err != nil {
		return APIErrorBadRequest("Invalid request body")
	}

	var orderReceiver *db.User

	// Set order receiver + check if user exists
	if body.GiftUserId != 0 {
		receiver, err := db.GetUserById(body.GiftUserId)

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return APIErrorBadRequest("The user you are trying to gift to doesn't exist.")
			}

			return APIErrorServerError("Error retrieving donator gift user id", err)
		}

		orderReceiver = receiver
	} else {
		body.GiftUserId = user.Id
		orderReceiver = user
	}

	price, err := getDonatorPrice(body.Months, true)

	if err != nil {
		return APIErrorBadRequest("You have provided an invalid amount of months.")
	}

	orders := []*db.Order{
		{
			UserId:         user.Id,
			SteamOrderId:   generateSteamOrderId(),
			IPAddress:      getSteamTransactionIp(c),
			ItemId:         db.OrderItemDonator,
			Quantity:       body.Months,
			Amount:         price,
			Description:    fmt.Sprintf("%v month(s) of Quaver Donator Perks for %v", body.Months, orderReceiver.Username),
			ReceiverUserId: body.GiftUserId,
			Receiver:       orderReceiver,
		},
	}

	// Add bundled item to the order
	if body.BundleItemId > 0 {
		item, err := db.GetOrderItemById(int(body.BundleItemId))

		if err != nil && err != gorm.ErrRecordNotFound {
			return APIErrorServerError("Error retrieving order item from db", err)
		}

		if item == nil {
			return APIErrorBadRequest("You provided a bundle item that does not exist.")
		}

		if !item.DonatorBundleItem {
			return APIErrorBadRequest("You provided an item that cannot be bundled with donator.")
		}

		if !item.InStock {
			return APIErrorBadRequest("The bundle item you have provided is no longer in stock.")
		}

		if !item.CanGift {
			return APIErrorBadRequest("This bundle item you have provided cannot be gifted.")
		}

		orders = append(orders, &db.Order{
			UserId:         user.Id,
			SteamOrderId:   orders[0].SteamOrderId,
			IPAddress:      getSteamTransactionIp(c),
			ItemId:         db.OrderItemId(item.Id),
			Quantity:       1,
			Amount:         float32(item.PriceSteam) / 100,
			Description:    fmt.Sprintf("%v for %v", item.Name, orderReceiver.Username),
			ReceiverUserId: body.GiftUserId,
			Receiver:       orderReceiver,
		})
	}

	parsed, apiErr := steamInitTransaction(user, orders)

	if apiErr != nil {
		return apiErr
	}

	for _, order := range orders {
		if err := order.Insert(); err != nil {
			return APIErrorServerError("Error saving order to db", err)
		}
	}

	returnUrl := fmt.Sprintf("%v/v2/orders/steam/finalize?order_id=%v%%26transaction_id=%v",
		config.Instance.APIUrl, orders[0].SteamOrderId, orders[0].SteamTransactionId)

	c.JSON(http.StatusOK, gin.H{
		"message":   "The transaction has been successfully initiated.",
		"steam_url": fmt.Sprintf("%v?returnurl=%v", parsed.Response.Params.SteamURL, returnUrl),
	})
	return nil
}

// FinalizeSteamTransaction Finalizes a Steam transaction
// Endpoint: GET /v2/orders/steam/finalize?order_id=&transaction_id=
func FinalizeSteamTransaction(c *gin.Context) *APIError {
	orderId := c.Query("order_id")
	transactionId := c.Query("transaction_id")

	if orderId == "" {
		return APIErrorBadRequest("You must provide a valid orderid query parameter.")
	}

	if transactionId == "" {
		return APIErrorBadRequest("You must provide a valid transactionid query parameter.")
	}

	orders, err := db.GetSteamOrdersByIds(orderId, transactionId)

	if err != nil {
		return APIErrorServerError("Error retrieving Steam orders by id in db", err)
	}

	if len(orders) == 0 {
		return APIErrorNotFound("Order")
	}

	if slices.ContainsFunc(orders, func(order *db.Order) bool {
		return order.Status == "Completed"
	}) {
		return APIErrorForbidden("This order was already completed.")
	}

	if _, apiErr := steamQueryTransaction(orderId, transactionId); apiErr != nil {
		return apiErr
	}

	if _, apiErr := steamFinalizeTransaction(orderId); apiErr != nil {
		return apiErr
	}

	for _, order := range orders {
		if err := order.Finalize(); err != nil {
			return APIErrorServerError("Error finalizing Steam order", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "The transaction has been successfully completed",
		"orders":  orders,
	})

	_ = webhooks.SendOrderWebhook(orders)
	return nil
}

// Generates a random 8 digit steam order id
func generateSteamOrderId() int {
	minimum := 10000000
	maximum := 99999999

	return rand.Intn(maximum-minimum+1) + minimum
}

// Returns the transaction ip address for steam
func getSteamTransactionIp(c *gin.Context) string {
	if config.Instance.IsProduction {
		return getIpFromRequest(c)
	}

	return "1.1.1.1"
}

// Response from Steam InitTxn API Call
// Endpoint: getSteamInitTransactionUrl()
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

func steamInitTransaction(user *db.User, orders []*db.Order) (*steamInitTxnResponse, *APIError) {
	var endpoint string

	if config.Instance.IsProduction {
		endpoint = "https://partner.steam-api.com/ISteamMicroTxn/InitTxn/v3/"
	} else {
		endpoint = "https://partner.steam-api.com/ISteamMicroTxnSandbox/InitTxn/v3"
	}

	data := map[string]string{
		"appid":       fmt.Sprintf("%v", config.Instance.Steam.AppId),
		"key":         config.Instance.Steam.PublisherKey,
		"steamid":     user.SteamId,
		"orderid":     fmt.Sprintf("%v", generateSteamOrderId()),
		"usersession": "web",
		"ipaddress":   orders[0].IPAddress,
		"language":    "en",
		"currency":    "USD",
		"itemcount":   fmt.Sprintf("%v", len(orders)),
	}

	// Add all items to the order
	for index, order := range orders {
		data[fmt.Sprintf("itemid[%v]", index)] = fmt.Sprintf("%v", order.ItemId)
		data[fmt.Sprintf("amount[%v]", index)] = fmt.Sprintf("%v", order.Amount*100)
		data[fmt.Sprintf("description[%v]", index)] = fmt.Sprintf("%v", order.Description)

		quantity := order.Quantity

		// Donator always has a quantity of 1.
		if order.ItemId == db.OrderItemDonator {
			quantity = 1
		}

		data[fmt.Sprintf("qty[%v]", index)] = fmt.Sprintf("%v", quantity)
	}

	resp, err := resty.New().R().
		SetFormData(data).
		Post(endpoint)

	if err != nil || resp.IsError() {
		logrus.Errorf("Steam InitTxn failed w/ error: %v - %v", resp.StatusCode(), string(resp.Body()))
		return nil, APIErrorServerError("Steam InitTxn Failed", err)
	}

	var parsed steamInitTxnResponse

	if err := json.Unmarshal(resp.Body(), &parsed); err != nil {
		return nil, APIErrorServerError("Error parsing Steam InitTxn response", err)
	}

	if parsed.Response.Result != "OK" {
		logrus.Errorf("Steam InitTxn failed w/ error: %v - %v", resp.StatusCode(), string(resp.Body()))
		return nil, APIErrorServerError("Steam InitTxn Failed", errors.New("result not OK"))
	}

	for _, order := range orders {
		order.SteamOrderId, _ = strconv.Atoi(parsed.Response.Params.OrderId)
		order.SteamTransactionId = parsed.Response.Params.TransactionId
	}

	return &parsed, nil
}

// Response from Steam QueryTxn API Call
// Endpoint: getSteamQueryTransactionUrl()
type steamQueryTxnResponse struct {
	Response struct {
		Result string `json:"result,omitempty"`
		Params struct {
			OrderId     string    `json:"orderid,omitempty"`
			TransId     string    `json:"transid,omitempty"`
			SteamId     string    `json:"steamid,omitempty"`
			Status      string    `json:"status,omitempty"`
			Currency    string    `json:"currency,omitempty"`
			Time        time.Time `json:"time,omitempty"`
			Country     string    `json:"country,omitempty"`
			USState     string    `json:"usstate,omitempty"`
			TimeCreated time.Time `json:"timecreated,omitempty"`
			Items       []struct {
				ItemId     int    `json:"itemid,omitempty"`
				Qty        int    `json:"qty,omitempty"`
				Amount     int    `json:"amount,omitempty"`
				Vat        int    `json:"vat,omitempty"`
				ItemStatus string `json:"itemstatus,omitempty"`
			} `json:"items,omitempty"`
		} `json:"params,omitempty"`
	} `json:"response"`
}

// Requests the endpoint to query a steam transaction
func steamQueryTransaction(orderId string, transactionId string) (*steamQueryTxnResponse, *APIError) {
	var endpoint string

	if config.Instance.IsProduction {
		endpoint = "https://partner.steam-api.com/ISteamMicroTxn/QueryTxn/v3/"
	} else {
		endpoint = "https://partner.steam-api.com/ISteamMicroTxnSandbox/QueryTxn/v3/"
	}

	resp, err := resty.New().R().
		SetQueryParams(map[string]string{
			"key":     config.Instance.Steam.PublisherKey,
			"appid":   fmt.Sprintf("%v", config.Instance.Steam.AppId),
			"orderid": orderId,
			"transid": transactionId,
		}).
		Get(endpoint)

	if err != nil || resp.IsError() {
		logrus.Errorf("Steam QueryTxn failed w/ error: %v - %v", resp.StatusCode(), string(resp.Body()))
		return nil, APIErrorServerError("Steam QueryTxn Failed", err)
	}

	var parsed steamQueryTxnResponse

	if err := json.Unmarshal(resp.Body(), &parsed); err != nil {
		return nil, APIErrorServerError("Error parsing Steam QueryTxn response", err)
	}

	if parsed.Response.Result != "OK" {
		logrus.Errorf("Steam QueryTxn failed w/ error: %v - %v", resp.StatusCode(), string(resp.Body()))
		return nil, APIErrorServerError("Steam QueryTxn Failed", errors.New("result not OK"))
	}

	if parsed.Response.Params.Status != "Approved" {
		return nil, APIErrorBadRequest("The transaction was not approved on Steam.")
	}

	return &parsed, nil
}

type steamFinalizeTxnResponse struct {
	Response struct {
		Result string `json:"result"`
		Params struct {
			Orderid string `json:"orderid,omitempty"`
			Transid string `json:"transid,omitempty"`
		} `json:"params,omitempty"`
	} `json:"response"`
}

// Requests the Steam endpoint to finalize a transaction
func steamFinalizeTransaction(orderId string) (*steamFinalizeTxnResponse, *APIError) {
	var endpoint string

	if config.Instance.IsProduction {
		endpoint = "https://partner.steam-api.com/ISteamMicroTxn/FinalizeTxn/v2/"
	} else {
		endpoint = "https://partner.steam-api.com/ISteamMicroTxnSandbox/FinalizeTxn/v2/"
	}

	resp, err := resty.New().R().
		SetFormData(map[string]string{
			"appid":   fmt.Sprintf("%v", config.Instance.Steam.AppId),
			"key":     config.Instance.Steam.PublisherKey,
			"orderid": orderId,
		}).
		Post(endpoint)

	if err != nil || resp.IsError() {
		logrus.Errorf("Steam FinalizeTxn failed w/ error: %v - %v", resp.StatusCode(), string(resp.Body()))
		return nil, APIErrorServerError("Steam FinalizeTxn Failed", err)
	}

	var parsed steamFinalizeTxnResponse

	if err := json.Unmarshal(resp.Body(), &parsed); err != nil {
		return nil, APIErrorServerError("Error parsing Steam FinalizeTxn response", err)
	}

	if parsed.Response.Result != "OK" {
		logrus.Errorf("Steam FinalizeTxn failed w/ error: %v - %v", resp.StatusCode(), string(resp.Body()))
		return nil, APIErrorServerError("Steam FinalizeTx Failed", errors.New("result not OK"))
	}

	return &parsed, nil
}
