// Package httpresponses is deprecated and will be removed in the future.
// Use ginres instead.
//
// Deprecated: Use ginres instead.
package httpresponses

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Error struct {
	ErrorCode int                    `json:"error_code"`
	Errors    map[string]interface{} `json:"errors"`
}

type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func OK(c *gin.Context, data interface{}) {
	r := &Response{
		Status:  http.StatusOK,
		Message: "success",
		Data:    data,
	}
	c.JSON(http.StatusOK, r)
}

func Created(c *gin.Context, data interface{}) {
	r := &Response{
		Status:  http.StatusCreated,
		Message: "created",
		Data:    data,
	}
	c.JSON(http.StatusCreated, r)
}

func UnProcessableEntity(c *gin.Context, err error) {
	e := Error{}

	e.ErrorCode = http.StatusUnprocessableEntity
	e.Errors = make(map[string]interface{})
	switch v := err.(type) {
	default:
		e.Errors["error"] = v.Error()
	}

	c.AbortWithStatusJSON(http.StatusUnprocessableEntity, e)
}

func BadRequest(c *gin.Context, err error) {

	e := Error{}

	e.ErrorCode = http.StatusBadRequest
	e.Errors = make(map[string]interface{})
	switch v := err.(type) {
	default:
		e.Errors["error"] = v.Error()
	}

	c.AbortWithStatusJSON(http.StatusBadRequest, e)
}

func InternalServerError(c *gin.Context, err error) {
	e := Error{}

	e.ErrorCode = http.StatusInternalServerError
	e.Errors = make(map[string]interface{})
	switch v := err.(type) {
	default:
		e.Errors["error"] = v.Error()
	}

	c.AbortWithStatusJSON(http.StatusInternalServerError, e)
}

func NotFound(c *gin.Context, err error) {
	e := Error{}

	e.ErrorCode = http.StatusNotFound
	e.Errors = make(map[string]interface{})
	switch v := err.(type) {
	default:
		e.Errors["error"] = v.Error()
	}

	c.AbortWithStatusJSON(http.StatusNotFound, e)
}

func Unauthorized(c *gin.Context, err error) {
	e := Error{}

	e.ErrorCode = http.StatusUnauthorized
	e.Errors = make(map[string]interface{})
	switch v := err.(type) {
	default:
		e.Errors["error"] = v.Error()
	}

	c.AbortWithStatusJSON(http.StatusUnauthorized, e)
}

func Forbidden(c *gin.Context, err error) {
	e := Error{}

	e.ErrorCode = http.StatusForbidden
	e.Errors = make(map[string]interface{})
	switch v := err.(type) {
	default:
		e.Errors["error"] = v.Error()
	}

	c.AbortWithStatusJSON(http.StatusForbidden, e)
}

func Conflict(c *gin.Context, err error) {
	e := Error{}

	e.ErrorCode = http.StatusConflict
	e.Errors = make(map[string]interface{})
	switch v := err.(type) {
	default:
		e.Errors["error"] = v.Error()
	}

	c.AbortWithStatusJSON(http.StatusConflict, e)
}

func NotImplemented(c *gin.Context, err error) {
	e := Error{}

	e.ErrorCode = http.StatusNotImplemented
	e.Errors = make(map[string]interface{})
	switch v := err.(type) {
	default:
		e.Errors["error"] = v.Error()
	}

	c.AbortWithStatusJSON(http.StatusNotImplemented, e)
}

func ServiceUnavailable(c *gin.Context, err error) {
	e := Error{}

	e.ErrorCode = http.StatusServiceUnavailable
	e.Errors = make(map[string]interface{})
	switch v := err.(type) {
	default:
		e.Errors["error"] = v.Error()
	}

	c.AbortWithStatusJSON(http.StatusServiceUnavailable, e)
}

func GatewayTimeout(c *gin.Context, err error) {
	e := Error{}

	e.ErrorCode = http.StatusGatewayTimeout
	e.Errors = make(map[string]interface{})
	switch v := err.(type) {
	default:
		e.Errors["error"] = v.Error()
	}

	c.AbortWithStatusJSON(http.StatusGatewayTimeout, e)
}

func TooManyRequests(c *gin.Context, err error) {
	e := Error{}

	e.ErrorCode = http.StatusTooManyRequests
	e.Errors = make(map[string]interface{})
	switch v := err.(type) {
	default:
		e.Errors["error"] = v.Error()
	}

	c.AbortWithStatusJSON(http.StatusTooManyRequests, e)
}
