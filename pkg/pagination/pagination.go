// Package pagination menyediakan helper untuk query paginasi dari request.
package pagination

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// Params mengemas parameter paginasi, sorting, dan search dari query string.
type Params struct {
	Page   int
	Limit  int
	SortBy string
	Order  string
	Search string
}

// Parse membaca query params dari gin.Context dengan nilai default yang aman.
func Parse(c *gin.Context) Params {
	page := c.GetInt("page")
	if page < 1 {
		page = 1
	}

	limit := c.GetInt("limit")
	if limit < 1 || limit > 100 {
		limit = 10
	}

	sortBy := c.DefaultQuery("sort_by", "created_at")
	// Whitelist kolom sort untuk mencegah SQL injection
	allowed := map[string]bool{
		"created_at": true, "updated_at": true, "name": true,
		"email": true, "status": true, "nominal": true, "tahun": true, "bulan": true,
	}
	if !allowed[sortBy] {
		sortBy = "created_at"
	}

	order := strings.ToLower(c.DefaultQuery("order", "desc"))
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	return Params{
		Page:   page,
		Limit:  limit,
		SortBy: sortBy,
		Order:  order,
		Search: c.Query("search"),
	}
}

// Offset menghitung offset untuk query SQL.
func (p Params) Offset() int { return (p.Page - 1) * p.Limit }

// OrderClause mengembalikan string ORDER BY yang aman.
func (p Params) OrderClause() string { return p.SortBy + " " + p.Order }
