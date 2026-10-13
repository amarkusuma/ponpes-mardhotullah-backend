package server

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/simpeg/backend/pkg/storage"
)

// serveUpload men-stream object dari MinIO. Hanya prefix galeri/ yang publik;
// transkrip diakses lewat endpoint terautentikasi.
func serveUpload(store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := strings.TrimPrefix(c.Param("key"), "/")
		if !strings.HasPrefix(key, "galeri/") || strings.Contains(key, "..") {
			c.Status(http.StatusNotFound)
			return
		}

		obj, info, err := store.Get(c.Request.Context(), key)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				c.Status(http.StatusNotFound)
				return
			}
			c.Status(http.StatusInternalServerError)
			return
		}
		defer obj.Close()

		c.Header("Cache-Control", "public, max-age=86400")
		c.DataFromReader(http.StatusOK, info.Size, info.ContentType, obj, nil)
	}
}
