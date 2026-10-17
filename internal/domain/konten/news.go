package konten

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	modelkonten "github.com/simpeg/backend/internal/model/konten"
	"github.com/simpeg/backend/pkg/apperror"
	mw "github.com/simpeg/backend/pkg/middleware"
	"github.com/simpeg/backend/pkg/pagination"
	"github.com/simpeg/backend/pkg/response"
)

// ── DTO ───────────────────────────────────────────────────────────────────────

type CreateNewsRequest struct {
	Judul     string                    `json:"judul"     binding:"required,min=5"`
	Slug      string                    `json:"slug"      binding:"required"`
	Kategori  modelkonten.KategoriNews  `json:"kategori"  binding:"required,oneof=kegiatan prestasi pengumuman agenda lainnya"`
	Thumbnail string                    `json:"thumbnail"`
	Konten    string                    `json:"konten"    binding:"required"`
	Status    modelkonten.StatusArtikel `json:"status"    binding:"omitempty,oneof=draft published archived"`
}

type UpdateNewsRequest struct {
	Judul     string                    `json:"judul"     binding:"omitempty,min=5"`
	Kategori  modelkonten.KategoriNews  `json:"kategori"  binding:"omitempty,oneof=kegiatan prestasi pengumuman agenda lainnya"`
	Thumbnail *string                   `json:"thumbnail"` // nil = tidak diubah, "" = hapus thumbnail
	Konten    string                    `json:"konten"`
	Status    modelkonten.StatusArtikel `json:"status"    binding:"omitempty,oneof=draft published archived"`
}

type NewsResponse struct {
	ID        uuid.UUID                 `json:"id"`
	Judul     string                    `json:"judul"`
	Slug      string                    `json:"slug"`
	Kategori  modelkonten.KategoriNews  `json:"kategori"`
	Thumbnail string                    `json:"thumbnail,omitempty"`
	Konten    string                    `json:"konten,omitempty"`
	Status    modelkonten.StatusArtikel `json:"status"`
	Penulis   string                    `json:"penulis"`
	CreatedAt time.Time                 `json:"created_at"`
	UpdatedAt time.Time                 `json:"updated_at"`
}

func toNewsResponse(n modelkonten.News) NewsResponse {
	return NewsResponse{
		ID:        n.ID,
		Judul:     n.Judul,
		Slug:      n.Slug,
		Kategori:  n.Kategori,
		Thumbnail: n.Thumbnail,
		Konten:    n.Konten,
		Status:    n.Status,
		Penulis:   n.Penulis.Name,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
}

func toNewsResponses(list []modelkonten.News) []NewsResponse {
	result := make([]NewsResponse, 0, len(list))
	for _, n := range list {
		result = append(result, toNewsResponse(n))
	}
	return result
}

// ── Repository ────────────────────────────────────────────────────────────────

// newsListColumns: kolom untuk daftar ringkas (tanpa konten agar response ringan).
var newsListColumns = []string{
	"id", "created_at", "updated_at", "judul", "slug", "kategori",
	"thumbnail", "status", "penulis_id",
}

type NewsFilter struct {
	Search        string
	Status        string
	Kategori      string
	OnlyPublished bool
}

func (r *Repository) ListNews(params pagination.Params, f NewsFilter) ([]modelkonten.News, int64, error) {
	var list []modelkonten.News
	var total int64

	q := r.db.Model(&modelkonten.News{})
	if f.OnlyPublished {
		q = q.Where("status = ?", modelkonten.StatusPublished)
	} else if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Search != "" {
		q = q.Where("judul ILIKE ?", "%"+f.Search+"%")
	}
	if f.Kategori != "" {
		q = q.Where("kategori = ?", f.Kategori)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := q.Select(newsListColumns).Preload("Penulis").
		Order("created_at desc").Limit(params.Limit).Offset(params.Offset()).Find(&list).Error
	return list, total, err
}

func (r *Repository) ListNewsTerbaru(limit int) ([]modelkonten.News, error) {
	var list []modelkonten.News
	err := r.db.Model(&modelkonten.News{}).Select(newsListColumns).Preload("Penulis").
		Where("status = ?", modelkonten.StatusPublished).
		Order("created_at desc").Limit(limit).Find(&list).Error
	return list, err
}

func (r *Repository) FindNewsByID(id uuid.UUID) (*modelkonten.News, error) {
	var n modelkonten.News
	err := r.db.Preload("Penulis").First(&n, "id = ?", id).Error
	return &n, err
}

func (r *Repository) FindNewsBySlug(slug string, onlyPublished bool) (*modelkonten.News, error) {
	var n modelkonten.News
	q := r.db.Preload("Penulis").Where("slug = ?", slug)
	if onlyPublished {
		q = q.Where("status = ?", modelkonten.StatusPublished)
	}
	err := q.First(&n).Error
	return &n, err
}

// SlugNewsExists juga memeriksa baris soft-delete karena unique index tetap berlaku.
func (r *Repository) SlugNewsExists(slug string) (bool, error) {
	var count int64
	err := r.db.Unscoped().Model(&modelkonten.News{}).Where("slug = ?", slug).Count(&count).Error
	return count > 0, err
}

func (r *Repository) CreateNews(n *modelkonten.News) error { return r.db.Create(n).Error }
func (r *Repository) UpdateNews(n *modelkonten.News) error { return r.db.Save(n).Error }
func (r *Repository) DeleteNews(id uuid.UUID) error {
	return r.db.Delete(&modelkonten.News{}, "id = ?", id).Error
}

// ── Service ───────────────────────────────────────────────────────────────────

func (s *Service) ListNews(params pagination.Params, f NewsFilter) ([]NewsResponse, int64, error) {
	list, total, err := s.repo.ListNews(params, f)
	if err != nil {
		return nil, 0, apperror.ErrInternal
	}
	return toNewsResponses(list), total, nil
}

func (s *Service) ListNewsTerbaru(limit int) ([]NewsResponse, error) {
	list, err := s.repo.ListNewsTerbaru(limit)
	if err != nil {
		return nil, apperror.ErrInternal
	}
	return toNewsResponses(list), nil
}

func (s *Service) GetNews(id uuid.UUID) (*NewsResponse, error) {
	n, err := s.repo.FindNewsByID(id)
	if err != nil {
		return nil, notFoundOrInternal(err, "berita")
	}
	r := toNewsResponse(*n)
	return &r, nil
}

func (s *Service) GetNewsPublikBySlug(slug string) (*NewsResponse, error) {
	n, err := s.repo.FindNewsBySlug(slug, true)
	if err != nil {
		return nil, notFoundOrInternal(err, "berita")
	}
	r := toNewsResponse(*n)
	return &r, nil
}

func (s *Service) CreateNews(penulisID uuid.UUID, req CreateNewsRequest) (*NewsResponse, error) {
	exists, err := s.repo.SlugNewsExists(req.Slug)
	if err != nil {
		return nil, apperror.ErrInternal
	}
	if exists {
		return nil, apperror.ErrConflictMsg("slug")
	}

	status := req.Status
	if status == "" {
		status = modelkonten.StatusDraft
	}
	n := &modelkonten.News{
		Judul:     req.Judul,
		Slug:      req.Slug,
		Kategori:  req.Kategori,
		Thumbnail: req.Thumbnail,
		Konten:    req.Konten,
		Status:    status,
		PenulisID: penulisID,
	}
	if err := s.repo.CreateNews(n); err != nil {
		return nil, apperror.ErrInternal
	}
	return s.GetNews(n.ID)
}

func (s *Service) UpdateNews(id uuid.UUID, req UpdateNewsRequest) (*NewsResponse, error) {
	n, err := s.repo.FindNewsByID(id)
	if err != nil {
		return nil, notFoundOrInternal(err, "berita")
	}

	if req.Judul != "" {
		n.Judul = req.Judul
	}
	if req.Kategori != "" {
		n.Kategori = req.Kategori
	}
	if req.Thumbnail != nil {
		n.Thumbnail = *req.Thumbnail
	}
	if req.Konten != "" {
		n.Konten = req.Konten
	}
	if req.Status != "" {
		n.Status = req.Status
	}

	if err := s.repo.UpdateNews(n); err != nil {
		return nil, apperror.ErrInternal
	}
	r := toNewsResponse(*n)
	return &r, nil
}

func (s *Service) DeleteNews(id uuid.UUID) error {
	if _, err := s.repo.FindNewsByID(id); err != nil {
		return notFoundOrInternal(err, "berita")
	}
	if err := s.repo.DeleteNews(id); err != nil {
		return apperror.ErrInternal
	}
	return nil
}

func notFoundOrInternal(err error, name string) error {
	if err == gorm.ErrRecordNotFound {
		return apperror.ErrNotFoundMsg(name)
	}
	return apperror.ErrInternal
}

// ── Handler ───────────────────────────────────────────────────────────────────

const (
	defaultNewsTerbaruLimit = 3
	maxNewsTerbaruLimit     = 20
)

// GET /api/admin/news?status=&kategori=&search=&page=&limit=
func (h *Handler) ListNews(c *gin.Context) {
	p := pagination.Parse(c)
	list, total, err := h.svc.ListNews(p, NewsFilter{
		Search:   p.Search,
		Status:   c.Query("status"),
		Kategori: c.Query("kategori"),
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.WithPagination(c, "daftar berita", list, p.Page, p.Limit, total)
}

// GET /api/admin/news/:id
func (h *Handler) GetNews(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	n, err := h.svc.GetNews(id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "detail berita", n)
}

// POST /api/admin/news
func (h *Handler) CreateNews(c *gin.Context) {
	user := mw.MustCurrentUser(c)
	var req CreateNewsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}
	n, err := h.svc.CreateNews(user.ID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, "berita berhasil dibuat", n)
}

// PUT /api/admin/news/:id
func (h *Handler) UpdateNews(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req UpdateNewsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}
	n, err := h.svc.UpdateNews(id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "berita berhasil diperbarui", n)
}

// DELETE /api/admin/news/:id
func (h *Handler) DeleteNews(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.DeleteNews(id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// POST /api/admin/news/upload-thumbnail
func (h *Handler) UploadThumbnailNews(c *gin.Context) {
	h.uploadGambar(c, newsPrefix)
}

// GET /api/publik/news?kategori=&search=&page=&limit=
// Selalu hanya berita published; field konten tidak disertakan.
func (h *Handler) ListNewsPublik(c *gin.Context) {
	p := pagination.Parse(c)
	list, total, err := h.svc.ListNews(p, NewsFilter{
		Search:        p.Search,
		Kategori:      c.Query("kategori"),
		OnlyPublished: true,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.WithPagination(c, "daftar berita", list, p.Page, p.Limit, total)
}

// GET /api/publik/news/terbaru?limit=3
func (h *Handler) ListNewsTerbaru(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit < 1 {
		limit = defaultNewsTerbaruLimit
	}
	if limit > maxNewsTerbaruLimit {
		limit = maxNewsTerbaruLimit
	}
	list, err := h.svc.ListNewsTerbaru(limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "berita terbaru", list)
}

// GET /api/publik/news/:slug
func (h *Handler) GetNewsPublikBySlug(c *gin.Context) {
	n, err := h.svc.GetNewsPublikBySlug(c.Param("slug"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "detail berita", n)
}
