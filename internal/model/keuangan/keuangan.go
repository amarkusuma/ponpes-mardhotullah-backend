// Package keuangan berisi model Tagihan dan Pembayaran.
package keuangan

import (
	"time"

	"github.com/google/uuid"
	"github.com/simpeg/backend/internal/model"
	"github.com/simpeg/backend/internal/model/akademik"
)

// ─────────────────────────────────────────────
//  Enum types
// ─────────────────────────────────────────────

type StatusTagihan string

const (
	StatusBelumBayar StatusTagihan = "belum_bayar"
	StatusSebagian   StatusTagihan = "sebagian"
	StatusLunas      StatusTagihan = "lunas"
)

type MetodePembayaran string

const (
	MetodeTunai    MetodePembayaran = "tunai"
	MetodeTransfer MetodePembayaran = "transfer"
	MetodeQRIS     MetodePembayaran = "qris"
)

// ─────────────────────────────────────────────
//  Tagihan
// ─────────────────────────────────────────────

// Tagihan dibuat per santri per bulan per jenis (SPP, Asrama, dll.).
// Satu tagihan bisa dilunasi dengan beberapa kali Pembayaran (cicil).
type Tagihan struct {
	model.Base
	SantriID   uuid.UUID     `gorm:"type:uuid;not null;index"         json:"santri_id"`
	Santri     akademik.Santri `gorm:"foreignKey:SantriID"             json:"santri,omitempty"`
	Jenis      string        `gorm:"type:varchar(100);not null"       json:"jenis"` // SPP, Asrama, Kitab...
	Nominal    float64       `gorm:"type:decimal(15,2);not null"      json:"nominal"`
	Bulan      int           `gorm:"not null"                         json:"bulan"` // 1–12
	Tahun      int           `gorm:"not null"                         json:"tahun"`
	Status     StatusTagihan `gorm:"type:varchar(20);default:'belum_bayar'" json:"status"`
	Pembayaran []Pembayaran  `gorm:"foreignKey:TagihanID"             json:"pembayaran,omitempty"`
}

// ─────────────────────────────────────────────
//  Pembayaran
// ─────────────────────────────────────────────

// Pembayaran mencatat setiap transaksi pembayaran tagihan.
type Pembayaran struct {
	model.Base
	TagihanID    uuid.UUID        `gorm:"type:uuid;not null;index"     json:"tagihan_id"`
	Tagihan      Tagihan          `gorm:"foreignKey:TagihanID"         json:"tagihan,omitempty"`
	JumlahBayar  float64          `gorm:"type:decimal(15,2);not null"  json:"jumlah_bayar"`
	TanggalBayar time.Time        `json:"tanggal_bayar"`
	Metode       MetodePembayaran `gorm:"type:varchar(20)"             json:"metode"`
	Keterangan   string           `gorm:"type:text"                    json:"keterangan,omitempty"`
}
