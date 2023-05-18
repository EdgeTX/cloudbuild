package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewErrorResponse(message string) *ErrorResponse {
	return &ErrorResponse{
		Error: message,
	}
}

func errorResponse(c *gin.Context, statusCode int, err error) {
	c.AbortWithStatusJSON(
		statusCode,
		NewErrorResponse(err.Error()),
	)
}

func BadRequestResponse(c *gin.Context, err error) {
	errorResponse(c, http.StatusBadRequest, err)
}

func ServiceUnavailableResponse(c *gin.Context, err error) {
	errorResponse(c, http.StatusServiceUnavailable, err)
}

func UnprocessableEntityResponse(c *gin.Context, message string) {
	c.AbortWithStatusJSON(
		http.StatusUnprocessableEntity,
		NewErrorResponse(message),
	)
}
