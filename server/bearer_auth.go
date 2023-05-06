package server

import (
	"errors"
	"net/http"
	"strings"

	"github.com/edgetx/cloudbuild/auth"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var (
	ErrBadBearerToken = errors.New("missing or incorrectly formatted bearer token")
)

func splitAuthToken(token string) (string, string, error) {
	parts := strings.Split(token, "-")
	if len(parts) != 2 {
		return "", "", ErrBadBearerToken
	}
	accessKey := parts[0]
	secretKey := parts[1]
	return accessKey, secretKey, nil
}

func extractBearerToken(header string) (string, string, error) {
	if header == "" {
		return "", "", ErrBadBearerToken
	}

	bearerToken := strings.Split(header, "Bearer ")
	if len(bearerToken) != 2 {
		return "", "", ErrBadBearerToken
	}

	return splitAuthToken(bearerToken[1])
}

func BearerAuth(auth *auth.AuthTokenDB, handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHdr := c.GetHeader("Authorization")
		if authHdr == "" {
			c.AbortWithStatus(
				http.StatusUnauthorized,
			)
			return
		}
		accessKey, secretKey, err := extractBearerToken(authHdr)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				NewErrorResponse(err.Error()),
			)
			return
		}
		log.Debugln("AccessKey:", accessKey, "SecretKey:", secretKey)
		err = auth.Authenticate(accessKey, secretKey)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				NewErrorResponse(err.Error()),
			)
			return
		}
		handler(c)
	}
}
