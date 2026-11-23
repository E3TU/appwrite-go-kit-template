// Package handlers contains the template's HTTP handlers.
//
// Handlers talk to the Appwrite SDK through the library's services. They own
// no Appwrite logic beyond mapping rows to JSON and HTTP status codes.
package handlers

import (
	"net/http"

	"github.com/e3tu/appwrite-base"
	"github.com/gin-gonic/gin"
)

// Handlers holds what every handler needs.
type Handlers struct {
	Base  *base.Base
	Todos string // table id for todos
}

// Health is a liveness probe that does not touch Appwrite.
func (h *Handlers) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": base.AppwriteMinVersion + "-" + base.AppwriteMaxVersion,
	})
}

// Capabilities reports which capabilities the process has wired.
func (h *Handlers) Capabilities(c *gin.Context) {
	report := map[string]bool{}
	if _, ok := h.Base.Auth(); ok {
		report["auth"] = true
	}
	if _, ok := h.Base.Storage(); ok {
		report["storage"] = true
	}
	if _, ok := h.Base.DB(); ok {
		report["db"] = true
	}
	if _, ok := h.Base.Functions(); ok {
		report["functions"] = true
	}
	c.JSON(http.StatusOK, report)
}
