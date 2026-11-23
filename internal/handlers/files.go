package handlers

import (
	"io"
	"net/http"

	"github.com/appwrite/sdk-for-go/v7/file"
	"github.com/appwrite/sdk-for-go/v7/id"
	"github.com/appwrite/sdk-for-go/v7/models"
	"github.com/appwrite/sdk-for-go/v7/permission"
	"github.com/appwrite/sdk-for-go/v7/role"
	"github.com/e3tu/appwrite-base/storage"
	"github.com/gin-gonic/gin"
)

// UploadFile accepts a multipart field named "file" and stores it in the
// configured bucket.
//
// The content type comes from the multipart part, not from the client's
// filename, so a renamed file cannot spoof it.
func (h *Handlers) UploadFile(c *gin.Context) {
	st, ok := h.Base.Storage()
	if !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "storage capability disabled"})
		return
	}

	header, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "multipart field \"file\" is required"})
		return
	}

	fh, err := header.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer fh.Close()

	data, err := io.ReadAll(fh)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bucket := c.PostForm("bucket")
	if bucket == "" {
		bucket = st.DefaultBucketID()
	}
	if bucket == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no bucket configured"})
		return
	}

	created, err := st.Storage().CreateFile(
		bucket,
		id.Unique(),
		file.InputFile{Name: header.Filename, Data: data},
		st.Storage().WithCreateFilePermissions([]string{
			permission.Read(role.Any()),
			permission.Update(role.User(userID(c), "")),
			permission.Delete(role.User(userID(c), "")),
		}),
	)
	if err != nil {
		writeSDKError(c, err)
		return
	}

	c.JSON(http.StatusCreated, fileResponse(st, bucket, created))
}

// PublicFile returns the browser-facing URL for one file.
//
// The URL is built from the public endpoint, so this is the route to check
// when a deployment sits behind a reverse proxy.
func (h *Handlers) PublicFile(c *gin.Context) {
	st, ok := h.Base.Storage()
	if !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "storage capability disabled"})
		return
	}

	u, err := st.PublicFileURL(c.Query("bucket"), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": u})
}

func fileResponse(st *storage.Service, bucket string, f *models.File) gin.H {
	view, _ := st.PublicFileURL(bucket, f.Id)
	preview, _ := st.PublicPreviewURL(bucket, f.Id)

	return gin.H{
		"id":       f.Id,
		"name":     f.Name,
		"mimeType": f.MimeType,
		"size":     f.SizeOriginal,
		"url":      view,
		"preview":  preview,
	}
}
