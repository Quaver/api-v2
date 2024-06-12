package handlers

import (
	"fmt"
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

func APIErrorForbidden(message string) *APIError {
	return &APIError{Status: http.StatusForbidden, Message: message}
}

func APIErrorUnauthorized(message string) *APIError {
	return &APIError{Status: http.StatusUnauthorized, Message: message}
}

func APIErrorNotFound(message string) *APIError {
	return &APIError{Status: http.StatusNotFound, Message: fmt.Sprintf("%v not found", message)}
}

func APIErrorServerError(message string, err error) *APIError {
	return &APIError{Status: http.StatusInternalServerError, Message: message, Error: err}
}
