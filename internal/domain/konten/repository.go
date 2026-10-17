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

// artikelListColumns: kolom untuk daftar ringkas (tanpa konten agar response ringan).
var artikelListColumns = []string{
	"id", "created_at", "updated_at", "judul", "slug", "kategori",
	"penulis_id", "status", "thumbnail", "views",
}

func (r *Repository) publishedArtikel() *gorm.DB {
	return r.db.Model(&modelkonten.Artikel{}).
		Where("status = ?", modelkonten.StatusPublished)
}

func (r *Repository) ListArtikelPublik(params pagination.Params, kategori string, populer bool) ([]modelkonten.Artikel, int64, error) {
	var list []modelkonten.Artikel
	var total int64

	q := r.publishedArtikel()
	if params.Search != "" {
		q = q.Where("judul ILIKE ?", "%"+params.Search+"%")
	}
	if kategori != "" {
		q = q.Where("kategori = ?", kategori)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := "created_at desc"
	if populer {
		order = "views desc, created_at desc"
	}
	err := q.Preload("Penulis").
		Order(order).Limit(params.Limit).Offset(params.Offset()).Find(&list).Error
	return list, total, err
}

func (r *Repository) FindArtikelPublikBySlug(slug string) (*modelkonten.Artikel, error) {
	var a modelkonten.Artikel
	err := r.publishedArtikel().Preload("Penulis").Where("slug = ?", slug).First(&a).Error
	return &a, err
}

func (r *Repository) ListArtikelTerkait(excludeID uuid.UUID, kategori string, limit int) ([]modelkonten.Artikel, error) {
	var list []modelkonten.Artikel
	if kategori != "" {
		err := r.publishedArtikel().Select(artikelListColumns).Preload("Penulis").
			Where("id <> ? AND kategori = ?", excludeID, kategori).
			Order("created_at desc").Limit(limit).Find(&list).Error
		if err != nil {
			return nil, err
		}
	}
	if len(list) >= limit {
		return list, nil
	}

	exclude := []uuid.UUID{excludeID}
	for _, a := range list {
		exclude = append(exclude, a.ID)
	}
	var lainnya []modelkonten.Artikel
	err := r.publishedArtikel().Select(artikelListColumns).Preload("Penulis").
		Where("id NOT IN ?", exclude).
		Order("created_at desc").Limit(limit - len(list)).Find(&lainnya).Error
	return append(list, lainnya...), err
}

// IncrementArtikelViews menaikkan views tanpa menyentuh updated_at.
// Mengembalikan false bila artikel published dengan slug tsb tidak ada.
func (r *Repository) IncrementArtikelViews(slug string) (bool, error) {
	res := r.publishedArtikel().Where("slug = ?", slug).
		UpdateColumn("views", gorm.Expr("views + 1"))
	return res.RowsAffected > 0, res.Error
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

func (r *Repository) ListBanner(halaman string, onlyActive bool) ([]modelkonten.Banner, error) {
	var list []modelkonten.Banner
	q := r.db.Order("urutan asc, created_at asc")
	if halaman != "" {
		q = q.Where("halaman = ?", halaman)
	}
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

// GORM mengganti is_active=false dengan default kolom (true) saat Create,
// jadi banner nonaktif disimpan dalam transaksi lalu di-update ke false.
func (r *Repository) CreateBanner(b *modelkonten.Banner) error {
	isActive := b.IsActive
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(b).Error; err != nil {
			return err
		}
		if !isActive {
			b.IsActive = false
			return tx.Model(b).Update("is_active", false).Error
		}
		return nil
	})
}
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
