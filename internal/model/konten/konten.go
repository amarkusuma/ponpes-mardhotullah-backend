// Package konten berisi model Artikel dan Banner untuk portal web pesantren.
package konten

import (
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
	Views     int64         `gorm:"not null;default:0;index"              json:"views"`
}

// ─────────────────────────────────────────────
//  News (Kabar & Berita)
// ─────────────────────────────────────────────

type KategoriNews string

const (
	KategoriKegiatan   KategoriNews = "kegiatan"
	KategoriPrestasi   KategoriNews = "prestasi"
	KategoriPengumuman KategoriNews = "pengumuman"
	KategoriAgenda     KategoriNews = "agenda"
	KategoriLainnya    KategoriNews = "lainnya"
)

// News memakai status yang sama dengan Artikel (draft, published, archived).
type News struct {
	model.Base
	Judul     string        `gorm:"type:varchar(500);not null"             json:"judul"`
	Slug      string        `gorm:"type:varchar(600);uniqueIndex;not null" json:"slug"`
	Kategori  KategoriNews  `gorm:"type:varchar(30);not null;index"        json:"kategori"`
	Thumbnail string        `gorm:"type:varchar(500)"                      json:"thumbnail,omitempty"`
	Konten    string        `gorm:"type:text"                              json:"konten"`
	Status    StatusArtikel `gorm:"type:varchar(20);default:'draft';index" json:"status"`
	PenulisID uuid.UUID     `gorm:"type:uuid;not null;index"               json:"penulis_id"`
	Penulis   model.User    `gorm:"foreignKey:PenulisID"                   json:"penulis,omitempty"`
}

func (News) TableName() string { return "news" }

// ─────────────────────────────────────────────
//  Pusat Informasi (perpustakaan digital PDF)
// ─────────────────────────────────────────────

type DigitalLibrary struct {
	model.Base
	Judul     string        `gorm:"type:varchar(500);not null"             json:"judul"`
	Slug      string        `gorm:"type:varchar(600);uniqueIndex;not null" json:"slug"`
	Kategori  string        `gorm:"type:varchar(100);index"                json:"kategori"`
	Thumbnail string        `gorm:"type:varchar(500)"                      json:"thumbnail,omitempty"`
	Status    StatusArtikel `gorm:"type:varchar(20);default:'draft';index" json:"status"`
	FileURL   string        `gorm:"type:varchar(500);not null"             json:"file_url"`
	FileSize  int64         `gorm:"not null;default:0"                     json:"file_size"` // byte
}

func (DigitalLibrary) TableName() string { return "digital_library" }

// ─────────────────────────────────────────────
//  Banner
// ─────────────────────────────────────────────

// HalamanBanner: halaman website yang header-nya memakai banner.
type HalamanBanner string

const (
	HalamanHome        HalamanBanner = "home"
	HalamanProfile     HalamanBanner = "profile"
	HalamanGallery     HalamanBanner = "gallery"
	HalamanFAQ         HalamanBanner = "faq"
	HalamanPendaftaran HalamanBanner = "pendaftaran"
	HalamanBerita      HalamanBanner = "berita"
	HalamanArtikel     HalamanBanner = "artikel"
	HalamanPustaka     HalamanBanner = "pusat-informasi"
)

// Banner adalah gambar header (hero) sebuah halaman website. Bila satu halaman
// punya beberapa banner aktif, web menampilkannya bergantian sesuai urutan.
//
// Kolom lama link, tanggal_mulai, tanggal_selesai sengaja tidak dihapus dari
// tabel (AutoMigrate tidak men-drop kolom); kolom tsb nullable dan tidak dipakai.
type Banner struct {
	model.Base
	Judul     string        `gorm:"type:varchar(255);not null"                json:"judul"`
	GambarURL string        `gorm:"type:varchar(500);not null"                json:"gambar_url"`
	Halaman   HalamanBanner `gorm:"type:varchar(30);not null;default:'home';index" json:"halaman"`
	Urutan    int           `gorm:"default:0;index"                           json:"urutan"`
	IsActive  bool          `gorm:"default:true"                              json:"is_active"`
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
