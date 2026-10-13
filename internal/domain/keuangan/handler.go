package keuangan

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/simpeg/backend/pkg/apperror"
	mw "github.com/simpeg/backend/pkg/middleware"
	"github.com/simpeg/backend/pkg/pagination"
	"github.com/simpeg/backend/pkg/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ── Admin Endpoints ───────────────────────────────────────────────────────────

// GET /api/admin/tagihan
func (h *Handler) AdminListTagihan(c *gin.Context) {
	p := pagination.Parse(c)
	filter := buildFilter(c)

	list, total, err := h.svc.ListTagihan(p, filter)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.WithPagination(c, "daftar tagihan", list, p.Page, p.Limit, total)
}

// POST /api/admin/tagihan
func (h *Handler) AdminCreateTagihan(c *gin.Context) {
	var req CreateTagihanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}
	t, err := h.svc.CreateTagihan(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, "tagihan berhasil dibuat", t)
}

// POST /api/admin/pembayaran
func (h *Handler) AdminCreatePembayaran(c *gin.Context) {
	var req CreatePembayaranRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}
	p, err := h.svc.CreatePembayaran(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, "pembayaran berhasil dicatat", p)
}

// ── Santri Endpoints ──────────────────────────────────────────────────────────

// GET /api/santri/tagihan
func (h *Handler) SantriTagihan(c *gin.Context) {
	user := mw.MustCurrentUser(c)
	list, err := h.svc.GetTagihanSantri(user.ID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "tagihan santri", list)
}

// ── Wali Endpoints ────────────────────────────────────────────────────────────

// GET /api/wali/tagihan-anak
func (h *Handler) WaliTagihanAnak(c *gin.Context) {
	// Wali lihat tagihan anaknya (butuh santri_id query param)
	raw := c.Query("santri_user_id")
	if raw == "" {
		response.Error(c, apperror.ErrValidation("santri_user_id wajib diisi"))
		return
	}
	santriUserID, err := uuid.Parse(raw)
	if err != nil {
		response.Error(c, apperror.ErrValidation("santri_user_id tidak valid"))
		return
	}

	list, err := h.svc.GetTagihanSantri(santriUserID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "tagihan anak", list)
}

// GET /api/wali/riwayat-pembayaran
func (h *Handler) WaliRiwayatPembayaran(c *gin.Context) {
	raw := c.Query("santri_user_id")
	if raw == "" {
		response.Error(c, apperror.ErrValidation("santri_user_id wajib diisi"))
		return
	}
	santriUserID, err := uuid.Parse(raw)
	if err != nil {
		response.Error(c, apperror.ErrValidation("santri_user_id tidak valid"))
		return
	}

	list, err := h.svc.GetRiwayatPembayaran(santriUserID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "riwayat pembayaran anak", list)
}

// GET /api/santri/riwayat-pembayaran (santri lihat miliknya sendiri)
func (h *Handler) SantriRiwayatPembayaran(c *gin.Context) {
	user := mw.MustCurrentUser(c)
	list, err := h.svc.GetRiwayatPembayaran(user.ID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "riwayat pembayaran", list)
}

// RegisterAdminRoutes mendaftarkan route keuangan admin.
func (h *Handler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	rg.GET("/tagihan", h.AdminListTagihan)
	rg.POST("/tagihan", h.AdminCreateTagihan)
	rg.POST("/pembayaran", h.AdminCreatePembayaran)
}

func (h *Handler) RegisterSantriRoutes(rg *gin.RouterGroup) {
	rg.GET("/tagihan", h.SantriTagihan)
	rg.GET("/riwayat-pembayaran", h.SantriRiwayatPembayaran)
}

func (h *Handler) RegisterWaliRoutes(rg *gin.RouterGroup) {
	rg.GET("/tagihan-anak", h.WaliTagihanAnak)
	rg.GET("/riwayat-pembayaran", h.WaliRiwayatPembayaran)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func buildFilter(c *gin.Context) ListTagihanFilter {
	f := ListTagihanFilter{
		Status: c.Query("status"),
	}
	if raw := c.Query("santri_id"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			f.SantriID = &id
		}
	}
	if bulan := c.GetInt("bulan"); bulan > 0 {
		f.Bulan = bulan
	}
	if tahun := c.GetInt("tahun"); tahun > 0 {
		f.Tahun = tahun
	}
	return f
}
