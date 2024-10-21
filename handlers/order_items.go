package handlers

import (
	"github.com/Quaver/api2/db"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

// GetOrderItemById \Retrieves an individual order item by its id
// Endpoint: GET /v2/items/:id
func GetOrderItemById(c *gin.Context) *APIError {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return APIErrorBadRequest("Invalid id")
	}

	item, err := db.GetOrderItemById(db.OrderItemId(id))

	if err != nil && err != gorm.ErrRecordNotFound {
		return APIErrorServerError("Error retrieving order item from db", err)
	}

	if item == nil {
		return APIErrorNotFound("Order item")
	}

	c.JSON(http.StatusOK, gin.H{"order_item": item})
	return nil
}
