package spp

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	modelkeuangan "github.com/simpeg/backend/internal/model/keuangan"
	"github.com/simpeg/backend/pkg/apperror"
	"github.com/simpeg/backend/pkg/pagination"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GenerateSPP membuat tagihan SPP secara massal untuk seluruh santri aktif.
// Santri yang sudah memiliki tagihan bulan+tahun+jenis yang sama akan dilewati.
func (s *Service) GenerateSPP(req GenerateSPPRequest) (*GenerateResult, error) {
	if req.Jenis == "" {
		req.Jenis = "SPP"
	}

	santriList, err := s.repo.ListAktifSantri()
	if err != nil {
		return nil, apperror.ErrInternal
	}

	var toCreate []modelkeuangan.Tagihan
	skipped := 0

	for _, santri := range santriList {
		exists, err := s.repo.CheckExists(santri.ID, req.Bulan, req.Tahun, req.Jenis)
		if err != nil {
			return nil, apperror.ErrInternal
		}
		if exists {
			skipped++
			continue
		}

		toCreate = append(toCreate, modelkeuangan.Tagihan{
			SantriID: santri.ID,
			Jenis:    req.Jenis,
			Nominal:  req.Nominal,
			Bulan:    req.Bulan,
			Tahun:    req.Tahun,
			Status:   modelkeuangan.StatusBelumBayar,
		})
	}

	generated := len(toCreate)
	if generated > 0 {
		if err := s.repo.CreateTagihanBatch(toCreate); err != nil {
			return nil, apperror.ErrInternal
		}
	}

	return &GenerateResult{
		Generated: generated,
		Skipped:   skipped,
		Total:     len(santriList),
	}, nil
}

// ListTagihan mengembalikan daftar tagihan SPP dengan paginasi dan filter.
func (s *Service) ListTagihan(params pagination.Params, filter ListSPPFilter) ([]TagihanSPPResponse, int64, error) {
	rows, total, err := s.repo.ListTagihanSPP(params, filter)
	if err != nil {
		return nil, 0, apperror.ErrInternal
	}

	result := make([]TagihanSPPResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, s.toResponse(row))
	}
	return result, total, nil
}

// GetDetail mengembalikan detail tagihan beserta daftar pembayarannya.
func (s *Service) GetDetail(id uuid.UUID) (*TagihanDetailResponse, error) {
	tagihan, err := s.repo.FindTagihanByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.ErrNotFoundMsg("tagihan")
		}
		return nil, apperror.ErrInternal
	}

	// Hitung total dibayar dan kelas_nama
	totalDibayar := 0.0
	for _, p := range tagihan.Pembayaran {
		totalDibayar += p.JumlahBayar
	}

	kelasNama, _ := s.repo.GetKelasNama(tagihan.SantriID)

	namaSantri := ""
	nik := ""
	if tagihan.Santri.User.Name != "" {
		namaSantri = tagihan.Santri.User.Name
		nik = tagihan.Santri.NIK
	}

	base := TagihanSPPResponse{
		ID:           tagihan.ID,
		SantriID:     tagihan.SantriID,
		NamaSantri:   namaSantri,
		NIK:          nik,
		KelasNama:    kelasNama,
		Jenis:        tagihan.Jenis,
		Nominal:      tagihan.Nominal,
		Bulan:        tagihan.Bulan,
		Tahun:        tagihan.Tahun,
		Status:       tagihan.Status,
		TotalDibayar: totalDibayar,
		Sisa:         tagihan.Nominal - totalDibayar,
		CreatedAt:    tagihan.CreatedAt,
	}

	pembayaranList := make([]PembayaranResponse, 0, len(tagihan.Pembayaran))
	for _, p := range tagihan.Pembayaran {
		pembayaranList = append(pembayaranList, PembayaranResponse{
			ID:           p.ID,
			TagihanID:    p.TagihanID,
			JumlahBayar:  p.JumlahBayar,
			Metode:       p.Metode,
			TanggalBayar: p.TanggalBayar,
			Keterangan:   p.Keterangan,
			CreatedAt:    p.CreatedAt,
		})
	}

	return &TagihanDetailResponse{
		TagihanSPPResponse: base,
		Pembayaran:         pembayaranList,
	}, nil
}

// BayarTagihan mencatat pembayaran tagihan secara atomik.
// Validasi: tagihan belum lunas, jumlah bayar tidak melebihi sisa.
func (s *Service) BayarTagihan(tagihanID uuid.UUID, req BayarTagihanRequest) (*PembayaranResponse, error) {
	tagihan, err := s.repo.FindTagihanByID(tagihanID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.ErrNotFoundMsg("tagihan")
		}
		return nil, apperror.ErrInternal
	}

	if tagihan.Status == modelkeuangan.StatusLunas {
		return nil, &apperror.AppError{
			Code:    400,
			Message: "tagihan sudah lunas",
		}
	}

	// Hitung sisa tagihan
	totalDibayar, err := s.repo.SumPembayaran(tagihanID)
	if err != nil {
		return nil, apperror.ErrInternal
	}
	sisa := tagihan.Nominal - totalDibayar

	if req.JumlahBayar > sisa {
		return nil, &apperror.AppError{
			Code:    400,
			Message: "jumlah bayar melebihi sisa tagihan",
			Detail:  "sisa tagihan: " + formatRupiah(sisa),
		}
	}

	// Jalankan dalam transaksi
	tx := s.repo.BeginTx()
	if tx.Error != nil {
		return nil, apperror.ErrInternal
	}

	pembayaran := &modelkeuangan.Pembayaran{
		TagihanID:    tagihanID,
		JumlahBayar:  req.JumlahBayar,
		TanggalBayar: req.TanggalBayar,
		Metode:       req.Metode,
		Keterangan:   req.Keterangan,
	}

	if err := s.repo.CreatePembayaranTx(tx, pembayaran); err != nil {
		tx.Rollback()
		return nil, apperror.ErrInternal
	}

	// Tentukan status baru
	newTotal := totalDibayar + req.JumlahBayar
	var newStatus modelkeuangan.StatusTagihan
	switch {
	case newTotal >= tagihan.Nominal:
		newStatus = modelkeuangan.StatusLunas
	case newTotal > 0:
		newStatus = modelkeuangan.StatusSebagian
	default:
		newStatus = modelkeuangan.StatusBelumBayar
	}

	if err := s.repo.UpdateStatusTx(tx, tagihanID, newStatus); err != nil {
		tx.Rollback()
		return nil, apperror.ErrInternal
	}

	if err := tx.Commit().Error; err != nil {
		return nil, apperror.ErrInternal
	}

	return &PembayaranResponse{
		ID:           pembayaran.ID,
		TagihanID:    pembayaran.TagihanID,
		JumlahBayar:  pembayaran.JumlahBayar,
		Metode:       pembayaran.Metode,
		TanggalBayar: pembayaran.TanggalBayar,
		Keterangan:   pembayaran.Keterangan,
		CreatedAt:    pembayaran.CreatedAt,
	}, nil
}

// GetPembayaran mengembalikan daftar pembayaran untuk satu tagihan.
func (s *Service) GetPembayaran(tagihanID uuid.UUID) ([]PembayaranResponse, error) {
	tagihan, err := s.repo.FindTagihanByID(tagihanID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.ErrNotFoundMsg("tagihan")
		}
		return nil, apperror.ErrInternal
	}

	result := make([]PembayaranResponse, 0, len(tagihan.Pembayaran))
	for _, p := range tagihan.Pembayaran {
		result = append(result, PembayaranResponse{
			ID:           p.ID,
			TagihanID:    p.TagihanID,
			JumlahBayar:  p.JumlahBayar,
			Metode:       p.Metode,
			TanggalBayar: p.TanggalBayar,
			Keterangan:   p.Keterangan,
			CreatedAt:    p.CreatedAt,
		})
	}
	return result, nil
}

// GetRekap mengembalikan rekapitulasi tagihan per bulan.
func (s *Service) GetRekap(jenis string, tahun int) ([]RekapBulanan, error) {
	list, err := s.repo.GetRekapBulanan(jenis, tahun)
	if err != nil {
		return nil, apperror.ErrInternal
	}
	return list, nil
}

// GetUnpaid mengembalikan tagihan yang belum atau sebagian bayar.
func (s *Service) GetUnpaid(jenis string, bulan, tahun int) ([]UnpaidItem, error) {
	list, err := s.repo.ListUnpaid(jenis, bulan, tahun)
	if err != nil {
		return nil, apperror.ErrInternal
	}
	return list, nil
}

// GetAllForExport mengembalikan semua tagihan sesuai filter tanpa paginasi.
func (s *Service) GetAllForExport(filter ListSPPFilter) ([]TagihanSPPResponse, error) {
	rows, err := s.repo.ListTagihanSPPAll(filter)
	if err != nil {
		return nil, apperror.ErrInternal
	}

	result := make([]TagihanSPPResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, s.toResponse(row))
	}
	return result, nil
}

// toResponse mengonversi tagScan menjadi TagihanSPPResponse.
func (s *Service) toResponse(row tagScan) TagihanSPPResponse {
	var createdAt time.Time
	switch v := row.CreatedAt.(type) {
	case time.Time:
		createdAt = v
	case string:
		createdAt, _ = time.Parse(time.RFC3339, v)
	}

	return TagihanSPPResponse{
		ID:           row.ID,
		SantriID:     row.SantriID,
		NamaSantri:   row.NamaSantri,
		NIK:          row.NIK,
		KelasNama:    row.KelasNama,
		Jenis:        row.Jenis,
		Nominal:      row.Nominal,
		Bulan:        row.Bulan,
		Tahun:        row.Tahun,
		Status:       row.Status,
		TotalDibayar: row.TotalDibayar,
		Sisa:         row.Nominal - row.TotalDibayar,
		CreatedAt:    createdAt,
	}
}

// formatRupiah memformat float64 sebagai string angka (tanpa simbol, cukup untuk pesan error).
func formatRupiah(v float64) string {
	return fmt.Sprintf("%.2f", v)
}
