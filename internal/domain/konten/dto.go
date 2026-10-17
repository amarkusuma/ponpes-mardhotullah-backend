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
	Thumbnail *string                   `json:"thumbnail"` // nil = tidak diubah, "" = hapus thumbnail
	Status    modelkonten.StatusArtikel `json:"status"    binding:"omitempty,oneof=draft published archived"`
}

// ── Banner ───────────────────────────────────────────────────────────────────

type CreateBannerRequest struct {
	Judul     string                    `json:"judul"      binding:"max=255"`
	GambarURL string                    `json:"gambar_url" binding:"required,max=500"`
	Halaman   modelkonten.HalamanBanner `json:"halaman"    binding:"required,oneof=home profile gallery faq pendaftaran berita artikel pusat-informasi"`
	Urutan    int                       `json:"urutan"     binding:"min=0"`
	IsActive  *bool                     `json:"is_active"`
}

// Field pointer: nil = tidak diubah (judul boleh dikosongkan, urutan boleh 0).
type UpdateBannerRequest struct {
	Judul     *string                   `json:"judul"      binding:"omitempty,max=255"`
	GambarURL string                    `json:"gambar_url" binding:"max=500"`
	Halaman   modelkonten.HalamanBanner `json:"halaman"    binding:"omitempty,oneof=home profile gallery faq pendaftaran berita artikel pusat-informasi"`
	Urutan    *int                      `json:"urutan"     binding:"omitempty,min=0"`
	IsActive  *bool                     `json:"is_active"`
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
	Views     int64                       `json:"views"`
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

type UploadFileResponse struct {
	URL  string `json:"url"`
	Size int64  `json:"size"` // byte
}
