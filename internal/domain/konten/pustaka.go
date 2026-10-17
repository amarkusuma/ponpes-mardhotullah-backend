package konten

// Pusat Informasi: perpustakaan digital berisi dokumen PDF (tabel digital_library).

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	modelkonten "github.com/simpeg/backend/internal/model/konten"
	"github.com/simpeg/backend/pkg/apperror"
	"github.com/simpeg/backend/pkg/pagination"
	"github.com/simpeg/backend/pkg/response"
)

// ── DTO ───────────────────────────────────────────────────────────────────────

type CreatePustakaRequest struct {
	Judul     string                    `json:"judul"     binding:"required,min=3,max=500"`
	Slug      string                    `json:"slug"      binding:"required,max=600"`
	Kategori  string                    `json:"kategori"  binding:"max=100"`
	Thumbnail string                    `json:"thumbnail" binding:"max=500"`
	FileURL   string                    `json:"file_url"  binding:"required,max=500"`
	FileSize  int64                     `json:"file_size" binding:"min=0"`
	Status    modelkonten.StatusArtikel `json:"status"    binding:"omitempty,oneof=draft published archived"`
}

// Field pointer: nil = tidak diubah, "" = dikosongkan.
type UpdatePustakaRequest struct {
	Judul     string                    `json:"judul"     binding:"omitempty,min=3,max=500"`
	Kategori  *string                   `json:"kategori"  binding:"omitempty,max=100"`
	Thumbnail *string                   `json:"thumbnail" binding:"omitempty,max=500"`
	FileURL   string                    `json:"file_url"  binding:"max=500"`
	FileSize  *int64                    `json:"file_size" binding:"omitempty,min=0"`
	Status    modelkonten.StatusArtikel `json:"status"    binding:"omitempty,oneof=draft published archived"`
}

type PustakaResponse struct {
	ID        uuid.UUID                 `json:"id"`
	Judul     string                    `json:"judul"`
	Slug      string                    `json:"slug"`
	Kategori  string                    `json:"kategori"`
	Thumbnail string                    `json:"thumbnail,omitempty"`
	Status    modelkonten.StatusArtikel `json:"status"`
	FileURL   string                    `json:"file_url"`
	FileSize  int64                     `json:"file_size"`
	CreatedAt time.Time                 `json:"created_at"`
	UpdatedAt time.Time                 `json:"updated_at"`
}

func toPustakaResponse(d modelkonten.DigitalLibrary) PustakaResponse {
	return PustakaResponse{
		ID:        d.ID,
		Judul:     d.Judul,
		Slug:      d.Slug,
		Kategori:  d.Kategori,
		Thumbnail: d.Thumbnail,
		Status:    d.Status,
		FileURL:   d.FileURL,
		FileSize:  d.FileSize,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

// ── Repository ────────────────────────────────────────────────────────────────

type PustakaFilter struct {
	Search        string
	Status        string
	Kategori      string
	OnlyPublished bool
}

func (r *Repository) ListPustaka(params pagination.Params, f PustakaFilter) ([]modelkonten.DigitalLibrary, int64, error) {
	var list []modelkonten.DigitalLibrary
	var total int64

	q := r.db.Model(&modelkonten.DigitalLibrary{})
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
	err := q.Order("created_at desc").Limit(params.Limit).Offset(params.Offset()).Find(&list).Error
	return list, total, err
}

// ListKategoriPustaka: kategori unik dari dokumen published (kategori bebas/free text).
func (r *Repository) ListKategoriPustaka() ([]string, error) {
	var list []string
	err := r.db.Model(&modelkonten.DigitalLibrary{}).
		Where("status = ? AND kategori <> ''", modelkonten.StatusPublished).
		Distinct("kategori").Order("kategori asc").Pluck("kategori", &list).Error
	return list, err
}

func (r *Repository) FindPustakaByID(id uuid.UUID) (*modelkonten.DigitalLibrary, error) {
	var d modelkonten.DigitalLibrary
	err := r.db.First(&d, "id = ?", id).Error
	return &d, err
}

func (r *Repository) FindPustakaPublikBySlug(slug string) (*modelkonten.DigitalLibrary, error) {
	var d modelkonten.DigitalLibrary
	err := r.db.Where("slug = ? AND status = ?", slug, modelkonten.StatusPublished).First(&d).Error
	return &d, err
}

// SlugPustakaExists juga memeriksa baris soft-delete karena unique index tetap berlaku.
func (r *Repository) SlugPustakaExists(slug string) (bool, error) {
	var count int64
	err := r.db.Unscoped().Model(&modelkonten.DigitalLibrary{}).Where("slug = ?", slug).Count(&count).Error
	return count > 0, err
}

func (r *Repository) CreatePustaka(d *modelkonten.DigitalLibrary) error { return r.db.Create(d).Error }
func (r *Repository) UpdatePustaka(d *modelkonten.DigitalLibrary) error { return r.db.Save(d).Error }
func (r *Repository) DeletePustaka(id uuid.UUID) error {
	return r.db.Delete(&modelkonten.DigitalLibrary{}, "id = ?", id).Error
}

// ── Service ───────────────────────────────────────────────────────────────────

func (s *Service) ListPustaka(params pagination.Params, f PustakaFilter) ([]PustakaResponse, int64, error) {
	list, total, err := s.repo.ListPustaka(params, f)
	if err != nil {
		return nil, 0, apperror.ErrInternal
	}
	result := make([]PustakaResponse, 0, len(list))
	for _, d := range list {
		result = append(result, toPustakaResponse(d))
	}
	return result, total, nil
}

func (s *Service) ListKategoriPustaka() ([]string, error) {
	list, err := s.repo.ListKategoriPustaka()
	if err != nil {
		return nil, apperror.ErrInternal
	}
	return list, nil
}

func (s *Service) GetPustaka(id uuid.UUID) (*PustakaResponse, error) {
	d, err := s.repo.FindPustakaByID(id)
	if err != nil {
		return nil, notFoundOrInternal(err, "dokumen")
	}
	r := toPustakaResponse(*d)
	return &r, nil
}

func (s *Service) GetPustakaPublikBySlug(slug string) (*PustakaResponse, error) {
	d, err := s.repo.FindPustakaPublikBySlug(slug)
	if err != nil {
		return nil, notFoundOrInternal(err, "dokumen")
	}
	r := toPustakaResponse(*d)
	return &r, nil
}

func (s *Service) CreatePustaka(req CreatePustakaRequest) (*PustakaResponse, error) {
	exists, err := s.repo.SlugPustakaExists(req.Slug)
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
	d := &modelkonten.DigitalLibrary{
		Judul:     req.Judul,
		Slug:      req.Slug,
		Kategori:  strings.TrimSpace(req.Kategori),
		Thumbnail: req.Thumbnail,
		Status:    status,
		FileURL:   req.FileURL,
		FileSize:  req.FileSize,
	}
	if err := s.repo.CreatePustaka(d); err != nil {
		return nil, apperror.ErrInternal
	}
	r := toPustakaResponse(*d)
	return &r, nil
}

func (s *Service) UpdatePustaka(id uuid.UUID, req UpdatePustakaRequest) (*PustakaResponse, error) {
	d, err := s.repo.FindPustakaByID(id)
	if err != nil {
		return nil, notFoundOrInternal(err, "dokumen")
	}

	if req.Judul != "" {
		d.Judul = req.Judul
	}
	if req.Kategori != nil {
		d.Kategori = strings.TrimSpace(*req.Kategori)
	}
	if req.Thumbnail != nil {
		d.Thumbnail = *req.Thumbnail
	}
	if req.FileURL != "" {
		d.FileURL = req.FileURL
	}
	if req.FileSize != nil {
		d.FileSize = *req.FileSize
	}
	if req.Status != "" {
		d.Status = req.Status
	}

	if err := s.repo.UpdatePustaka(d); err != nil {
		return nil, apperror.ErrInternal
	}
	r := toPustakaResponse(*d)
	return &r, nil
}

func (s *Service) DeletePustaka(id uuid.UUID) error {
	if _, err := s.repo.FindPustakaByID(id); err != nil {
		return notFoundOrInternal(err, "dokumen")
	}
	if err := s.repo.DeletePustaka(id); err != nil {
		return apperror.ErrInternal
	}
	return nil
}

// ── Handler ───────────────────────────────────────────────────────────────────

// GET /api/admin/pusat-informasi?status=&kategori=&search=&page=&limit=
func (h *Handler) ListPustaka(c *gin.Context) {
	p := pagination.Parse(c)
	list, total, err := h.svc.ListPustaka(p, PustakaFilter{
		Search:   p.Search,
		Status:   c.Query("status"),
		Kategori: c.Query("kategori"),
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.WithPagination(c, "daftar pusat informasi", list, p.Page, p.Limit, total)
}

// GET /api/admin/pusat-informasi/:id
func (h *Handler) GetPustaka(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	d, err := h.svc.GetPustaka(id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "detail dokumen", d)
}

// POST /api/admin/pusat-informasi
func (h *Handler) CreatePustaka(c *gin.Context) {
	var req CreatePustakaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}
	d, err := h.svc.CreatePustaka(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, "dokumen berhasil dibuat", d)
}

// PUT /api/admin/pusat-informasi/:id
func (h *Handler) UpdatePustaka(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req UpdatePustakaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}
	d, err := h.svc.UpdatePustaka(id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "dokumen berhasil diperbarui", d)
}

// DELETE /api/admin/pusat-informasi/:id
func (h *Handler) DeletePustaka(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.DeletePustaka(id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// POST /api/admin/pusat-informasi/upload-thumbnail (field: gambar)
func (h *Handler) UploadThumbnailPustaka(c *gin.Context) {
	h.uploadGambar(c, pustakaPrefix)
}

// POST /api/admin/pusat-informasi/upload-file (field: file, PDF maks 25MB)
func (h *Handler) UploadFilePustaka(c *gin.Context) {
	h.uploadFile(c, pustakaPrefix, pdfSpec)
}

// GET /api/publik/pusat-informasi?kategori=&search=&page=&limit=
// Selalu hanya dokumen published.
func (h *Handler) ListPustakaPublik(c *gin.Context) {
	p := pagination.Parse(c)
	list, total, err := h.svc.ListPustaka(p, PustakaFilter{
		Search:        p.Search,
		Kategori:      c.Query("kategori"),
		OnlyPublished: true,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.WithPagination(c, "daftar pusat informasi", list, p.Page, p.Limit, total)
}

// GET /api/publik/pusat-informasi/kategori
func (h *Handler) ListKategoriPustaka(c *gin.Context) {
	list, err := h.svc.ListKategoriPustaka()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "kategori pusat informasi", list)
}

// GET /api/publik/pusat-informasi/:slug
func (h *Handler) GetPustakaPublikBySlug(c *gin.Context) {
	d, err := h.svc.GetPustakaPublikBySlug(c.Param("slug"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "detail dokumen", d)
}
