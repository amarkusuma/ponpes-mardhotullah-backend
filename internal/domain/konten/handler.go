package konten

import (
	"fmt"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/simpeg/backend/pkg/apperror"
	mw "github.com/simpeg/backend/pkg/middleware"
	"github.com/simpeg/backend/pkg/pagination"
	"github.com/simpeg/backend/pkg/response"
	"github.com/simpeg/backend/pkg/storage"
)

type Handler struct {
	svc     *Service
	storage *storage.Storage
}

func NewHandler(svc *Service, store *storage.Storage) *Handler {
	return &Handler{svc: svc, storage: store}
}

// ── Artikel ───────────────────────────────────────────────────────────────────

// GET /api/admin/artikel  |  GET /api/publik/artikel
func (h *Handler) ListArtikel(c *gin.Context) {
	p := pagination.Parse(c)
	status := c.Query("status")
	kategori := c.Query("kategori")

	list, total, err := h.svc.ListArtikel(p, status, kategori)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.WithPagination(c, "daftar artikel", list, p.Page, p.Limit, total)
}

// GET /api/admin/artikel/:id
func (h *Handler) GetArtikel(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	a, err := h.svc.GetArtikel(id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "detail artikel", a)
}

// GET /api/publik/artikel/:slug
func (h *Handler) GetArtikelBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		response.Error(c, apperror.ErrValidation("slug wajib diisi"))
		return
	}
	a, err := h.svc.GetArtikelBySlug(slug)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "detail artikel", a)
}

// POST /api/admin/artikel
func (h *Handler) CreateArtikel(c *gin.Context) {
	user := mw.MustCurrentUser(c)
	var req CreateArtikelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}
	a, err := h.svc.CreateArtikel(user.ID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, "artikel berhasil dibuat", a)
}

// PUT /api/admin/artikel/:id
func (h *Handler) UpdateArtikel(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req UpdateArtikelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}
	a, err := h.svc.UpdateArtikel(id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "artikel berhasil diperbarui", a)
}

// DELETE /api/admin/artikel/:id
func (h *Handler) DeleteArtikel(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.DeleteArtikel(id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// ── Banner ────────────────────────────────────────────────────────────────────

// GET /api/admin/banner  |  GET /api/publik/banner
func (h *Handler) ListBanner(c *gin.Context) {
	onlyActive := c.Query("active_only") == "true"
	list, err := h.svc.ListBanner(onlyActive)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "daftar banner", list)
}

// POST /api/admin/banner
func (h *Handler) CreateBanner(c *gin.Context) {
	var req CreateBannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}
	b, err := h.svc.CreateBanner(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, "banner berhasil dibuat", b)
}

// PUT /api/admin/banner/:id
func (h *Handler) UpdateBanner(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req UpdateBannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}
	b, err := h.svc.UpdateBanner(id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "banner berhasil diperbarui", b)
}

// DELETE /api/admin/banner/:id
func (h *Handler) DeleteBanner(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.DeleteBanner(id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// ── Galeri ────────────────────────────────────────────────────────────────────

const galeriPrefix = "galeri/"
const maxFileSize = 5 << 20 // 5 MB

// POST /api/admin/galeri/upload
func (h *Handler) UploadGambarGaleri(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(maxFileSize); err != nil {
		response.Error(c, apperror.ErrValidation("file terlalu besar (maks 5MB)"))
		return
	}

	file, header, err := c.Request.FormFile("gambar")
	if err != nil {
		response.Error(c, apperror.ErrValidation("field 'gambar' wajib diisi"))
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true}
	if !allowed[ext] {
		response.Error(c, apperror.ErrValidation("format file tidak didukung (jpg, png, webp, gif)"))
		return
	}

	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixMilli(), uuid.New().String()[:8], ext)
	contentType := mime.TypeByExtension(ext)

	if err := h.storage.Put(c.Request.Context(), galeriPrefix+filename, file, header.Size, contentType); err != nil {
		response.Error(c, apperror.ErrInternal)
		return
	}

	url := "/uploads/galeri/" + filename
	response.OK(c, "gambar berhasil diupload", UploadGambarResponse{URL: url})
}

// GET /api/admin/galeri
func (h *Handler) ListGaleri(c *gin.Context) {
	p := pagination.Parse(c)
	kategori := c.Query("kategori")
	onlyPublished := c.Query("published_only") == "true"

	list, total, err := h.svc.ListGaleri(p, kategori, onlyPublished)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.WithPagination(c, "daftar galeri", list, p.Page, p.Limit, total)
}

// POST /api/admin/galeri
func (h *Handler) CreateGaleri(c *gin.Context) {
	var req CreateGaleriRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}
	g, err := h.svc.CreateGaleri(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, "galeri berhasil ditambahkan", g)
}

// PUT /api/admin/galeri/:id
func (h *Handler) UpdateGaleri(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req UpdateGaleriRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}
	g, err := h.svc.UpdateGaleri(id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "galeri berhasil diperbarui", g)
}

// DELETE /api/admin/galeri/:id
func (h *Handler) DeleteGaleri(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.DeleteGaleri(id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// RegisterAdminRoutes mendaftarkan route konten admin.
func (h *Handler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	artikel := rg.Group("/artikel")
	artikel.GET("", h.ListArtikel)
	artikel.GET("/:id", h.GetArtikel)
	artikel.POST("", h.CreateArtikel)
	artikel.PUT("/:id", h.UpdateArtikel)
	artikel.DELETE("/:id", h.DeleteArtikel)

	banner := rg.Group("/banner")
	banner.GET("", h.ListBanner)
	banner.POST("", h.CreateBanner)
	banner.PUT("/:id", h.UpdateBanner)
	banner.DELETE("/:id", h.DeleteBanner)

	galeri := rg.Group("/galeri")
	galeri.GET("", h.ListGaleri)
	galeri.POST("", h.CreateGaleri)
	galeri.PUT("/:id", h.UpdateGaleri)
	galeri.DELETE("/:id", h.DeleteGaleri)
	galeri.POST("/upload", h.UploadGambarGaleri)
}

// RegisterPublikRoutes untuk endpoint publik (tidak perlu auth).
func (h *Handler) RegisterPublikRoutes(rg *gin.RouterGroup) {
	rg.GET("/artikel", h.ListArtikel)
	rg.GET("/artikel/:slug", h.GetArtikelBySlug)
	rg.GET("/banner", h.ListBanner)
	rg.GET("/galeri", h.ListGaleri)
}

func parseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		return uuid.Nil, apperror.ErrValidation(param + " tidak valid")
	}
	return id, nil
}
