package keuangan

import (
	"time"

	"github.com/google/uuid"
	modelkeuangan "github.com/simpeg/backend/internal/model/keuangan"
)

// ── Request DTOs ────────────────────────────────────────────────────────────

type CreateTagihanRequest struct {
	SantriID uuid.UUID `json:"santri_id" binding:"required"`
	Jenis    string    `json:"jenis"     binding:"required"`
	Nominal  float64   `json:"nominal"   binding:"required,gt=0"`
	Bulan    int       `json:"bulan"     binding:"required,min=1,max=12"`
	Tahun    int       `json:"tahun"     binding:"required"`
}

type CreatePembayaranRequest struct {
	TagihanID    uuid.UUID                      `json:"tagihan_id"    binding:"required"`
	JumlahBayar  float64                        `json:"jumlah_bayar"  binding:"required,gt=0"`
	TanggalBayar time.Time                      `json:"tanggal_bayar" binding:"required"`
	Metode       modelkeuangan.MetodePembayaran `json:"metode"        binding:"required,oneof=tunai transfer qris"`
	Keterangan   string                         `json:"keterangan"`
}

type ListTagihanFilter struct {
	SantriID *uuid.UUID
	Status   string
	Bulan    int
	Tahun    int
}

// ── Response DTOs ───────────────────────────────────────────────────────────

type TagihanResponse struct {
	ID           uuid.UUID                    `json:"id"`
	SantriID     uuid.UUID                    `json:"santri_id"`
	NamaSantri   string                       `json:"nama_santri"`
	NIK          string                       `json:"nik"`
	Jenis        string                       `json:"jenis"`
	Nominal      float64                      `json:"nominal"`
	Bulan        int                          `json:"bulan"`
	Tahun        int                          `json:"tahun"`
	Status       modelkeuangan.StatusTagihan  `json:"status"`
	TotalDibayar float64                      `json:"total_dibayar"`
	Sisa         float64                      `json:"sisa"`
	CreatedAt    time.Time                    `json:"created_at"`
}

type PembayaranResponse struct {
	ID           uuid.UUID                      `json:"id"`
	TagihanID    uuid.UUID                      `json:"tagihan_id"`
	JumlahBayar  float64                        `json:"jumlah_bayar"`
	TanggalBayar time.Time                      `json:"tanggal_bayar"`
	Metode       modelkeuangan.MetodePembayaran `json:"metode"`
	Keterangan   string                         `json:"keterangan,omitempty"`
	CreatedAt    time.Time                      `json:"created_at"`
}
