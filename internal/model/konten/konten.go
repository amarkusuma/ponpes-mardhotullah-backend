// Package konten berisi model Artikel dan Banner untuk portal web pesantren.
package konten

import (
	"time"

	"github.com/google/uuid"
	"github.com/simpeg/backend/internal/model"
)

// ─────────────────────────────────────────────
//  Enum types
// ─────────────────────────────────────────────

type StatusArtikel string

const (
	StatusDraft     StatusArtikel = "draft"
	StatusPublished StatusArtikel = "published"
	StatusArchived  StatusArtikel = "archived"
)

// ─────────────────────────────────────────────
//  Artikel
// ─────────────────────────────────────────────

type Artikel struct {
	model.Base
	Judul     string        `gorm:"type:varchar(500);not null"            json:"judul"`
	Slug      string        `gorm:"type:varchar(600);uniqueIndex;not null" json:"slug"`
	Konten    string        `gorm:"type:text"                             json:"konten"`
	Kategori  string        `gorm:"type:varchar(100);index"               json:"kategori"`
	PenulisID uuid.UUID     `gorm:"type:uuid;not null;index"              json:"penulis_id"`
	Penulis   model.User    `gorm:"foreignKey:PenulisID"                  json:"penulis,omitempty"`
	Status    StatusArtikel `gorm:"type:varchar(20);default:'draft'"       json:"status"`
	Thumbnail string        `gorm:"type:varchar(500)"                     json:"thumbnail,omitempty"`
}

// ─────────────────────────────────────────────
//  Banner
// ─────────────────────────────────────────────

// Banner untuk carousel / hero section halaman utama.
type Banner struct {
	model.Base
	Judul          string    `gorm:"type:varchar(255);not null" json:"judul"`
	GambarURL      string    `gorm:"type:varchar(500);not null" json:"gambar_url"`
	Link           string    `gorm:"type:varchar(500)"          json:"link,omitempty"`
	Urutan         int       `gorm:"default:0;index"            json:"urutan"`
	IsActive       bool      `gorm:"default:true"               json:"is_active"`
	TanggalMulai   time.Time `json:"tanggal_mulai"`
	TanggalSelesai time.Time `json:"tanggal_selesai"`
}

// ─────────────────────────────────────────────
//  Galeri
// ─────────────────────────────────────────────

type Galeri struct {
	model.Base
	Judul       string `gorm:"type:varchar(255);not null" json:"judul"`
	Deskripsi   string `gorm:"type:text"                  json:"deskripsi,omitempty"`
	GambarURL   string `gorm:"type:varchar(500);not null" json:"gambar_url"`
	Kategori    string `gorm:"type:varchar(100);index"    json:"kategori,omitempty"`
	Urutan      int    `gorm:"default:0;index"            json:"urutan"`
	IsPublished bool   `gorm:"default:true"               json:"is_published"`
}
