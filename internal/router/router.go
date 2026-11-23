// Package router mounts the template's routes.
//
// Routes for a disabled capability are not registered at all, so a missing
// endpoint is a configuration fact rather than a runtime error.
package router

import (
	"github.com/E3TU/appwrite-go-kit-template/internal/handlers"
	"github.com/e3tu/appwrite-base"
	"github.com/gin-gonic/gin"
)

// New builds the Gin engine for b.
//
// sessionHeader is the optional custom session header the auth middleware
// reads before falling back to the cookie.
func New(b *base.Base, todosTable, sessionHeader string) *gin.Engine {
	h := &handlers.Handlers{Base: b, Todos: todosTable}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/healthz", h.Health)
	r.GET("/v1/capabilities", h.Capabilities)

	v1 := r.Group("/v1")

	_, hasDB := b.DB()
	_, hasStorage := b.Storage()
	authSvc, hasAuth := b.Auth()

	// Reads stay public so the template can be explored with curl.
	// Writes are registered on this group: guarded when auth is enabled,
	// open when it is not.
	writes := v1.Group("")
	if hasAuth {
		writes.Use(authSvc.Middleware(sessionHeader))
		writes.GET("/me", h.Me)
	}

	if hasDB {
		v1.GET("/todos", h.ListTodos)
		writes.POST("/todos", h.CreateTodo)
		writes.PUT("/todos/:id", h.UpdateTodo)
		writes.DELETE("/todos/:id", h.DeleteTodo)
	}
	if hasStorage {
		writes.POST("/files", h.UploadFile)
		v1.GET("/files/:id/url", h.PublicFile)
	}

	return r
}
