package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/simpeg/backend/pkg/apperror"
	mw "github.com/simpeg/backend/pkg/middleware"
	"github.com/simpeg/backend/pkg/response"
)

// Handler menangani semua request HTTP domain autentikasi.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Login godoc
// POST /api/auth/login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}

	result, err := h.svc.Login(req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "login berhasil", result)
}

// Logout godoc
// POST /api/auth/logout
func (h *Handler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}

	if err := h.svc.Logout(req.RefreshToken); err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "logout berhasil", nil)
}

// Refresh godoc
// POST /api/auth/refresh
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}

	pair, err := h.svc.Refresh(req.RefreshToken)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "token diperbarui", pair)
}

// Me godoc
// GET /api/auth/me — protected
func (h *Handler) Me(c *gin.Context) {
	currentUser := mw.MustCurrentUser(c)

	user, err := h.svc.Me(currentUser.ID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "profil user", toUserInfo(user))
}

// RegisterRoutes mendaftarkan semua route auth ke grup yang diberikan.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	auth := rg.Group("/auth")
	auth.POST("/login", h.Login)
	auth.POST("/logout", h.Logout)
	auth.POST("/refresh", h.Refresh)
	auth.GET("/me", authMW, h.Me)
}

// helper untuk decode UUID dari path param
func parseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		return uuid.Nil, apperror.ErrValidation("id tidak valid")
	}
	return id, nil
}
