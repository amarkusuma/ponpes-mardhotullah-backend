package konten

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/simpeg/backend/pkg/apperror"
	mw "github.com/simpeg/backend/pkg/middleware"
	"github.com/simpeg/backend/pkg/pagination"
	"github.com/simpeg/backend/pkg/response"
	"github.com/simpeg/backend/pkg/storage"
)

type Handler struct {
	svc     *Service
	storage *storage.Storage
	redis   *redis.Client
}

func NewHandler(svc *Service, store *storage.Storage, rdb *redis.Client) *Handler {
	return &Handler{svc: svc, storage: store, redis: rdb}
}

// Nilai query param ?sort= untuk list artikel publik.
const (
	SortPopuler = "populer"
	SortTerbaru = "terbaru"
)

const (
	defaultTerkaitLimit = 4
	maxTerkaitLimit     = 10
	// Satu IP hanya dihitung sekali per artikel dalam jendela ini.
	viewDedupWindow = time.Hour
)

// ── Artikel ───────────────────────────────────────────────────────────────────

// GET /api/admin/artikel
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

// GET /api/publik/artikel?sort=terbaru|populer&kategori=&search=&page=&limit=
// Selalu hanya artikel published, apa pun query status yang dikirim.
func (h *Handler) ListArtikelPublik(c *gin.Context) {
	p := pagination.Parse(c)
	list, total, err := h.svc.ListArtikelPublik(p, c.Query("kategori"), c.Query("sort"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.WithPagination(c, "daftar artikel", list, p.Page, p.Limit, total)
}

// GET /api/publik/artikel/:slug
func (h *Handler) GetArtikelPublikBySlug(c *gin.Context) {
	a, err := h.svc.GetArtikelPublikBySlug(c.Param("slug"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "detail artikel", a)
}

// GET /api/publik/artikel/:slug/terkait?limit=4
func (h *Handler) ListArtikelTerkait(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit < 1 {
		limit = defaultTerkaitLimit
	}
	if limit > maxTerkaitLimit {
		limit = maxTerkaitLimit
	}
	list, err := h.svc.ListArtikelTerkait(c.Param("slug"), limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "artikel terkait", list)
}

// POST /api/publik/artikel/:slug/view
// Dicatat terpisah dari GET detail karena halaman detail di-cache (ISR) di web.
func (h *Handler) TambahViewArtikel(c *gin.Context) {
	slug := c.Param("slug")
	if h.redis != nil {
		key := "artikel:view:" + slug + ":" + c.ClientIP()
		first, err := h.redis.SetNX(c.Request.Context(), key, 1, viewDedupWindow).Result()
		if err == nil && !first {
			c.Status(http.StatusNoContent)
			return
		}
	}
	if err := h.svc.TambahViewArtikel(slug); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
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

// GET /api/admin/banner?halaman=&active_only=true
func (h *Handler) ListBanner(c *gin.Context) {
	onlyActive := c.Query("active_only") == "true"
	list, err := h.svc.ListBanner(c.Query("halaman"), onlyActive)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "daftar banner", list)
}

// GET /api/publik/banner?halaman=home
// Hanya banner aktif, urut berdasarkan urutan.
func (h *Handler) ListBannerPublik(c *gin.Context) {
	list, err := h.svc.ListBanner(c.Query("halaman"), true)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "daftar banner", list)
}

// POST /api/admin/banner/upload
func (h *Handler) UploadGambarBanner(c *gin.Context) {
	h.uploadGambar(c, bannerPrefix)
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

const (
	galeriPrefix  = "galeri/"
	artikelPrefix = "artikel/"
	newsPrefix    = "news/"
	bannerPrefix  = "banner/"
	pustakaPrefix = "pustaka/"
)
const (
	maxFileSize = 5 << 20  // 5 MB (gambar)
	maxPDFSize  = 25 << 20 // 25 MB (pdf pusat informasi)
)

// POST /api/admin/galeri/upload
func (h *Handler) UploadGambarGaleri(c *gin.Context) {
	h.uploadGambar(c, galeriPrefix)
}

// POST /api/admin/artikel/upload-thumbnail
func (h *Handler) UploadThumbnailArtikel(c *gin.Context) {
	h.uploadGambar(c, artikelPrefix)
}

// uploadSpec mendeskripsikan jenis file yang boleh diupload.
type uploadSpec struct {
	field   string          // nama field multipart
	maxSize int64           // byte
	exts    map[string]bool // ekstensi (huruf kecil, dengan titik)
	// magic: prefix byte awal file yang wajib cocok (kosong = tidak dicek)
	magic   []byte
	label   string // untuk pesan error, mis. "jpg, png, webp, gif"
	message string // pesan sukses
}

var (
	gambarSpec = uploadSpec{
		field:   "gambar",
		maxSize: maxFileSize,
		exts:    map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true},
		label:   "jpg, png, webp, gif",
		message: "gambar berhasil diupload",
	}
	pdfSpec = uploadSpec{
		field:   "file",
		maxSize: maxPDFSize,
		exts:    map[string]bool{".pdf": true},
		magic:   []byte("%PDF-"),
		label:   "pdf",
		message: "file berhasil diupload",
	}
)

// uploadGambar menyimpan field multipart 'gambar' ke storage di bawah prefix
// dan mengembalikan URL publiknya (/uploads/<prefix><filename>).
func (h *Handler) uploadGambar(c *gin.Context, prefix string) {
	h.uploadFile(c, prefix, gambarSpec)
}

func (h *Handler) uploadFile(c *gin.Context, prefix string, spec uploadSpec) {
	tooLarge := apperror.ErrValidation(fmt.Sprintf("file terlalu besar (maks %dMB)", spec.maxSize>>20))

	// Batasi body request; ParseMultipartForm sendiri hanya membatasi memori.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, spec.maxSize+(1<<20))
	if err := c.Request.ParseMultipartForm(maxFileSize); err != nil {
		response.Error(c, tooLarge)
		return
	}

	file, header, err := c.Request.FormFile(spec.field)
	if err != nil {
		response.Error(c, apperror.ErrValidation(fmt.Sprintf("field '%s' wajib diisi", spec.field)))
		return
	}
	defer file.Close()

	if header.Size > spec.maxSize {
		response.Error(c, tooLarge)
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !spec.exts[ext] {
		response.Error(c, apperror.ErrValidation(fmt.Sprintf("format file tidak didukung (%s)", spec.label)))
		return
	}

	if len(spec.magic) > 0 {
		head := make([]byte, len(spec.magic))
		if _, err := io.ReadFull(file, head); err != nil || !bytes.Equal(head, spec.magic) {
			response.Error(c, apperror.ErrValidation(fmt.Sprintf("isi file bukan %s yang valid", spec.label)))
			return
		}
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			response.Error(c, apperror.ErrInternal)
			return
		}
	}

	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixMilli(), uuid.New().String()[:8], ext)
	contentType := mime.TypeByExtension(ext)

	if err := h.storage.Put(c.Request.Context(), prefix+filename, file, header.Size, contentType); err != nil {
		response.Error(c, apperror.ErrInternal)
		return
	}

	response.OK(c, spec.message, UploadFileResponse{
		URL:  "/uploads/" + prefix + filename,
		Size: header.Size,
	})
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
	artikel.POST("/upload-thumbnail", h.UploadThumbnailArtikel)

	banner := rg.Group("/banner")
	banner.GET("", h.ListBanner)
	banner.POST("", h.CreateBanner)
	banner.PUT("/:id", h.UpdateBanner)
	banner.DELETE("/:id", h.DeleteBanner)
	banner.POST("/upload", h.UploadGambarBanner)

	news := rg.Group("/news")
	news.GET("", h.ListNews)
	news.GET("/:id", h.GetNews)
	news.POST("", h.CreateNews)
	news.PUT("/:id", h.UpdateNews)
	news.DELETE("/:id", h.DeleteNews)
	news.POST("/upload-thumbnail", h.UploadThumbnailNews)

	pustaka := rg.Group("/pusat-informasi")
	pustaka.GET("", h.ListPustaka)
	pustaka.GET("/:id", h.GetPustaka)
	pustaka.POST("", h.CreatePustaka)
	pustaka.PUT("/:id", h.UpdatePustaka)
	pustaka.DELETE("/:id", h.DeletePustaka)
	pustaka.POST("/upload-thumbnail", h.UploadThumbnailPustaka)
	pustaka.POST("/upload-file", h.UploadFilePustaka)

	galeri := rg.Group("/galeri")
	galeri.GET("", h.ListGaleri)
	galeri.POST("", h.CreateGaleri)
	galeri.PUT("/:id", h.UpdateGaleri)
	galeri.DELETE("/:id", h.DeleteGaleri)
	galeri.POST("/upload", h.UploadGambarGaleri)
}

// RegisterPublikRoutes untuk endpoint publik (tidak perlu auth).
func (h *Handler) RegisterPublikRoutes(rg *gin.RouterGroup) {
	rg.GET("/artikel", h.ListArtikelPublik)
	rg.GET("/artikel/:slug", h.GetArtikelPublikBySlug)
	rg.GET("/artikel/:slug/terkait", h.ListArtikelTerkait)
	rg.POST("/artikel/:slug/view", h.TambahViewArtikel)
	rg.GET("/news", h.ListNewsPublik)
	rg.GET("/news/terbaru", h.ListNewsTerbaru)
	rg.GET("/news/:slug", h.GetNewsPublikBySlug)
	rg.GET("/pusat-informasi", h.ListPustakaPublik)
	rg.GET("/pusat-informasi/kategori", h.ListKategoriPustaka)
	rg.GET("/pusat-informasi/:slug", h.GetPustakaPublikBySlug)
	rg.GET("/banner", h.ListBannerPublik)
	rg.GET("/galeri", h.ListGaleri)
}

func parseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		return uuid.Nil, apperror.ErrValidation(param + " tidak valid")
	}
	return id, nil
}
