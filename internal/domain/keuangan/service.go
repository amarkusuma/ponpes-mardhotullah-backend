package keuangan

import (
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

// ── Tagihan ───────────────────────────────────────────────────────────────────

func (s *Service) ListTagihan(params pagination.Params, filter ListTagihanFilter) ([]TagihanResponse, int64, error) {
	list, total, err := s.repo.ListTagihan(params, filter)
	if err != nil {
		return nil, 0, apperror.ErrInternal
	}

	result := make([]TagihanResponse, 0, len(list))
	for _, t := range list {
		result = append(result, s.toTagihanResponse(t))
	}
	return result, total, nil
}

func (s *Service) CreateTagihan(req CreateTagihanRequest) (*TagihanResponse, error) {
	t := &modelkeuangan.Tagihan{
		SantriID: req.SantriID,
		Jenis:    req.Jenis,
		Nominal:  req.Nominal,
		Bulan:    req.Bulan,
		Tahun:    req.Tahun,
		Status:   modelkeuangan.StatusBelumBayar,
	}
	if err := s.repo.CreateTagihan(t); err != nil {
		return nil, apperror.ErrInternal
	}
	// Reload dengan preload
	reloaded, err := s.repo.FindTagihanByID(t.ID)
	if err != nil {
		return nil, apperror.ErrInternal
	}
	r := s.toTagihanResponse(*reloaded)
	return &r, nil
}

// GetTagihanSantri mengembalikan tagihan milik santri berdasarkan userID.
func (s *Service) GetTagihanSantri(santriUserID uuid.UUID) ([]TagihanResponse, error) {
	santri, err := s.repo.FindSantriByUserID(santriUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.ErrNotFoundMsg("profil santri")
		}
		return nil, apperror.ErrInternal
	}

	params := pagination.Params{Page: 1, Limit: 100, SortBy: "tahun", Order: "desc"}
	filter := ListTagihanFilter{SantriID: &santri.ID}
	list, _, err := s.ListTagihan(params, filter)
	return list, err
}

// ── Pembayaran ────────────────────────────────────────────────────────────────

func (s *Service) CreatePembayaran(req CreatePembayaranRequest) (*PembayaranResponse, error) {
	tagihan, err := s.repo.FindTagihanByID(req.TagihanID)
	if err != nil {
		return nil, apperror.ErrNotFoundMsg("tagihan")
	}

	p := &modelkeuangan.Pembayaran{
		TagihanID:    req.TagihanID,
		JumlahBayar:  req.JumlahBayar,
		TanggalBayar: req.TanggalBayar,
		Metode:       req.Metode,
		Keterangan:   req.Keterangan,
	}
	if err := s.repo.CreatePembayaran(p); err != nil {
		return nil, apperror.ErrInternal
	}

	// Update status tagihan otomatis
	s.updateTagihanStatus(tagihan)

	return &PembayaranResponse{
		ID:           p.ID,
		TagihanID:    p.TagihanID,
		JumlahBayar:  p.JumlahBayar,
		TanggalBayar: p.TanggalBayar,
		Metode:       p.Metode,
		Keterangan:   p.Keterangan,
		CreatedAt:    p.CreatedAt,
	}, nil
}

func (s *Service) GetRiwayatPembayaran(santriUserID uuid.UUID) ([]PembayaranResponse, error) {
	santri, err := s.repo.FindSantriByUserID(santriUserID)
	if err != nil {
		return nil, apperror.ErrNotFoundMsg("profil santri")
	}

	list, err := s.repo.ListRiwayatPembayaranBySantri(santri.ID)
	if err != nil {
		return nil, apperror.ErrInternal
	}

	result := make([]PembayaranResponse, 0, len(list))
	for _, p := range list {
		result = append(result, PembayaranResponse{
			ID:           p.ID,
			TagihanID:    p.TagihanID,
			JumlahBayar:  p.JumlahBayar,
			TanggalBayar: p.TanggalBayar,
			Metode:       p.Metode,
			Keterangan:   p.Keterangan,
			CreatedAt:    p.CreatedAt,
		})
	}
	return result, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (s *Service) toTagihanResponse(t modelkeuangan.Tagihan) TagihanResponse {
	totalDibayar := 0.0
	for _, p := range t.Pembayaran {
		totalDibayar += p.JumlahBayar
	}

	namaSantri := ""
	nik := ""
	if t.Santri.User.Name != "" {
		namaSantri = t.Santri.User.Name
		nik = t.Santri.NIK
	}

	return TagihanResponse{
		ID:           t.ID,
		SantriID:     t.SantriID,
		NamaSantri:   namaSantri,
		NIK:          nik,
		Jenis:        t.Jenis,
		Nominal:      t.Nominal,
		Bulan:        t.Bulan,
		Tahun:        t.Tahun,
		Status:       t.Status,
		TotalDibayar: totalDibayar,
		Sisa:         t.Nominal - totalDibayar,
		CreatedAt:    t.CreatedAt,
	}
}

func (s *Service) updateTagihanStatus(tagihan *modelkeuangan.Tagihan) {
	total, err := s.repo.SumPembayaran(tagihan.ID)
	if err != nil {
		return
	}

	var status modelkeuangan.StatusTagihan
	switch {
	case total >= tagihan.Nominal:
		status = modelkeuangan.StatusLunas
	case total > 0:
		status = modelkeuangan.StatusSebagian
	default:
		status = modelkeuangan.StatusBelumBayar
	}

	s.repo.UpdateTagihanStatus(tagihan.ID, status) //nolint:errcheck
}
