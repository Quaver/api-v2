package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

type APIError struct {
	Status  int    `json:"status"`
	Message string `json:"error"`
	Error   error
}

func APIErrorBadRequest(message string) *APIError {
	return &APIError{Status: http.StatusBadRequest, Message: message}
}

func APIErrorServerError(message string, err error) *APIError {
	return &APIError{Status: http.StatusInternalServerError, Message: message, Error: err}
}

func APIErrorNotFound(message string) *APIError {
	return &APIError{Status: http.StatusNotFound, Message: fmt.Sprintf("`%v` not found", message)}
}

// Handles logging and responding to the user
func handleAPIError(c *gin.Context, err *APIError) {
	if err.Error != nil {
		logrus.Errorf("%v - %v", err.Message, err.Error)
	}

	if err.Status == http.StatusInternalServerError {
		ReturnError(c, err.Status, "Internal Server Error")
		return
	}

	ReturnError(c, err.Status, err.Message)
}

func ReturnError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"error": message,
	})
}
