package akademik

import (
	"time"

	"github.com/google/uuid"
	"github.com/simpeg/backend/internal/model/akademik"
)

// ── Kelas ────────────────────────────────────────────────────────────────────

type CreateKelasRequest struct {
	NamaKelas     string     `json:"nama_kelas"      binding:"required"`
	Tingkat       int        `json:"tingkat"         binding:"required,min=1"`
	TahunAjaranID uuid.UUID  `json:"tahun_ajaran_id" binding:"required"`
	WaliKelasID   *uuid.UUID `json:"wali_kelas_id"`
}

type UpdateKelasRequest struct {
	NamaKelas   string     `json:"nama_kelas"`
	Tingkat     int        `json:"tingkat"`
	WaliKelasID *uuid.UUID `json:"wali_kelas_id"`
}

// ── MataPelajaran ─────────────────────────────────────────────────────────────

type CreateMapelRequest struct {
	Nama      string `json:"nama"      binding:"required"`
	Kode      string `json:"kode"      binding:"required"`
	Deskripsi string `json:"deskripsi"`
}

type UpdateMapelRequest struct {
	Nama      string `json:"nama"`
	Deskripsi string `json:"deskripsi"`
}

// ── JadwalPelajaran ───────────────────────────────────────────────────────────

type CreateJadwalRequest struct {
	KelasID    uuid.UUID `json:"kelas_id"    binding:"required"`
	MapelID    uuid.UUID `json:"mapel_id"    binding:"required"`
	GuruID     uuid.UUID `json:"guru_id"     binding:"required"`
	Hari       string    `json:"hari"        binding:"required,oneof=Senin Selasa Rabu Kamis Jumat Sabtu Ahad"`
	JamMulai   string    `json:"jam_mulai"   binding:"required"`
	JamSelesai string    `json:"jam_selesai" binding:"required"`
}

// ── Penilaian ─────────────────────────────────────────────────────────────────

type CreatePenilaianRequest struct {
	SantriID         uuid.UUID `json:"santri_id"          binding:"required"`
	MapelID          uuid.UUID `json:"mapel_id"           binding:"required"`
	KelasID          uuid.UUID `json:"kelas_id"           binding:"required"`
	JenisPenilaianID uuid.UUID `json:"jenis_penilaian_id" binding:"required"`
	TahunAjaranID    uuid.UUID `json:"tahun_ajaran_id"    binding:"required"`
	Nilai            float64   `json:"nilai"              binding:"required,min=0,max=100"`
	Semester         int       `json:"semester"           binding:"required,oneof=1 2"`
}

type UpdatePenilaianRequest struct {
	Nilai    float64 `json:"nilai"    binding:"required,min=0,max=100"`
	Semester int     `json:"semester"`
}

// ── Admin Profil Requests ─────────────────────────────────────────────────────

// LinkSantriRequest menghubungkan wali ke santri via NIK.
type LinkSantriRequest struct {
	NIK string `json:"nik" binding:"required"`
}

type UpdateGuruRequest struct {
	NIK          string `json:"nik"`
	Spesialisasi string `json:"spesialisasi"`
}

type UpdateSantriRequest struct {
	NIK         string `json:"nik"`
	TempatLahir string `json:"tempat_lahir"`
	Alamat      string `json:"alamat"`
}

type UpdateWaliRequest struct {
	NIK       string `json:"nik"`
	Hubungan  string `json:"hubungan"`
	Telepon   string `json:"telepon"`
	Pekerjaan string `json:"pekerjaan"`
	Alamat    string `json:"alamat"`
}

// ── Response types ────────────────────────────────────────────────────────────

type JadwalResponse struct {
	ID         uuid.UUID `json:"id"`
	Hari       string    `json:"hari"`
	JamMulai   string    `json:"jam_mulai"`
	JamSelesai string    `json:"jam_selesai"`
	NamaMapel  string    `json:"nama_mapel"`
	NamaGuru   string    `json:"nama_guru"`
	NamaKelas  string    `json:"nama_kelas"`
	KelasID    uuid.UUID `json:"kelas_id"`
}

type PenilaianResponse struct {
	ID             uuid.UUID `json:"id"`
	NamaSantri     string    `json:"nama_santri"`
	NIK            string    `json:"nik"`
	NamaMapel      string    `json:"nama_mapel"`
	JenisPenilaian string    `json:"jenis_penilaian"`
	Nilai          float64   `json:"nilai"`
	Semester       int       `json:"semester"`
	TahunAjaran    string    `json:"tahun_ajaran"`
}

type RekapNilaiItem struct {
	NamaSantri string          `json:"nama_santri"`
	NIK        string          `json:"nik"`
	Nilai      []NilaiPerJenis `json:"nilai"`
	RataRata   float64         `json:"rata_rata"`
}

type NilaiPerJenis struct {
	Jenis string  `json:"jenis"`
	Nilai float64 `json:"nilai"`
}

type SantriResponse struct {
	ID           uuid.UUID             `json:"id"`
	UserID       uuid.UUID             `json:"user_id"`
	NIK          string                `json:"nik"`
	Nama         string                `json:"nama"`
	Email        string                `json:"email,omitempty"`
	TempatLahir  string                `json:"tempat_lahir"`
	TanggalLahir time.Time             `json:"tanggal_lahir"`
	Alamat       string                `json:"alamat,omitempty"`
	Status       akademik.SantriStatus `json:"status"`
	Foto         string                `json:"foto,omitempty"`
}

type GuruResponse struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	NIK          string    `json:"nik"`
	Nama         string    `json:"nama"`
	Email        string    `json:"email,omitempty"`
	Spesialisasi string    `json:"spesialisasi"`
	Foto         string    `json:"foto,omitempty"`
}

type WaliResponse struct {
	ID        uuid.UUID        `json:"id"`
	UserID    uuid.UUID        `json:"user_id"`
	NIK       string           `json:"nik"`
	Nama      string           `json:"nama"`
	Email     string           `json:"email,omitempty"`
	Hubungan  string           `json:"hubungan"`
	Telepon   string           `json:"telepon"`
	Pekerjaan string           `json:"pekerjaan"`
	Alamat    string           `json:"alamat"`
	Santri    []SantriResponse `json:"santri"`
}

type TahunAjaranResponse struct {
	ID           uuid.UUID `json:"id"`
	TahunMulai   int       `json:"tahun_mulai"`
	TahunSelesai int       `json:"tahun_selesai"`
	IsActive     bool      `json:"is_active"`
}
