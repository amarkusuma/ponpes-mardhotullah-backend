// Package akademik berisi model untuk domain akademik pesantren:
// TahunAjaran, Guru, Santri, Wali, Kelas, KelasSantri, MataPelajaran.
package akademik

import (
	"time"

	"github.com/google/uuid"
	"github.com/simpeg/backend/internal/model"
)

// ─────────────────────────────────────────────
//  Enum types
// ─────────────────────────────────────────────

type SantriStatus string

const (
	StatusAktif   SantriStatus = "aktif"
	StatusAlumni  SantriStatus = "alumni"
	StatusDropout SantriStatus = "dropout"
	StatusKeluar  SantriStatus = "keluar"
)

// ─────────────────────────────────────────────
//  TahunAjaran
// ─────────────────────────────────────────────

type TahunAjaran struct {
	model.Base
	TahunMulai   int  `gorm:"not null"      json:"tahun_mulai"`
	TahunSelesai int  `gorm:"not null"      json:"tahun_selesai"`
	IsActive     bool `gorm:"default:false" json:"is_active"`
}

// ─────────────────────────────────────────────
//  Guru
// ─────────────────────────────────────────────

type Guru struct {
	model.Base
	UserID       uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex"               json:"user_id"`
	User         model.User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	NIK          string     `gorm:"type:varchar(50);uniqueIndex"                 json:"nik"`
	Spesialisasi string     `gorm:"type:varchar(255)"                            json:"spesialisasi"`
	Foto         string     `gorm:"type:varchar(500)"                            json:"foto,omitempty"`
}

// ─────────────────────────────────────────────
//  Wali
// ─────────────────────────────────────────────

type Wali struct {
	model.Base
	UserID    uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex"               json:"user_id"`
	User      model.User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	NIK       string     `gorm:"type:varchar(50);uniqueIndex"                 json:"nik"`
	Hubungan  string     `gorm:"type:varchar(100)"                            json:"hubungan"`
	Telepon   string     `gorm:"type:varchar(20)"                             json:"telepon"`
	Pekerjaan string     `gorm:"type:varchar(255)"                            json:"pekerjaan"`
	Alamat    string     `gorm:"type:text"                                    json:"alamat"`
	Santri    []Santri   `gorm:"many2many:santri_wali"                        json:"santri,omitempty"`
}

// ─────────────────────────────────────────────
//  Santri
// ─────────────────────────────────────────────

type Santri struct {
	model.Base
	UserID       uuid.UUID    `gorm:"type:uuid;not null;uniqueIndex"                json:"user_id"`
	User         model.User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"  json:"user,omitempty"`
	NIK          string       `gorm:"type:varchar(50);uniqueIndex"                   json:"nik"`
	TempatLahir  string       `gorm:"type:varchar(255)"                             json:"tempat_lahir"`
	TanggalLahir time.Time    `gorm:"type:date"                                     json:"tanggal_lahir"`
	Alamat       string       `gorm:"type:text"                                     json:"alamat"`
	Foto         string       `gorm:"type:varchar(500)"                             json:"foto,omitempty"`
	Status       SantriStatus `gorm:"type:varchar(20);default:'aktif'"              json:"status"`
	Wali         []Wali       `gorm:"many2many:santri_wali"                         json:"wali,omitempty"`
}

// ─────────────────────────────────────────────
//  Kelas & KelasSantri
// ─────────────────────────────────────────────

type Kelas struct {
	model.Base
	NamaKelas     string      `gorm:"type:varchar(100);not null"   json:"nama_kelas"`
	Tingkat       int         `gorm:"not null"                     json:"tingkat"`
	TahunAjaranID uuid.UUID   `gorm:"type:uuid;not null;index"     json:"tahun_ajaran_id"`
	TahunAjaran   TahunAjaran `gorm:"foreignKey:TahunAjaranID"     json:"tahun_ajaran,omitempty"`
	WaliKelasID   *uuid.UUID  `gorm:"type:uuid;index"              json:"wali_kelas_id,omitempty"`
	WaliKelas     *Guru       `gorm:"foreignKey:WaliKelasID"       json:"wali_kelas,omitempty"`
	KelasSantri   []KelasSantri `gorm:"foreignKey:KelasID"         json:"kelas_santri,omitempty"`
}

type KelasSantri struct {
	model.Base
	SantriID uuid.UUID `gorm:"type:uuid;not null;index"  json:"santri_id"`
	KelasID  uuid.UUID `gorm:"type:uuid;not null;index"  json:"kelas_id"`
	Santri   Santri    `gorm:"foreignKey:SantriID"       json:"santri,omitempty"`
	Kelas    Kelas     `gorm:"foreignKey:KelasID"        json:"kelas,omitempty"`
}

// ─────────────────────────────────────────────
//  MataPelajaran
// ─────────────────────────────────────────────

type MataPelajaran struct {
	model.Base
	Nama      string `gorm:"type:varchar(255);not null"            json:"nama"`
	Kode      string `gorm:"type:varchar(50);uniqueIndex;not null" json:"kode"`
	Deskripsi string `gorm:"type:text"                             json:"deskripsi,omitempty"`
}
