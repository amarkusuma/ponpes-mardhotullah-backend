// Package database menyediakan fungsi migrasi dan seeding.
package database

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/simpeg/backend/internal/model"
	"github.com/simpeg/backend/internal/model/akademik"
	"github.com/simpeg/backend/internal/model/jadwal"
	"github.com/simpeg/backend/internal/model/keuangan"
	"github.com/simpeg/backend/internal/model/konten"
	"github.com/simpeg/backend/internal/model/penilaian"
)

// Migrate menjalankan AutoMigrate untuk seluruh model sistem.
// Urutan penting: tabel yang direferensikan harus dibuat lebih dulu.
func Migrate(db *gorm.DB) error {
	// Aktifkan extension uuid-ossp agar uuid_generate_v4() tersedia.
	db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`)

	if err := db.AutoMigrate(
		// ── RBAC ──────────────────────────────────────────────────────
		&model.Permission{},
		&model.Role{},
		&model.User{},

		// ── Akademik ──────────────────────────────────────────────────
		&akademik.TahunAjaran{},
		&akademik.Guru{},
		&akademik.Wali{},
		&akademik.Santri{},
		&akademik.Kelas{},
		&akademik.KelasSantri{},
		&akademik.MataPelajaran{},

		// ── Jadwal ────────────────────────────────────────────────────
		&jadwal.JadwalPelajaran{},

		// ── Penilaian ─────────────────────────────────────────────────
		&penilaian.JenisPenilaian{},
		&penilaian.Penilaian{},
		&penilaian.Rapor{},

		// ── Keuangan ──────────────────────────────────────────────────
		&keuangan.Tagihan{},
		&keuangan.Pembayaran{},

		// ── Konten ────────────────────────────────────────────────────
		&konten.Artikel{},
		&konten.News{},
		&konten.DigitalLibrary{},
		&konten.Banner{},
		&konten.Galeri{},
	); err != nil {
		return fmt.Errorf("auto migrate gagal: %w", err)
	}

	// Unique constraint: satu tagihan per santri per bulan per tahun per jenis.
	// Dijalankan setelah AutoMigrate agar tabel sudah ada.
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_tagihan_unique
		ON tagihans(santri_id, bulan, tahun, jenis)
		WHERE deleted_at IS NULL`)

	return nil
}
