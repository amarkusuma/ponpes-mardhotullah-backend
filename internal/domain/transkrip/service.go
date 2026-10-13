// Package transkrip mengelola generate dan download transkrip nilai santri.
// PDF di-generate menggunakan pure-Go encoding/pdf approach via text template;
// untuk kebutuhan produksi, ganti dengan github.com/jung-kurt/gofpdf atau
// github.com/signintech/gopdf.
package transkrip

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/simpeg/backend/internal/model/akademik"
	"github.com/simpeg/backend/internal/model/penilaian"
	modelpenilaian "github.com/simpeg/backend/internal/model/penilaian"
	"github.com/simpeg/backend/pkg/apperror"
	"github.com/simpeg/backend/pkg/storage"
)

const transkripPrefix = "transkrip/"

type Service struct {
	db      *gorm.DB
	storage *storage.Storage
}

func NewService(db *gorm.DB, store *storage.Storage) *Service {
	return &Service{db: db, storage: store}
}

type TranskripData struct {
	NamaSantri  string
	NIK         string
	TempatLahir string
	TanggalLahir string
	TahunAjaran string
	Semester    int
	GeneratedAt string
	Nilai       []NilaiRow
	RataRata    float64
}

type NilaiRow struct {
	NamaMapel      string
	JenisPenilaian string
	Nilai          float64
}

// Generate membuat file transkrip untuk santri dan mengembalikan object key-nya.
func (s *Service) Generate(ctx context.Context, santriID uuid.UUID) (string, error) {
	// Load santri
	var santri akademik.Santri
	if err := s.db.Preload("User").First(&santri, "id = ?", santriID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", apperror.ErrNotFoundMsg("santri")
		}
		return "", apperror.ErrInternal
	}

	// Load tahun ajaran aktif
	var ta akademik.TahunAjaran
	s.db.Where("is_active = true").First(&ta)

	// Load penilaian
	var nilaiList []penilaian.Penilaian
	s.db.
		Preload("MataPelajaran").
		Preload("JenisPenilaian").
		Where("santri_id = ? AND tahun_ajaran_id = ?", santriID, ta.ID).
		Find(&nilaiList)

	// Build data
	rows := make([]NilaiRow, 0, len(nilaiList))
	sum := 0.0
	for _, n := range nilaiList {
		rows = append(rows, NilaiRow{
			NamaMapel:      n.MataPelajaran.Nama,
			JenisPenilaian: n.JenisPenilaian.Nama,
			Nilai:          n.Nilai,
		})
		sum += n.Nilai
	}
	rataRata := 0.0
	if len(rows) > 0 {
		rataRata = sum / float64(len(rows))
	}

	data := TranskripData{
		NamaSantri:   santri.User.Name,
		NIK:          santri.NIK,
		TempatLahir:  santri.TempatLahir,
		TanggalLahir: santri.TanggalLahir.Format("02 January 2006"),
		TahunAjaran:  fmt.Sprintf("%d/%d", ta.TahunMulai, ta.TahunSelesai),
		GeneratedAt:  time.Now().Format("02 January 2006, 15:04"),
		Nilai:        rows,
		RataRata:     rataRata,
	}

	// Render template
	content, err := renderTemplate(data)
	if err != nil {
		return "", apperror.ErrInternal
	}

	// Simpan file ke MinIO
	key := fmt.Sprintf("%stranskrip_%s_%d.txt", transkripPrefix, santriID.String(), time.Now().Unix())
	if err := s.storage.Put(ctx, key, strings.NewReader(content), int64(len(content)), "text/plain; charset=utf-8"); err != nil {
		return "", apperror.ErrInternal
	}

	// Simpan record rapor
	rapor := modelpenilaian.Rapor{
		SantriID:      santriID,
		Semester:      1,
		TahunAjaranID: ta.ID,
		FilePath:      key,
		GeneratedAt:   time.Now(),
	}
	s.db.Create(&rapor)

	return key, nil
}

// OpenFile membuka file transkrip dari MinIO. Caller wajib menutup reader.
func (s *Service) OpenFile(ctx context.Context, rapor *modelpenilaian.Rapor) (io.ReadCloser, int64, error) {
	// Record lama menyimpan path lokal "storage/transkrip/..."
	key := strings.TrimPrefix(rapor.FilePath, "storage/")

	obj, info, err := s.storage.Get(ctx, key)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, 0, apperror.ErrNotFoundMsg("file transkrip")
		}
		return nil, 0, apperror.ErrInternal
	}
	return obj, info.Size, nil
}

// GetRaporByID mengembalikan rapor/transkrip berdasarkan ID.
func (s *Service) GetRaporByID(raporID uuid.UUID) (*modelpenilaian.Rapor, error) {
	var rapor modelpenilaian.Rapor
	err := s.db.Preload("Santri.User").First(&rapor, "id = ?", raporID).Error
	if err != nil {
		return nil, apperror.ErrNotFoundMsg("transkrip")
	}
	return &rapor, nil
}

// ── Template sederhana (ganti dengan gofpdf untuk PDF produksi) ───────────────

var transkripTmpl = template.Must(template.New("transkrip").Parse(`
=====================================
    TRANSKRIP NILAI SANTRI
    Pesantren Management System
=====================================

Nama        : {{.NamaSantri}}
NIK         : {{.NIK}}
Tempat Lahir: {{.TempatLahir}}
Tgl Lahir   : {{.TanggalLahir}}
Tahun Ajaran: {{.TahunAjaran}}
Semester    : {{.Semester}}

-------------------------------------
 DAFTAR NILAI
-------------------------------------
{{- range .Nilai}}
  {{.NamaMapel}} ({{.JenisPenilaian}}): {{printf "%.2f" .Nilai}}
{{- end}}

-------------------------------------
  Rata-rata   : {{printf "%.2f" .RataRata}}
-------------------------------------

Digenerate pada: {{.GeneratedAt}}
=====================================
`))

func renderTemplate(data TranskripData) (string, error) {
	var buf bytes.Buffer
	if err := transkripTmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
