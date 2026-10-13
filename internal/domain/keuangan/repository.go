package keuangan

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/simpeg/backend/internal/model/akademik"
	modelkeuangan "github.com/simpeg/backend/internal/model/keuangan"
	"github.com/simpeg/backend/pkg/pagination"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ── Tagihan ───────────────────────────────────────────────────────────────────

func (r *Repository) ListTagihan(params pagination.Params, filter ListTagihanFilter) ([]modelkeuangan.Tagihan, int64, error) {
	var list []modelkeuangan.Tagihan
	var total int64

	q := r.db.Model(&modelkeuangan.Tagihan{}).Preload("Santri.User").Preload("Pembayaran")

	if filter.SantriID != nil {
		q = q.Where("santri_id = ?", *filter.SantriID)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.Bulan > 0 {
		q = q.Where("bulan = ?", filter.Bulan)
	}
	if filter.Tahun > 0 {
		q = q.Where("tahun = ?", filter.Tahun)
	}

	q.Count(&total)
	err := q.Order(params.OrderClause()).Limit(params.Limit).Offset(params.Offset()).Find(&list).Error
	return list, total, err
}

func (r *Repository) FindTagihanByID(id uuid.UUID) (*modelkeuangan.Tagihan, error) {
	var t modelkeuangan.Tagihan
	err := r.db.Preload("Santri.User").Preload("Pembayaran").
		First(&t, "id = ?", id).Error
	return &t, err
}

func (r *Repository) CreateTagihan(t *modelkeuangan.Tagihan) error {
	return r.db.Create(t).Error
}

func (r *Repository) UpdateTagihanStatus(id uuid.UUID, status modelkeuangan.StatusTagihan) error {
	return r.db.Model(&modelkeuangan.Tagihan{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// ── Pembayaran ────────────────────────────────────────────────────────────────

func (r *Repository) ListPembayaranByTagihan(tagihanID uuid.UUID) ([]modelkeuangan.Pembayaran, error) {
	var list []modelkeuangan.Pembayaran
	err := r.db.Where("tagihan_id = ?", tagihanID).
		Order("tanggal_bayar desc").
		Find(&list).Error
	return list, err
}

func (r *Repository) CreatePembayaran(p *modelkeuangan.Pembayaran) error {
	return r.db.Create(p).Error
}

func (r *Repository) SumPembayaran(tagihanID uuid.UUID) (float64, error) {
	var sum float64
	err := r.db.Model(&modelkeuangan.Pembayaran{}).
		Where("tagihan_id = ?", tagihanID).
		Select("COALESCE(SUM(jumlah_bayar), 0)").
		Scan(&sum).Error
	return sum, err
}

// ── Untuk Santri/Wali ─────────────────────────────────────────────────────────

func (r *Repository) FindSantriByUserID(userID uuid.UUID) (*akademik.Santri, error) {
	var s akademik.Santri
	err := r.db.Where("user_id = ?", userID).First(&s).Error
	return &s, err
}

// ListRiwayatPembayaranBySantri mengembalikan semua pembayaran milik santri.
func (r *Repository) ListRiwayatPembayaranBySantri(santriID uuid.UUID) ([]modelkeuangan.Pembayaran, error) {
	var list []modelkeuangan.Pembayaran
	err := r.db.
		Joins("JOIN tagihans t ON t.id = pembayarans.tagihan_id AND t.santri_id = ?", santriID).
		Preload("Tagihan").
		Order("pembayarans.tanggal_bayar desc").
		Find(&list).Error
	return list, err
}
