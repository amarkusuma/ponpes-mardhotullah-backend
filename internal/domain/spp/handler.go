package spp

import (
	"encoding/csv"
	"strconv"

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

// POST /generate
func (h *Handler) Generate(c *gin.Context) {
	var req GenerateSPPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}
	result, err := h.svc.GenerateSPP(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, "tagihan berhasil digenerate", result)
}

// GET /
func (h *Handler) List(c *gin.Context) {
	p := pagination.Parse(c)
	filter := buildFilter(c)
	list, total, err := h.svc.ListTagihan(p, filter)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.WithPagination(c, "daftar tagihan SPP", list, p.Page, p.Limit, total)
}

// GET /:id
func (h *Handler) Detail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperror.ErrValidation("id tidak valid"))
		return
	}
	detail, err := h.svc.GetDetail(id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "detail tagihan SPP", detail)
}

// POST /:id/pay
func (h *Handler) Pay(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperror.ErrValidation("id tidak valid"))
		return
	}
	var req BayarTagihanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}
	result, err := h.svc.BayarTagihan(id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, "pembayaran berhasil dicatat", result)
}

// GET /:id/payments
func (h *Handler) Payments(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperror.ErrValidation("id tidak valid"))
		return
	}
	list, err := h.svc.GetPembayaran(id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "riwayat pembayaran", list)
}

// GET /rekap?tahun=2025&jenis=SPP
func (h *Handler) Rekap(c *gin.Context) {
	tahun, _ := strconv.Atoi(c.Query("tahun"))
	jenis := c.DefaultQuery("jenis", "SPP")
	list, err := h.svc.GetRekap(jenis, tahun)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "rekap SPP", list)
}

// GET /unpaid?bulan=1&tahun=2025&jenis=SPP
func (h *Handler) Unpaid(c *gin.Context) {
	bulan, _ := strconv.Atoi(c.Query("bulan"))
	tahun, _ := strconv.Atoi(c.Query("tahun"))
	jenis := c.DefaultQuery("jenis", "SPP")
	list, err := h.svc.GetUnpaid(jenis, bulan, tahun)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "tagihan belum bayar", list)
}

// GET /export?format=csv (default csv)
func (h *Handler) Export(c *gin.Context) {
	filter := buildFilter(c)
	rows, err := h.svc.GetAllForExport(filter)
	if err != nil {
		response.Error(c, err)
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=spp-export.csv")
	c.Header("Cache-Control", "no-cache")

	w := csv.NewWriter(c.Writer)
	_ = w.Write([]string{
		"Nama Santri", "NIS", "Kelas", "Jenis",
		"Bulan", "Tahun", "Nominal", "Total Dibayar", "Sisa", "Status",
	})
	bulanNames := []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	for _, r := range rows {
		bulanStr := bulanNames[r.Bulan]
		_ = w.Write([]string{
			r.NamaSantri,
			r.NIK,
			r.KelasNama,
			r.Jenis,
			bulanStr,
			strconv.Itoa(r.Tahun),
			formatRupiah(r.Nominal),
			formatRupiah(r.TotalDibayar),
			formatRupiah(r.Sisa),
			string(r.Status),
		})
	}
	w.Flush()
}

// RegisterRoutes mendaftarkan semua endpoint SPP ke router group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/generate", h.Generate)
	rg.GET("", h.List)
	rg.GET("/rekap", h.Rekap)
	rg.GET("/unpaid", h.Unpaid)
	rg.GET("/export", h.Export)
	rg.GET("/:id", h.Detail)
	rg.POST("/:id/pay", h.Pay)
	rg.GET("/:id/payments", h.Payments)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func buildFilter(c *gin.Context) ListSPPFilter {
	f := ListSPPFilter{
		Search: c.Query("search"),
		Status: c.Query("status"),
		Jenis:  c.DefaultQuery("jenis", ""),
	}
	if b, err := strconv.Atoi(c.Query("bulan")); err == nil {
		f.Bulan = b
	}
	if y, err := strconv.Atoi(c.Query("tahun")); err == nil {
		f.Tahun = y
	}
	if raw := c.Query("kelas_id"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			f.KelasID = &id
		}
	}
	return f
}
