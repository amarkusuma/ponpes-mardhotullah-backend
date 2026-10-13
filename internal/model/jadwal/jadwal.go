// Package jadwal berisi model JadwalPelajaran.
package jadwal

import (
	"github.com/google/uuid"
	"github.com/simpeg/backend/internal/model"
	"github.com/simpeg/backend/internal/model/akademik"
)

type Hari string

const (
	HariSenin  Hari = "Senin"
	HariSelasa Hari = "Selasa"
	HariRabu   Hari = "Rabu"
	HariKamis  Hari = "Kamis"
	HariJumat  Hari = "Jumat"
	HariSabtu  Hari = "Sabtu"
	HariAhad   Hari = "Ahad"
)

// JadwalPelajaran menyimpan jadwal per kelas per hari.
// JamMulai & JamSelesai disimpan sebagai string "HH:MM" (format 24 jam).
type JadwalPelajaran struct {
	model.Base
	KelasID       uuid.UUID              `gorm:"type:uuid;not null;index"  json:"kelas_id"`
	Kelas         akademik.Kelas         `gorm:"foreignKey:KelasID"        json:"kelas,omitempty"`
	MapelID       uuid.UUID              `gorm:"type:uuid;not null;index"  json:"mapel_id"`
	MataPelajaran akademik.MataPelajaran `gorm:"foreignKey:MapelID"        json:"mata_pelajaran,omitempty"`
	GuruID        uuid.UUID              `gorm:"type:uuid;not null;index"  json:"guru_id"`
	Guru          akademik.Guru          `gorm:"foreignKey:GuruID"         json:"guru,omitempty"`
	Hari          Hari                   `gorm:"type:varchar(10);not null" json:"hari"`
	JamMulai      string                 `gorm:"type:varchar(5);not null"  json:"jam_mulai"`  // "07:00"
	JamSelesai    string                 `gorm:"type:varchar(5);not null"  json:"jam_selesai"` // "08:30"
}
