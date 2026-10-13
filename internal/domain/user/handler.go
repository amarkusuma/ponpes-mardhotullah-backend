package user

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/simpeg/backend/pkg/apperror"
	"github.com/simpeg/backend/pkg/pagination"
	"github.com/simpeg/backend/pkg/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ── Dashboard ──────────────────────────────────────────────────────────────

// GET /api/admin/dashboard
func (h *Handler) Dashboard(c *gin.Context) {
	stats, err := h.svc.Dashboard()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "statistik dashboard", stats)
}

// ── User CRUD ──────────────────────────────────────────────────────────────

// GET /api/admin/users
func (h *Handler) List(c *gin.Context) {
	p := pagination.Parse(c)
	filter := ListUsersFilter{
		Search:   c.Query("search"),
		RoleName: c.Query("role"),
	}

	if active := c.Query("is_active"); active != "" {
		b := active == "true"
		filter.IsActive = &b
	}

	users, total, err := h.svc.List(p, filter)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.WithPagination(c, "daftar user", users, p.Page, p.Limit, total)
}

// GET /api/admin/users/:id
func (h *Handler) Get(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}

	u, err := h.svc.Get(id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "detail user", u)
}

// POST /api/admin/users
func (h *Handler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}

	u, err := h.svc.Create(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, "user berhasil dibuat", u)
}

// PUT /api/admin/users/:id
func (h *Handler) Update(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}

	u, err := h.svc.Update(id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "user berhasil diperbarui", u)
}

// DELETE /api/admin/users/:id
func (h *Handler) Delete(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}

	if err := h.svc.Delete(id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// PUT /api/admin/users/:id/roles
func (h *Handler) AssignRoles(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}

	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}

	if err := h.svc.AssignRoles(id, req); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "role berhasil diassign", nil)
}

// GET /api/admin/roles
func (h *Handler) ListRoles(c *gin.Context) {
	roles, err := h.svc.ListRoles()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "daftar role", roles)
}

// RegisterRoutes mendaftarkan semua route admin ke grup yang diberikan.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/dashboard", h.Dashboard)

	users := rg.Group("/users")
	users.GET("", h.List)
	users.GET("/:id", h.Get)
	users.POST("", h.Create)
	users.PUT("/:id", h.Update)
	users.DELETE("/:id", h.Delete)
	users.PUT("/:id/roles", h.AssignRoles)

	rg.GET("/roles", h.ListRoles)
}

func parseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		return uuid.Nil, apperror.ErrValidation("id tidak valid")
	}
	return id, nil
}
