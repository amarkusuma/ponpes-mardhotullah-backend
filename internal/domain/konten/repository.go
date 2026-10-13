package konten

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	modelkonten "github.com/simpeg/backend/internal/model/konten"
	"github.com/simpeg/backend/pkg/pagination"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ── Artikel ───────────────────────────────────────────────────────────────────

func (r *Repository) ListArtikel(params pagination.Params, search, status, kategori string) ([]modelkonten.Artikel, int64, error) {
	var list []modelkonten.Artikel
	var total int64

	q := r.db.Model(&modelkonten.Artikel{}).Preload("Penulis")
	if search != "" {
		q = q.Where("judul ILIKE ?", "%"+search+"%")
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if kategori != "" {
		q = q.Where("kategori = ?", kategori)
	}

	q.Count(&total)
	err := q.Order(params.OrderClause()).Limit(params.Limit).Offset(params.Offset()).Find(&list).Error
	return list, total, err
}

func (r *Repository) FindArtikelByID(id uuid.UUID) (*modelkonten.Artikel, error) {
	var a modelkonten.Artikel
	err := r.db.Preload("Penulis").First(&a, "id = ?", id).Error
	return &a, err
}

func (r *Repository) FindArtikelBySlug(slug string) (*modelkonten.Artikel, error) {
	var a modelkonten.Artikel
	err := r.db.Preload("Penulis").Where("slug = ?", slug).First(&a).Error
	return &a, err
}

func (r *Repository) CreateArtikel(a *modelkonten.Artikel) error { return r.db.Create(a).Error }
func (r *Repository) UpdateArtikel(a *modelkonten.Artikel) error { return r.db.Save(a).Error }
func (r *Repository) DeleteArtikel(id uuid.UUID) error {
	return r.db.Delete(&modelkonten.Artikel{}, "id = ?", id).Error
}

// ── Banner ────────────────────────────────────────────────────────────────────

func (r *Repository) ListBanner(onlyActive bool) ([]modelkonten.Banner, error) {
	var list []modelkonten.Banner
	q := r.db.Order("urutan asc")
	if onlyActive {
		q = q.Where("is_active = true")
	}
	err := q.Find(&list).Error
	return list, err
}

func (r *Repository) FindBannerByID(id uuid.UUID) (*modelkonten.Banner, error) {
	var b modelkonten.Banner
	err := r.db.First(&b, "id = ?", id).Error
	return &b, err
}

func (r *Repository) CreateBanner(b *modelkonten.Banner) error { return r.db.Create(b).Error }
func (r *Repository) UpdateBanner(b *modelkonten.Banner) error { return r.db.Save(b).Error }
func (r *Repository) DeleteBanner(id uuid.UUID) error {
	return r.db.Delete(&modelkonten.Banner{}, "id = ?", id).Error
}

// ── Galeri ────────────────────────────────────────────────────────────────────

func (r *Repository) ListGaleri(params pagination.Params, kategori string, onlyPublished bool) ([]modelkonten.Galeri, int64, error) {
	var list []modelkonten.Galeri
	var total int64

	q := r.db.Model(&modelkonten.Galeri{})
	if kategori != "" {
		q = q.Where("kategori = ?", kategori)
	}
	if onlyPublished {
		q = q.Where("is_published = true")
	}

	q.Count(&total)
	err := q.Order("urutan asc, created_at desc").Limit(params.Limit).Offset(params.Offset()).Find(&list).Error
	return list, total, err
}

func (r *Repository) FindGaleriByID(id uuid.UUID) (*modelkonten.Galeri, error) {
	var g modelkonten.Galeri
	err := r.db.First(&g, "id = ?", id).Error
	return &g, err
}

func (r *Repository) CreateGaleri(g *modelkonten.Galeri) error { return r.db.Create(g).Error }
func (r *Repository) UpdateGaleri(g *modelkonten.Galeri) error { return r.db.Save(g).Error }
func (r *Repository) DeleteGaleri(id uuid.UUID) error {
	return r.db.Delete(&modelkonten.Galeri{}, "id = ?", id).Error
}
