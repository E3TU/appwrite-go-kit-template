package handlers

import (
	"net/http"

	"github.com/appwrite/sdk-for-go/v7/id"
	"github.com/appwrite/sdk-for-go/v7/models"
	"github.com/appwrite/sdk-for-go/v7/permission"
	"github.com/appwrite/sdk-for-go/v7/query"
	"github.com/appwrite/sdk-for-go/v7/role"
	"github.com/e3tu/appwrite-base/database"
	"github.com/gin-gonic/gin"
)

// Todo is the wire shape for one row of the todos table.
//
// The Appwrite table needs the columns: title (varchar), desc (text),
// completed (boolean).
type Todo struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Desc      string `json:"desc"`
	Completed bool   `json:"completed"`
}

// TodoInput is the accepted request body for create and update.
type TodoInput struct {
	Title     string `json:"title"`
	Desc      string `json:"desc"`
	Completed bool   `json:"completed"`
}

// db returns the database service or writes 503 when it is disabled.
func (h *Handlers) db(c *gin.Context) (*database.Service, bool) {
	db, ok := h.Base.DB()
	if !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database capability disabled"})
		return nil, false
	}
	return db, true
}

// ListTodos returns the newest rows first. Supports ?limit= and ?cursor=.
func (h *Handlers) ListTodos(c *gin.Context) {
	db, ok := h.db(c)
	if !ok {
		return
	}

	queries := []string{query.OrderDesc("$createdAt")}
	if limit, ok := queryLimit(c.Query("limit")); ok {
		queries = append(queries, limit)
	}
	if cursor := c.Query("cursor"); cursor != "" {
		queries = append(queries, query.CursorAfter(cursor))
	}

	list, err := db.TablesDB().ListRows(
		db.DatabaseID(),
		h.Todos,
		db.TablesDB().WithListRowsQueries(queries),
	)
	if err != nil {
		writeSDKError(c, err)
		return
	}

	items := make([]Todo, 0, len(list.Rows))
	for i := range list.Rows {
		items = append(items, todoFromRow(&list.Rows[i]))
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": list.Total,
	})
}

// CreateTodo creates a row readable by anyone and writable by its owner.
func (h *Handlers) CreateTodo(c *gin.Context) {
	db, ok := h.db(c)
	if !ok {
		return
	}

	var in TodoInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if in.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}

	row, err := db.TablesDB().CreateRow(
		db.DatabaseID(),
		h.Todos,
		id.Unique(),
		map[string]any{
			"title":     in.Title,
			"desc":      in.Desc,
			"completed": in.Completed,
		},
		db.TablesDB().WithCreateRowPermissions(rowPermissions(userID(c))),
	)
	if err != nil {
		writeSDKError(c, err)
		return
	}

	c.JSON(http.StatusCreated, todoFromRow(row))
}

// UpdateTodo replaces the mutable fields of one row.
func (h *Handlers) UpdateTodo(c *gin.Context) {
	db, ok := h.db(c)
	if !ok {
		return
	}

	var in TodoInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	row, err := db.TablesDB().UpdateRow(
		db.DatabaseID(),
		h.Todos,
		c.Param("id"),
		db.TablesDB().WithUpdateRowData(map[string]any{
			"title":     in.Title,
			"desc":      in.Desc,
			"completed": in.Completed,
		}),
	)
	if err != nil {
		writeSDKError(c, err)
		return
	}

	c.JSON(http.StatusOK, todoFromRow(row))
}

// DeleteTodo removes one row.
func (h *Handlers) DeleteTodo(c *gin.Context) {
	db, ok := h.db(c)
	if !ok {
		return
	}

	if _, err := db.TablesDB().DeleteRow(db.DatabaseID(), h.Todos, c.Param("id")); err != nil {
		writeSDKError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// rowPermissions is read-for-anyone plus write for the owner, when there is
// one. Anonymous callers get no write permission rather than a wildcard.
func rowPermissions(uid string) []string {
	perms := []string{permission.Read(role.Any())}
	if uid == "" {
		return perms
	}
	return append(perms,
		permission.Update(role.User(uid, "")),
		permission.Delete(role.User(uid, "")),
	)
}

// todoFromRow decodes the row's column data into the wire shape.
//
// The SDK hides the raw row payload behind Decode rather than exposing a
// column map, so the table's columns are declared once here.
func todoFromRow(row *models.Row) Todo {
	var stored struct {
		Title     string `json:"title"`
		Desc      string `json:"desc"`
		Completed bool   `json:"completed"`
	}
	_ = row.Decode(&stored)

	return Todo{
		ID:        row.Id,
		Title:     stored.Title,
		Desc:      stored.Desc,
		Completed: stored.Completed,
	}
}

// userID returns the authenticated user id, or "" for anonymous callers.
func userID(c *gin.Context) string {
	u, ok := userFromContext(c)
	if !ok {
		return ""
	}
	return u.ID
}
