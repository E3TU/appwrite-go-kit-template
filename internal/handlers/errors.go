package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/appwrite/sdk-for-go/v7/query"
	"github.com/e3tu/appwrite-base/auth"
	"github.com/gin-gonic/gin"
)

// userFromContext adapts auth.UserFrom to the handler package.
func userFromContext(c *gin.Context) (*auth.User, bool) {
	return auth.UserFrom(c)
}

// writeSDKError maps an SDK error to an HTTP status.
//
// The Appwrite SDK reports the upstream status through GetStatusCode; anything
// else becomes 502 because the failure is ours or the deployment's, not the
// caller's.
func writeSDKError(c *gin.Context, err error) {
	status := http.StatusBadGateway

	var sdkErr interface{ GetStatusCode() int }
	if errors.As(err, &sdkErr) {
		switch code := sdkErr.GetStatusCode(); {
		case code == http.StatusNotFound:
			status = http.StatusNotFound
		case code == http.StatusUnauthorized, code == http.StatusForbidden:
			status = http.StatusForbidden
		case code >= 400 && code < 500:
			status = http.StatusBadRequest
		}
	}

	c.JSON(status, gin.H{"error": err.Error()})
}

// queryLimit builds the limit clause, clamped to the Appwrite maximum.
func queryLimit(raw string) (string, bool) {
	if raw == "" {
		return "", false
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return "", false
	}
	if n > 100 {
		n = 100
	}
	return query.Limit(n), true
}
