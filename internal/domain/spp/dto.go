package spp

import (
	"time"

	"github.com/google/uuid"
	modelkeuangan "github.com/simpeg/backend/internal/model/keuangan"
)

// ── Request DTOs ─────────────────────────────────────────────────────────────

type GenerateSPPRequest struct {
	Bulan   int     `json:"bulan"    binding:"required,min=1,max=12"`
	Tahun   int     `json:"tahun"    binding:"required,min=2000"`
	Nominal float64 `json:"nominal"  binding:"required,gt=0"`
	Jenis   string  `json:"jenis"` // default "SPP" jika kosong
}

type BayarTagihanRequest struct {
	JumlahBayar  float64                        `json:"jumlah_bayar"  binding:"required,gt=0"`
	Metode       modelkeuangan.MetodePembayaran `json:"metode"        binding:"required,oneof=tunai transfer qris"`
	TanggalBayar time.Time                      `json:"tanggal_bayar" binding:"required"`
	Keterangan   string                         `json:"keterangan"`
}

type ListSPPFilter struct {
	Search  string
	Status  string
	Bulan   int
	Tahun   int
	KelasID *uuid.UUID
	Jenis   string
}

// ── Response DTOs ─────────────────────────────────────────────────────────────

type TagihanSPPResponse struct {
	ID           uuid.UUID                   `json:"id"`
	SantriID     uuid.UUID                   `json:"santri_id"`
	NamaSantri   string                      `json:"nama_santri"`
	NIK          string                      `json:"nik"`
	KelasNama    string                      `json:"kelas_nama"`
	Jenis        string                      `json:"jenis"`
	Nominal      float64                     `json:"nominal"`
	Bulan        int                         `json:"bulan"`
	Tahun        int                         `json:"tahun"`
	Status       modelkeuangan.StatusTagihan `json:"status"`
	TotalDibayar float64                     `json:"total_dibayar"`
	Sisa         float64                     `json:"sisa"`
	CreatedAt    time.Time                   `json:"created_at"`
}

type PembayaranResponse struct {
	ID           uuid.UUID                      `json:"id"`
	TagihanID    uuid.UUID                      `json:"tagihan_id"`
	JumlahBayar  float64                        `json:"jumlah_bayar"`
	Metode       modelkeuangan.MetodePembayaran `json:"metode"`
	TanggalBayar time.Time                      `json:"tanggal_bayar"`
	Keterangan   string                         `json:"keterangan,omitempty"`
	CreatedAt    time.Time                      `json:"created_at"`
}

type TagihanDetailResponse struct {
	TagihanSPPResponse
	Pembayaran []PembayaranResponse `json:"pembayaran"`
}

type GenerateResult struct {
	Generated int `json:"generated"`
	Skipped   int `json:"skipped"`
	Total     int `json:"total"`
}

type RekapBulanan struct {
	Bulan         int     `json:"bulan"`
	Tahun         int     `json:"tahun"`
	TotalTagihan  float64 `json:"total_tagihan"`
	TotalMasuk    float64 `json:"total_masuk"`
	JumlahTagihan int64   `json:"jumlah_tagihan"`
	JumlahLunas   int64   `json:"jumlah_lunas"`
	JumlahBelum   int64   `json:"jumlah_belum"`
	JumlahCicilan int64   `json:"jumlah_cicilan"`
}

type UnpaidItem struct {
	TagihanID  uuid.UUID `json:"tagihan_id"`
	SantriID   uuid.UUID `json:"santri_id"`
	NamaSantri string    `json:"nama_santri"`
	NIK        string    `json:"nik"`
	KelasNama  string    `json:"kelas_nama"`
	Bulan      int       `json:"bulan"`
	Tahun      int       `json:"tahun"`
	Nominal    float64   `json:"nominal"`
	Sisa       float64   `json:"sisa"`
	Status     string    `json:"status"`
}
