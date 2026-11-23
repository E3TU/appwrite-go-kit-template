package handlers

import (
	"net/http"

	"github.com/e3tu/appwrite-base/auth"
	"github.com/gin-gonic/gin"
)

// Me returns the authenticated user. It must be mounted behind the auth
// middleware; reaching it without a session is a routing bug, so it answers
// 401 rather than panicking.
func (h *Handlers) Me(c *gin.Context) {
	u, ok := auth.UserFrom(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    u.ID,
		"email": u.Email,
		"name":  u.Name,
	})
}
