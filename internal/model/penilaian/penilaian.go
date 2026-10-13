// Package penilaian berisi model untuk sistem penilaian dan rapor santri.
package penilaian

import (
	"time"

	"github.com/google/uuid"
	"github.com/simpeg/backend/internal/model"
	"github.com/simpeg/backend/internal/model/akademik"
)

// ─────────────────────────────────────────────
//  JenisPenilaian  (reference table)
// ─────────────────────────────────────────────

// JenisPenilaian adalah lookup table: UTS, UAS, Harian, Praktik, dsb.
type JenisPenilaian struct {
	model.Base
	Nama string `gorm:"type:varchar(100);uniqueIndex;not null" json:"nama"`
}

// ─────────────────────────────────────────────
//  Penilaian
// ─────────────────────────────────────────────

// Penilaian menyimpan nilai per santri per mata pelajaran per jenis penilaian.
// Unique constraint direkomendasikan: (santri_id, mapel_id, jenis_penilaian_id, semester, tahun_ajaran_id).
type Penilaian struct {
	model.Base
	SantriID         uuid.UUID              `gorm:"type:uuid;not null;index"    json:"santri_id"`
	Santri           akademik.Santri        `gorm:"foreignKey:SantriID"         json:"santri,omitempty"`
	MapelID          uuid.UUID              `gorm:"type:uuid;not null;index"    json:"mapel_id"`
	MataPelajaran    akademik.MataPelajaran `gorm:"foreignKey:MapelID"          json:"mata_pelajaran,omitempty"`
	KelasID          uuid.UUID              `gorm:"type:uuid;not null;index"    json:"kelas_id"`
	Kelas            akademik.Kelas         `gorm:"foreignKey:KelasID"          json:"kelas,omitempty"`
	JenisPenilaianID uuid.UUID              `gorm:"type:uuid;not null;index"    json:"jenis_penilaian_id"`
	JenisPenilaian   JenisPenilaian         `gorm:"foreignKey:JenisPenilaianID" json:"jenis_penilaian,omitempty"`
	Nilai            float64                `gorm:"type:decimal(5,2);not null"  json:"nilai"`
	Semester         int                    `gorm:"not null"                    json:"semester"` // 1 atau 2
	TahunAjaranID    uuid.UUID              `gorm:"type:uuid;not null;index"    json:"tahun_ajaran_id"`
	TahunAjaran      akademik.TahunAjaran   `gorm:"foreignKey:TahunAjaranID"    json:"tahun_ajaran,omitempty"`
}

// ─────────────────────────────────────────────
//  Rapor
// ─────────────────────────────────────────────

// Rapor menyimpan file rapor PDF yang sudah di-generate per santri per semester.
type Rapor struct {
	model.Base
	SantriID      uuid.UUID            `gorm:"type:uuid;not null;index"  json:"santri_id"`
	Santri        akademik.Santri      `gorm:"foreignKey:SantriID"       json:"santri,omitempty"`
	Semester      int                  `gorm:"not null"                  json:"semester"`
	TahunAjaranID uuid.UUID            `gorm:"type:uuid;not null;index"  json:"tahun_ajaran_id"`
	TahunAjaran   akademik.TahunAjaran `gorm:"foreignKey:TahunAjaranID"  json:"tahun_ajaran,omitempty"`
	FilePath      string               `gorm:"type:varchar(500)"         json:"file_path"`
	GeneratedAt   time.Time            `json:"generated_at"`
}
