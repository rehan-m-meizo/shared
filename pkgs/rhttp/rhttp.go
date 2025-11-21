// Package res provides simple, consistent JSON responses for Gin handlers.
//
// It standardizes success and error responses with minimal boilerplate.
package rhttp

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response defines the standard JSON structure.
type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// send is the internal unified response builder.
func send(c *gin.Context, status int, message string, data interface{}, err interface{}) {
	r := &Response{
		Status:  status,
		Message: message,
		Data:    data,
		Error:   err,
	}
	c.AbortWithStatusJSON(status, r)
}

// --- Success helpers ---

func OK(c *gin.Context, data interface{}) {
	send(c, http.StatusOK, "success", data, nil)
}

func Created(c *gin.Context, data interface{}) {
	send(c, http.StatusCreated, "created", data, nil)
}

// --- Error helpers ---

func Error(c *gin.Context, status int, message string, err error) {
	errDetail := map[string]string{}
	if err != nil {
		errDetail["message"] = err.Error()
	}
	send(c, status, message, nil, errDetail)
}

func BadRequest(c *gin.Context, err error) {
	Error(c, http.StatusBadRequest, "bad requests", err)
}

func Unauthorized(c *gin.Context, err error) {
	Error(c, http.StatusUnauthorized, "unauthorized", err)
}

func Forbidden(c *gin.Context, err error) {
	Error(c, http.StatusForbidden, "forbidden", err)
}

func NotFound(c *gin.Context, err error) {
	Error(c, http.StatusNotFound, "not found", err)
}

func Conflict(c *gin.Context, err error) {
	Error(c, http.StatusConflict, "conflict", err)
}

func Unprocessable(c *gin.Context, err error) {
	Error(c, http.StatusUnprocessableEntity, "unprocessable entity", err)
}

func Internal(c *gin.Context, err error) {
	Error(c, http.StatusInternalServerError, "internal server error", err)
}

func ServiceUnavailable(c *gin.Context, err error) {
	Error(c, http.StatusServiceUnavailable, "service unavailable", err)
}

func GatewayTimeout(c *gin.Context, err error) {
	Error(c, http.StatusGatewayTimeout, "gateway timeout", err)
}

func TooManyRequests(c *gin.Context, err error) {
	Error(c, http.StatusTooManyRequests, "too many requests", err)
}

func NotImplemented(c *gin.Context, message string) {
	Error(c, http.StatusNotImplemented, "not implemented", errors.New(message))
}
