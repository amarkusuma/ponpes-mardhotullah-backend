package konten

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	modelkonten "github.com/simpeg/backend/internal/model/konten"
	"github.com/simpeg/backend/pkg/apperror"
	"github.com/simpeg/backend/pkg/pagination"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ── Artikel ───────────────────────────────────────────────────────────────────

func (s *Service) ListArtikel(params pagination.Params, status, kategori string) ([]ArtikelResponse, int64, error) {
	list, total, err := s.repo.ListArtikel(params, params.Search, status, kategori)
	if err != nil {
		return nil, 0, apperror.ErrInternal
	}

	result := make([]ArtikelResponse, 0, len(list))
	for _, a := range list {
		result = append(result, toArtikelResponse(a))
	}
	return result, total, nil
}

func (s *Service) GetArtikel(id uuid.UUID) (*ArtikelResponse, error) {
	a, err := s.repo.FindArtikelByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.ErrNotFoundMsg("artikel")
		}
		return nil, apperror.ErrInternal
	}
	r := toArtikelResponse(*a)
	return &r, nil
}

func (s *Service) GetArtikelBySlug(slug string) (*ArtikelResponse, error) {
	a, err := s.repo.FindArtikelBySlug(slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.ErrNotFoundMsg("artikel")
		}
		return nil, apperror.ErrInternal
	}
	r := toArtikelResponse(*a)
	return &r, nil
}

func (s *Service) CreateArtikel(penulisID uuid.UUID, req CreateArtikelRequest) (*ArtikelResponse, error) {
	// Cek slug unik
	if _, err := s.repo.FindArtikelBySlug(req.Slug); err == nil {
		return nil, apperror.ErrConflictMsg("slug")
	}

	status := req.Status
	if status == "" {
		status = modelkonten.StatusDraft
	}

	a := &modelkonten.Artikel{
		Judul:     req.Judul,
		Slug:      req.Slug,
		Konten:    req.Konten,
		Kategori:  req.Kategori,
		Thumbnail: req.Thumbnail,
		PenulisID: penulisID,
		Status:    status,
	}
	if err := s.repo.CreateArtikel(a); err != nil {
		return nil, apperror.ErrInternal
	}

	r, err := s.GetArtikel(a.ID)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (s *Service) UpdateArtikel(id uuid.UUID, req UpdateArtikelRequest) (*ArtikelResponse, error) {
	a, err := s.repo.FindArtikelByID(id)
	if err != nil {
		return nil, apperror.ErrNotFoundMsg("artikel")
	}

	if req.Judul != "" {
		a.Judul = req.Judul
	}
	if req.Konten != "" {
		a.Konten = req.Konten
	}
	if req.Kategori != "" {
		a.Kategori = req.Kategori
	}
	if req.Thumbnail != "" {
		a.Thumbnail = req.Thumbnail
	}
	if req.Status != "" {
		a.Status = req.Status
	}

	if err := s.repo.UpdateArtikel(a); err != nil {
		return nil, apperror.ErrInternal
	}

	r := toArtikelResponse(*a)
	return &r, nil
}

func (s *Service) DeleteArtikel(id uuid.UUID) error {
	if _, err := s.repo.FindArtikelByID(id); err != nil {
		return apperror.ErrNotFoundMsg("artikel")
	}
	return s.repo.DeleteArtikel(id)
}

// ── Banner ────────────────────────────────────────────────────────────────────

func (s *Service) ListBanner(onlyActive bool) ([]modelkonten.Banner, error) {
	return s.repo.ListBanner(onlyActive)
}

func (s *Service) CreateBanner(req CreateBannerRequest) (*modelkonten.Banner, error) {
	b := &modelkonten.Banner{
		Judul:          req.Judul,
		GambarURL:      req.GambarURL,
		Link:           req.Link,
		Urutan:         req.Urutan,
		IsActive:       req.IsActive,
		TanggalMulai:   req.TanggalMulai,
		TanggalSelesai: req.TanggalSelesai,
	}
	if err := s.repo.CreateBanner(b); err != nil {
		return nil, apperror.ErrInternal
	}
	return b, nil
}

func (s *Service) UpdateBanner(id uuid.UUID, req UpdateBannerRequest) (*modelkonten.Banner, error) {
	b, err := s.repo.FindBannerByID(id)
	if err != nil {
		return nil, apperror.ErrNotFoundMsg("banner")
	}

	if req.Judul != "" {
		b.Judul = req.Judul
	}
	if req.GambarURL != "" {
		b.GambarURL = req.GambarURL
	}
	if req.Link != "" {
		b.Link = req.Link
	}
	if req.Urutan > 0 {
		b.Urutan = req.Urutan
	}
	if req.IsActive != nil {
		b.IsActive = *req.IsActive
	}
	if !req.TanggalMulai.IsZero() {
		b.TanggalMulai = req.TanggalMulai
	}
	if !req.TanggalSelesai.IsZero() {
		b.TanggalSelesai = req.TanggalSelesai
	}

	if err := s.repo.UpdateBanner(b); err != nil {
		return nil, apperror.ErrInternal
	}
	return b, nil
}

func (s *Service) DeleteBanner(id uuid.UUID) error {
	if _, err := s.repo.FindBannerByID(id); err != nil {
		return apperror.ErrNotFoundMsg("banner")
	}
	return s.repo.DeleteBanner(id)
}

// ── Galeri ────────────────────────────────────────────────────────────────────

func (s *Service) ListGaleri(params pagination.Params, kategori string, onlyPublished bool) ([]GaleriResponse, int64, error) {
	list, total, err := s.repo.ListGaleri(params, kategori, onlyPublished)
	if err != nil {
		return nil, 0, apperror.ErrInternal
	}
	result := make([]GaleriResponse, 0, len(list))
	for _, g := range list {
		result = append(result, toGaleriResponse(g))
	}
	return result, total, nil
}

func (s *Service) GetGaleri(id uuid.UUID) (*GaleriResponse, error) {
	g, err := s.repo.FindGaleriByID(id)
	if err != nil {
		return nil, apperror.ErrNotFoundMsg("galeri")
	}
	r := toGaleriResponse(*g)
	return &r, nil
}

func (s *Service) CreateGaleri(req CreateGaleriRequest) (*GaleriResponse, error) {
	isPublished := true
	if req.IsPublished != nil {
		isPublished = *req.IsPublished
	}
	g := &modelkonten.Galeri{
		Judul:       req.Judul,
		Deskripsi:   req.Deskripsi,
		GambarURL:   req.GambarURL,
		Kategori:    req.Kategori,
		Urutan:      req.Urutan,
		IsPublished: isPublished,
	}
	if err := s.repo.CreateGaleri(g); err != nil {
		return nil, apperror.ErrInternal
	}
	r := toGaleriResponse(*g)
	return &r, nil
}

func (s *Service) UpdateGaleri(id uuid.UUID, req UpdateGaleriRequest) (*GaleriResponse, error) {
	g, err := s.repo.FindGaleriByID(id)
	if err != nil {
		return nil, apperror.ErrNotFoundMsg("galeri")
	}
	if req.Judul != "" {
		g.Judul = req.Judul
	}
	if req.Deskripsi != "" {
		g.Deskripsi = req.Deskripsi
	}
	if req.GambarURL != "" {
		g.GambarURL = req.GambarURL
	}
	if req.Kategori != "" {
		g.Kategori = req.Kategori
	}
	if req.Urutan > 0 {
		g.Urutan = req.Urutan
	}
	if req.IsPublished != nil {
		g.IsPublished = *req.IsPublished
	}
	if err := s.repo.UpdateGaleri(g); err != nil {
		return nil, apperror.ErrInternal
	}
	r := toGaleriResponse(*g)
	return &r, nil
}

func (s *Service) DeleteGaleri(id uuid.UUID) error {
	if _, err := s.repo.FindGaleriByID(id); err != nil {
		return apperror.ErrNotFoundMsg("galeri")
	}
	return s.repo.DeleteGaleri(id)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func toArtikelResponse(a modelkonten.Artikel) ArtikelResponse {
	penulis := ""
	if a.Penulis.Name != "" {
		penulis = a.Penulis.Name
	}
	return ArtikelResponse{
		ID:        a.ID,
		Judul:     a.Judul,
		Slug:      a.Slug,
		Konten:    a.Konten,
		Kategori:  a.Kategori,
		Thumbnail: a.Thumbnail,
		Status:    a.Status,
		Penulis:   penulis,
		CreatedAt: a.CreatedAt,
	}
}

func toGaleriResponse(g modelkonten.Galeri) GaleriResponse {
	return GaleriResponse{
		ID:          g.ID,
		Judul:       g.Judul,
		Deskripsi:   g.Deskripsi,
		GambarURL:   g.GambarURL,
		Kategori:    g.Kategori,
		Urutan:      g.Urutan,
		IsPublished: g.IsPublished,
		CreatedAt:   g.CreatedAt,
	}
}
