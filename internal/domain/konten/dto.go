package konten

import (
	"time"

	"github.com/google/uuid"
	modelkonten "github.com/simpeg/backend/internal/model/konten"
)

// ── Artikel ──────────────────────────────────────────────────────────────────

type CreateArtikelRequest struct {
	Judul     string                      `json:"judul"     binding:"required,min=5"`
	Slug      string                      `json:"slug"      binding:"required"`
	Konten    string                      `json:"konten"    binding:"required"`
	Kategori  string                      `json:"kategori"`
	Thumbnail string                      `json:"thumbnail"`
	Status    modelkonten.StatusArtikel   `json:"status"    binding:"omitempty,oneof=draft published archived"`
}

type UpdateArtikelRequest struct {
	Judul     string                    `json:"judul"`
	Konten    string                    `json:"konten"`
	Kategori  string                    `json:"kategori"`
	Thumbnail string                    `json:"thumbnail"`
	Status    modelkonten.StatusArtikel `json:"status"    binding:"omitempty,oneof=draft published archived"`
}

// ── Banner ───────────────────────────────────────────────────────────────────

type CreateBannerRequest struct {
	Judul          string    `json:"judul"           binding:"required"`
	GambarURL      string    `json:"gambar_url"      binding:"required,url"`
	Link           string    `json:"link"`
	Urutan         int       `json:"urutan"`
	IsActive       bool      `json:"is_active"`
	TanggalMulai   time.Time `json:"tanggal_mulai"   binding:"required"`
	TanggalSelesai time.Time `json:"tanggal_selesai" binding:"required"`
}

type UpdateBannerRequest struct {
	Judul          string    `json:"judul"`
	GambarURL      string    `json:"gambar_url"`
	Link           string    `json:"link"`
	Urutan         int       `json:"urutan"`
	IsActive       *bool     `json:"is_active"`
	TanggalMulai   time.Time `json:"tanggal_mulai"`
	TanggalSelesai time.Time `json:"tanggal_selesai"`
}

type ArtikelResponse struct {
	ID        uuid.UUID                   `json:"id"`
	Judul     string                      `json:"judul"`
	Slug      string                      `json:"slug"`
	Konten    string                      `json:"konten,omitempty"`
	Kategori  string                      `json:"kategori"`
	Thumbnail string                      `json:"thumbnail,omitempty"`
	Status    modelkonten.StatusArtikel   `json:"status"`
	Penulis   string                      `json:"penulis"`
	CreatedAt time.Time                   `json:"created_at"`
}

// ── Galeri ───────────────────────────────────────────────────────────────────

type CreateGaleriRequest struct {
	Judul       string `json:"judul"      binding:"required,min=2"`
	Deskripsi   string `json:"deskripsi"`
	GambarURL   string `json:"gambar_url" binding:"required"`
	Kategori    string `json:"kategori"`
	Urutan      int    `json:"urutan"`
	IsPublished *bool  `json:"is_published"`
}

type UpdateGaleriRequest struct {
	Judul       string `json:"judul"`
	Deskripsi   string `json:"deskripsi"`
	GambarURL   string `json:"gambar_url"`
	Kategori    string `json:"kategori"`
	Urutan      int    `json:"urutan"`
	IsPublished *bool  `json:"is_published"`
}

type GaleriResponse struct {
	ID          uuid.UUID `json:"id"`
	Judul       string    `json:"judul"`
	Deskripsi   string    `json:"deskripsi,omitempty"`
	GambarURL   string    `json:"gambar_url"`
	Kategori    string    `json:"kategori,omitempty"`
	Urutan      int       `json:"urutan"`
	IsPublished bool      `json:"is_published"`
	CreatedAt   time.Time `json:"created_at"`
}

type UploadGambarResponse struct {
	URL string `json:"url"`
}
