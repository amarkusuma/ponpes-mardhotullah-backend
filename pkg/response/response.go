// Package response menyediakan helper untuk format JSON yang konsisten.
package response

import (
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/simpeg/backend/pkg/apperror"
)

// Envelope adalah envelope standar seluruh response API.
type Envelope struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	Pagination *Meta       `json:"pagination,omitempty"`
}

// Meta berisi informasi paginasi.
type Meta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func OK(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Envelope{Success: true, Message: message, Data: data})
}

func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Envelope{Success: true, Message: message, Data: data})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// WithPagination menulis response dengan metadata paginasi.
func WithPagination(c *gin.Context, message string, data interface{}, page, limit int, total int64) {
	c.JSON(http.StatusOK, Envelope{
		Success: true,
		Message: message,
		Data:    data,
		Pagination: &Meta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: int(math.Ceil(float64(total) / float64(limit))),
		},
	})
}

// Error menerjemahkan error ke response HTTP yang sesuai.
func Error(c *gin.Context, err error) {
	var appErr *apperror.AppError
	switch e := err.(type) {
	case *apperror.AppError:
		appErr = e
	default:
		appErr = apperror.ErrInternal
	}

	body := gin.H{"detail": appErr.Detail}
	if appErr.Detail == "" {
		body = nil
	}

	c.AbortWithStatusJSON(appErr.Code, Envelope{
		Success: false,
		Message: appErr.Message,
		Data:    body,
	})
}
