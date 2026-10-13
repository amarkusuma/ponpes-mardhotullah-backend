package akademik

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/simpeg/backend/internal/model/akademik"
	"github.com/simpeg/backend/internal/model/jadwal"
	"github.com/simpeg/backend/internal/model/penilaian"
	"github.com/simpeg/backend/pkg/pagination"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ── TahunAjaran ───────────────────────────────────────────────────────────────

func (r *Repository) ListTahunAjaran() ([]akademik.TahunAjaran, error) {
	var list []akademik.TahunAjaran
	err := r.db.Order("tahun_mulai desc").Find(&list).Error
	return list, err
}

func (r *Repository) ActiveTahunAjaran() (*akademik.TahunAjaran, error) {
	var ta akademik.TahunAjaran
	err := r.db.Where("is_active = true").First(&ta).Error
	return &ta, err
}

// ── Kelas ─────────────────────────────────────────────────────────────────────

func (r *Repository) ListKelas(params pagination.Params, tahunAjaranID *uuid.UUID) ([]akademik.Kelas, int64, error) {
	var list []akademik.Kelas
	var total int64

	q := r.db.Model(&akademik.Kelas{}).
		Preload("TahunAjaran").
		Preload("WaliKelas.User")

	if tahunAjaranID != nil {
		q = q.Where("tahun_ajaran_id = ?", *tahunAjaranID)
	}

	q.Count(&total)
	err := q.Order(params.OrderClause()).Limit(params.Limit).Offset(params.Offset()).Find(&list).Error
	return list, total, err
}

func (r *Repository) FindKelasByID(id uuid.UUID) (*akademik.Kelas, error) {
	var k akademik.Kelas
	err := r.db.Preload("TahunAjaran").Preload("WaliKelas.User").
		Preload("KelasSantri.Santri.User").
		First(&k, "id = ?", id).Error
	return &k, err
}

func (r *Repository) CreateKelas(k *akademik.Kelas) error { return r.db.Create(k).Error }
func (r *Repository) UpdateKelas(k *akademik.Kelas) error { return r.db.Save(k).Error }
func (r *Repository) DeleteKelas(id uuid.UUID) error {
	return r.db.Delete(&akademik.Kelas{}, "id = ?", id).Error
}

// ── MataPelajaran ─────────────────────────────────────────────────────────────

func (r *Repository) ListMapel(params pagination.Params, search string) ([]akademik.MataPelajaran, int64, error) {
	var list []akademik.MataPelajaran
	var total int64

	q := r.db.Model(&akademik.MataPelajaran{})
	if search != "" {
		q = q.Where("nama ILIKE ? OR kode ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	q.Count(&total)
	err := q.Order(params.OrderClause()).Limit(params.Limit).Offset(params.Offset()).Find(&list).Error
	return list, total, err
}

func (r *Repository) FindMapelByID(id uuid.UUID) (*akademik.MataPelajaran, error) {
	var m akademik.MataPelajaran
	err := r.db.First(&m, "id = ?", id).Error
	return &m, err
}

func (r *Repository) CreateMapel(m *akademik.MataPelajaran) error { return r.db.Create(m).Error }
func (r *Repository) UpdateMapel(m *akademik.MataPelajaran) error { return r.db.Save(m).Error }
func (r *Repository) DeleteMapel(id uuid.UUID) error {
	return r.db.Delete(&akademik.MataPelajaran{}, "id = ?", id).Error
}

// ── JadwalPelajaran ───────────────────────────────────────────────────────────

func (r *Repository) ListJadwalByKelas(kelasID uuid.UUID) ([]jadwal.JadwalPelajaran, error) {
	var list []jadwal.JadwalPelajaran
	err := r.db.
		Preload("MataPelajaran").
		Preload("Guru.User").
		Preload("Kelas").
		Where("kelas_id = ?", kelasID).
		Order("hari, jam_mulai").
		Find(&list).Error
	return list, err
}

func (r *Repository) ListJadwalByGuru(guruID uuid.UUID) ([]jadwal.JadwalPelajaran, error) {
	var list []jadwal.JadwalPelajaran
	err := r.db.
		Preload("MataPelajaran").
		Preload("Kelas").
		Where("guru_id = ?", guruID).
		Order("hari, jam_mulai").
		Find(&list).Error
	return list, err
}

func (r *Repository) CreateJadwal(j *jadwal.JadwalPelajaran) error { return r.db.Create(j).Error }
func (r *Repository) DeleteJadwal(id uuid.UUID) error {
	return r.db.Delete(&jadwal.JadwalPelajaran{}, "id = ?", id).Error
}

// ── Penilaian ─────────────────────────────────────────────────────────────────

func (r *Repository) ListPenilaianBySantri(santriID uuid.UUID) ([]penilaian.Penilaian, error) {
	var list []penilaian.Penilaian
	err := r.db.
		Preload("MataPelajaran").
		Preload("JenisPenilaian").
		Preload("TahunAjaran").
		Preload("Kelas").
		Where("santri_id = ?", santriID).
		Order("created_at desc").
		Find(&list).Error
	return list, err
}

func (r *Repository) RekapNilai(kelasID, mapelID uuid.UUID) ([]penilaian.Penilaian, error) {
	var list []penilaian.Penilaian
	err := r.db.
		Preload("Santri.User").
		Preload("JenisPenilaian").
		Where("kelas_id = ? AND mapel_id = ?", kelasID, mapelID).
		Order("santri_id").
		Find(&list).Error
	return list, err
}

func (r *Repository) FindPenilaianByID(id uuid.UUID) (*penilaian.Penilaian, error) {
	var p penilaian.Penilaian
	err := r.db.First(&p, "id = ?", id).Error
	return &p, err
}

func (r *Repository) CreatePenilaian(p *penilaian.Penilaian) error { return r.db.Create(p).Error }
func (r *Repository) UpdatePenilaian(p *penilaian.Penilaian) error { return r.db.Save(p).Error }

// ── Guru ──────────────────────────────────────────────────────────────────────

func (r *Repository) FindGuruByUserID(userID uuid.UUID) (*akademik.Guru, error) {
	var g akademik.Guru
	err := r.db.Preload("User").Where("user_id = ?", userID).First(&g).Error
	return &g, err
}

func (r *Repository) ListGuru() ([]akademik.Guru, error) {
	var list []akademik.Guru
	err := r.db.Preload("User").Order("created_at desc").Find(&list).Error
	return list, err
}

func (r *Repository) UpdateGuru(g *akademik.Guru) error { return r.db.Save(g).Error }

// ── Santri ────────────────────────────────────────────────────────────────────

func (r *Repository) ListSantriByKelas(kelasID uuid.UUID) ([]akademik.Santri, error) {
	var list []akademik.Santri
	err := r.db.
		Joins("JOIN kelas_santris ks ON ks.santri_id = santris.id AND ks.kelas_id = ?", kelasID).
		Preload("User").
		Find(&list).Error
	return list, err
}

func (r *Repository) FindSantriByUserID(userID uuid.UUID) (*akademik.Santri, error) {
	var s akademik.Santri
	err := r.db.Preload("User").Where("user_id = ?", userID).First(&s).Error
	return &s, err
}

func (r *Repository) FindSantriByNIK(nik string) (*akademik.Santri, error) {
	var s akademik.Santri
	err := r.db.Preload("User").Where("nik = ?", nik).First(&s).Error
	return &s, err
}

func (r *Repository) ListSantri() ([]akademik.Santri, error) {
	var list []akademik.Santri
	err := r.db.Preload("User").Order("created_at desc").Find(&list).Error
	return list, err
}

func (r *Repository) UpdateSantri(s *akademik.Santri) error { return r.db.Save(s).Error }

// ── Wali ──────────────────────────────────────────────────────────────────────

func (r *Repository) FindWaliByUserID(userID uuid.UUID) (*akademik.Wali, error) {
	var w akademik.Wali
	err := r.db.Preload("User").Preload("Santri.User").Where("user_id = ?", userID).First(&w).Error
	return &w, err
}

// FindWaliSantri mengembalikan semua santri yang diasuh wali ini.
func (r *Repository) FindWaliSantri(waliUserID uuid.UUID) ([]akademik.Santri, error) {
	wali, err := r.FindWaliByUserID(waliUserID)
	if err != nil {
		return nil, err
	}
	return wali.Santri, nil
}

func (r *Repository) ListWali() ([]akademik.Wali, error) {
	var list []akademik.Wali
	err := r.db.Preload("User").Preload("Santri.User").Order("created_at desc").Find(&list).Error
	return list, err
}

func (r *Repository) UpdateWali(w *akademik.Wali) error { return r.db.Save(w).Error }

func (r *Repository) AppendSantriToWali(wali *akademik.Wali, santri *akademik.Santri) error {
	return r.db.Model(wali).Association("Santri").Append(santri)
}

func (r *Repository) RemoveSantriFromWali(wali *akademik.Wali, santri *akademik.Santri) error {
	return r.db.Model(wali).Association("Santri").Delete(santri)
}

// ── JenisPenilaian ────────────────────────────────────────────────────────────

func (r *Repository) ListJenisPenilaian() ([]penilaian.JenisPenilaian, error) {
	var list []penilaian.JenisPenilaian
	err := r.db.Find(&list).Error
	return list, err
}
