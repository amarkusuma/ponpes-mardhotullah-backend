package akademik

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	modelakademik "github.com/simpeg/backend/internal/model/akademik"
	"github.com/simpeg/backend/internal/model/jadwal"
	"github.com/simpeg/backend/internal/model/penilaian"
	"github.com/simpeg/backend/pkg/apperror"
	"github.com/simpeg/backend/pkg/pagination"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ── TahunAjaran ───────────────────────────────────────────────────────────────

func (s *Service) ListTahunAjaran() ([]TahunAjaranResponse, error) {
	list, err := s.repo.ListTahunAjaran()
	if err != nil {
		return nil, apperror.ErrInternal
	}
	result := make([]TahunAjaranResponse, 0, len(list))
	for _, ta := range list {
		result = append(result, TahunAjaranResponse{
			ID:           ta.ID,
			TahunMulai:   ta.TahunMulai,
			TahunSelesai: ta.TahunSelesai,
			IsActive:     ta.IsActive,
		})
	}
	return result, nil
}

// ── Kelas ─────────────────────────────────────────────────────────────────────

func (s *Service) ListKelas(params pagination.Params, tahunAjaranID *uuid.UUID) ([]modelakademik.Kelas, int64, error) {
	list, total, err := s.repo.ListKelas(params, tahunAjaranID)
	if err != nil {
		return nil, 0, apperror.ErrInternal
	}
	return list, total, nil
}

func (s *Service) GetKelas(id uuid.UUID) (*modelakademik.Kelas, error) {
	k, err := s.repo.FindKelasByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.ErrNotFoundMsg("kelas")
		}
		return nil, apperror.ErrInternal
	}
	return k, nil
}

func (s *Service) CreateKelas(req CreateKelasRequest) (*modelakademik.Kelas, error) {
	k := &modelakademik.Kelas{
		NamaKelas:     req.NamaKelas,
		Tingkat:       req.Tingkat,
		TahunAjaranID: req.TahunAjaranID,
		WaliKelasID:   req.WaliKelasID,
	}
	if err := s.repo.CreateKelas(k); err != nil {
		return nil, apperror.ErrInternal
	}
	return s.GetKelas(k.ID)
}

func (s *Service) UpdateKelas(id uuid.UUID, req UpdateKelasRequest) (*modelakademik.Kelas, error) {
	k, err := s.GetKelas(id)
	if err != nil {
		return nil, err
	}
	if req.NamaKelas != "" {
		k.NamaKelas = req.NamaKelas
	}
	if req.Tingkat > 0 {
		k.Tingkat = req.Tingkat
	}
	k.WaliKelasID = req.WaliKelasID
	if err := s.repo.UpdateKelas(k); err != nil {
		return nil, apperror.ErrInternal
	}
	return k, nil
}

func (s *Service) DeleteKelas(id uuid.UUID) error {
	if _, err := s.GetKelas(id); err != nil {
		return err
	}
	return s.repo.DeleteKelas(id)
}

// ── MataPelajaran ─────────────────────────────────────────────────────────────

func (s *Service) ListMapel(params pagination.Params) ([]modelakademik.MataPelajaran, int64, error) {
	list, total, err := s.repo.ListMapel(params, params.Search)
	if err != nil {
		return nil, 0, apperror.ErrInternal
	}
	return list, total, nil
}

func (s *Service) GetMapel(id uuid.UUID) (*modelakademik.MataPelajaran, error) {
	m, err := s.repo.FindMapelByID(id)
	if err != nil {
		return nil, apperror.ErrNotFoundMsg("mata pelajaran")
	}
	return m, nil
}

func (s *Service) CreateMapel(req CreateMapelRequest) (*modelakademik.MataPelajaran, error) {
	m := &modelakademik.MataPelajaran{
		Nama:      req.Nama,
		Kode:      req.Kode,
		Deskripsi: req.Deskripsi,
	}
	if err := s.repo.CreateMapel(m); err != nil {
		return nil, apperror.ErrInternal
	}
	return m, nil
}

func (s *Service) UpdateMapel(id uuid.UUID, req UpdateMapelRequest) (*modelakademik.MataPelajaran, error) {
	m, err := s.GetMapel(id)
	if err != nil {
		return nil, err
	}
	if req.Nama != "" {
		m.Nama = req.Nama
	}
	if req.Deskripsi != "" {
		m.Deskripsi = req.Deskripsi
	}
	if err := s.repo.UpdateMapel(m); err != nil {
		return nil, apperror.ErrInternal
	}
	return m, nil
}

func (s *Service) DeleteMapel(id uuid.UUID) error {
	if _, err := s.GetMapel(id); err != nil {
		return err
	}
	return s.repo.DeleteMapel(id)
}

// ── Jadwal Pelajaran ──────────────────────────────────────────────────────────

func (s *Service) GetJadwalKelas(kelasID uuid.UUID) ([]JadwalResponse, error) {
	list, err := s.repo.ListJadwalByKelas(kelasID)
	if err != nil {
		return nil, apperror.ErrInternal
	}
	return toJadwalResponses(list), nil
}

func (s *Service) GetJadwalGuru(guruUserID uuid.UUID) ([]JadwalResponse, error) {
	guru, err := s.repo.FindGuruByUserID(guruUserID)
	if err != nil {
		return nil, apperror.ErrNotFoundMsg("profil guru")
	}
	list, err := s.repo.ListJadwalByGuru(guru.ID)
	if err != nil {
		return nil, apperror.ErrInternal
	}
	return toJadwalResponses(list), nil
}

func (s *Service) CreateJadwal(req CreateJadwalRequest) (*jadwal.JadwalPelajaran, error) {
	j := &jadwal.JadwalPelajaran{
		KelasID:    req.KelasID,
		MapelID:    req.MapelID,
		GuruID:     req.GuruID,
		Hari:       jadwal.Hari(req.Hari),
		JamMulai:   req.JamMulai,
		JamSelesai: req.JamSelesai,
	}
	if err := s.repo.CreateJadwal(j); err != nil {
		return nil, apperror.ErrInternal
	}
	return j, nil
}

func (s *Service) DeleteJadwal(id uuid.UUID) error {
	return s.repo.DeleteJadwal(id)
}

// ── Penilaian ─────────────────────────────────────────────────────────────────

func (s *Service) GetNilaiSantri(santriUserID uuid.UUID) ([]PenilaianResponse, error) {
	santri, err := s.repo.FindSantriByUserID(santriUserID)
	if err != nil {
		return nil, apperror.ErrNotFoundMsg("profil santri")
	}

	list, err := s.repo.ListPenilaianBySantri(santri.ID)
	if err != nil {
		return nil, apperror.ErrInternal
	}

	result := make([]PenilaianResponse, 0, len(list))
	for _, p := range list {
		result = append(result, PenilaianResponse{
			ID:             p.ID,
			NamaSantri:     santri.User.Name,
			NIK:            santri.NIK,
			NamaMapel:      p.MataPelajaran.Nama,
			JenisPenilaian: p.JenisPenilaian.Nama,
			Nilai:          p.Nilai,
			Semester:       p.Semester,
			TahunAjaran:    fmt.Sprintf("%d/%d", p.TahunAjaran.TahunMulai, p.TahunAjaran.TahunSelesai),
		})
	}
	return result, nil
}

func (s *Service) RekapNilai(kelasID, mapelID uuid.UUID) ([]RekapNilaiItem, error) {
	list, err := s.repo.RekapNilai(kelasID, mapelID)
	if err != nil {
		return nil, apperror.ErrInternal
	}

	santriMap := make(map[uuid.UUID]*RekapNilaiItem)
	santriOrder := make([]uuid.UUID, 0)

	for _, p := range list {
		if _, exists := santriMap[p.SantriID]; !exists {
			santriMap[p.SantriID] = &RekapNilaiItem{
				NamaSantri: p.Santri.User.Name,
				NIK:        p.Santri.NIK,
				Nilai:      []NilaiPerJenis{},
			}
			santriOrder = append(santriOrder, p.SantriID)
		}
		santriMap[p.SantriID].Nilai = append(santriMap[p.SantriID].Nilai, NilaiPerJenis{
			Jenis: p.JenisPenilaian.Nama,
			Nilai: p.Nilai,
		})
	}

	result := make([]RekapNilaiItem, 0, len(santriOrder))
	for _, id := range santriOrder {
		item := santriMap[id]
		sum := 0.0
		for _, n := range item.Nilai {
			sum += n.Nilai
		}
		if len(item.Nilai) > 0 {
			item.RataRata = sum / float64(len(item.Nilai))
		}
		result = append(result, *item)
	}
	return result, nil
}

func (s *Service) CreatePenilaian(req CreatePenilaianRequest) (*penilaian.Penilaian, error) {
	p := &penilaian.Penilaian{
		SantriID:         req.SantriID,
		MapelID:          req.MapelID,
		KelasID:          req.KelasID,
		JenisPenilaianID: req.JenisPenilaianID,
		TahunAjaranID:    req.TahunAjaranID,
		Nilai:            req.Nilai,
		Semester:         req.Semester,
	}
	if err := s.repo.CreatePenilaian(p); err != nil {
		return nil, apperror.ErrInternal
	}
	return p, nil
}

func (s *Service) UpdatePenilaian(id uuid.UUID, req UpdatePenilaianRequest) (*penilaian.Penilaian, error) {
	p, err := s.repo.FindPenilaianByID(id)
	if err != nil {
		return nil, apperror.ErrNotFoundMsg("penilaian")
	}
	p.Nilai = req.Nilai
	if req.Semester > 0 {
		p.Semester = req.Semester
	}
	if err := s.repo.UpdatePenilaian(p); err != nil {
		return nil, apperror.ErrInternal
	}
	return p, nil
}

// ── Admin Profil: Guru ────────────────────────────────────────────────────────

func (s *Service) ListGuruProfil() ([]GuruResponse, error) {
	list, err := s.repo.ListGuru()
	if err != nil {
		return nil, apperror.ErrInternal
	}
	result := make([]GuruResponse, 0, len(list))
	for _, g := range list {
		result = append(result, GuruResponse{
			ID:           g.ID,
			UserID:       g.UserID,
			NIK:          g.NIK,
			Nama:         g.User.Name,
			Email:        g.User.Email,
			Spesialisasi: g.Spesialisasi,
			Foto:         g.Foto,
		})
	}
	return result, nil
}

func (s *Service) UpdateGuruProfil(userID uuid.UUID, req UpdateGuruRequest) (*GuruResponse, error) {
	guru, err := s.repo.FindGuruByUserID(userID)
	if err != nil {
		return nil, apperror.ErrNotFoundMsg("profil guru")
	}
	if req.NIK != "" {
		guru.NIK = req.NIK
	}
	if req.Spesialisasi != "" {
		guru.Spesialisasi = req.Spesialisasi
	}
	if err := s.repo.UpdateGuru(guru); err != nil {
		return nil, apperror.ErrInternal
	}
	return &GuruResponse{
		ID: guru.ID, UserID: guru.UserID, NIK: guru.NIK,
		Nama: guru.User.Name, Email: guru.User.Email,
		Spesialisasi: guru.Spesialisasi, Foto: guru.Foto,
	}, nil
}

// ── Admin Profil: Santri ──────────────────────────────────────────────────────

func (s *Service) ListSantriProfil() ([]SantriResponse, error) {
	list, err := s.repo.ListSantri()
	if err != nil {
		return nil, apperror.ErrInternal
	}
	result := make([]SantriResponse, 0, len(list))
	for _, santri := range list {
		result = append(result, toSantriResponse(santri))
	}
	return result, nil
}

func (s *Service) UpdateSantriProfil(userID uuid.UUID, req UpdateSantriRequest) (*SantriResponse, error) {
	santri, err := s.repo.FindSantriByUserID(userID)
	if err != nil {
		return nil, apperror.ErrNotFoundMsg("profil santri")
	}
	if req.NIK != "" {
		santri.NIK = req.NIK
	}
	if req.TempatLahir != "" {
		santri.TempatLahir = req.TempatLahir
	}
	if req.Alamat != "" {
		santri.Alamat = req.Alamat
	}
	if err := s.repo.UpdateSantri(santri); err != nil {
		return nil, apperror.ErrInternal
	}
	r := toSantriResponse(*santri)
	return &r, nil
}

// ── Admin Profil: Wali ────────────────────────────────────────────────────────

func (s *Service) ListWaliProfil() ([]WaliResponse, error) {
	list, err := s.repo.ListWali()
	if err != nil {
		return nil, apperror.ErrInternal
	}
	result := make([]WaliResponse, 0, len(list))
	for _, w := range list {
		result = append(result, toWaliResponse(w))
	}
	return result, nil
}

func (s *Service) UpdateWaliProfil(userID uuid.UUID, req UpdateWaliRequest) (*WaliResponse, error) {
	wali, err := s.repo.FindWaliByUserID(userID)
	if err != nil {
		return nil, apperror.ErrNotFoundMsg("profil wali")
	}
	if req.NIK != "" {
		wali.NIK = req.NIK
	}
	if req.Hubungan != "" {
		wali.Hubungan = req.Hubungan
	}
	if req.Telepon != "" {
		wali.Telepon = req.Telepon
	}
	if req.Pekerjaan != "" {
		wali.Pekerjaan = req.Pekerjaan
	}
	if req.Alamat != "" {
		wali.Alamat = req.Alamat
	}
	if err := s.repo.UpdateWali(wali); err != nil {
		return nil, apperror.ErrInternal
	}
	r := toWaliResponse(*wali)
	return &r, nil
}

// LinkSantriToWali menghubungkan wali ke santri menggunakan NIK.
func (s *Service) LinkSantriToWali(waliUserID uuid.UUID, nik string) error {
	wali, err := s.repo.FindWaliByUserID(waliUserID)
	if err != nil {
		return apperror.ErrNotFoundMsg("profil wali")
	}
	santri, err := s.repo.FindSantriByNIK(nik)
	if err != nil {
		return apperror.ErrNotFoundMsg("santri dengan NIK " + nik)
	}
	return s.repo.AppendSantriToWali(wali, santri)
}

// UnlinkSantriFromWali memutus relasi wali ↔ santri menggunakan santri_id.
func (s *Service) UnlinkSantriFromWali(waliUserID, santriID uuid.UUID) error {
	wali, err := s.repo.FindWaliByUserID(waliUserID)
	if err != nil {
		return apperror.ErrNotFoundMsg("profil wali")
	}
	santri := &modelakademik.Santri{}
	santri.ID = santriID
	return s.repo.RemoveSantriFromWali(wali, santri)
}

// ── Santri portal endpoints ───────────────────────────────────────────────────

func (s *Service) GetDaftarSantriByKelas(kelasID uuid.UUID) ([]SantriResponse, error) {
	list, err := s.repo.ListSantriByKelas(kelasID)
	if err != nil {
		return nil, apperror.ErrInternal
	}
	result := make([]SantriResponse, 0, len(list))
	for _, santri := range list {
		result = append(result, toSantriResponse(santri))
	}
	return result, nil
}

func (s *Service) GetProfilSantri(userID uuid.UUID) (*SantriResponse, error) {
	santri, err := s.repo.FindSantriByUserID(userID)
	if err != nil {
		return nil, apperror.ErrNotFoundMsg("profil santri")
	}
	r := toSantriResponse(*santri)
	return &r, nil
}

func (s *Service) GetWaliSantri(waliUserID uuid.UUID) ([]SantriResponse, error) {
	list, err := s.repo.FindWaliSantri(waliUserID)
	if err != nil {
		return nil, apperror.ErrInternal
	}
	result := make([]SantriResponse, 0, len(list))
	for _, santri := range list {
		result = append(result, toSantriResponse(santri))
	}
	return result, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func toSantriResponse(s modelakademik.Santri) SantriResponse {
	return SantriResponse{
		ID:           s.ID,
		UserID:       s.UserID,
		NIK:          s.NIK,
		Nama:         s.User.Name,
		Email:        s.User.Email,
		TempatLahir:  s.TempatLahir,
		TanggalLahir: s.TanggalLahir,
		Alamat:       s.Alamat,
		Status:       s.Status,
		Foto:         s.Foto,
	}
}

func toWaliResponse(w modelakademik.Wali) WaliResponse {
	santriList := make([]SantriResponse, 0, len(w.Santri))
	for _, s := range w.Santri {
		santriList = append(santriList, toSantriResponse(s))
	}
	return WaliResponse{
		ID:        w.ID,
		UserID:    w.UserID,
		NIK:       w.NIK,
		Nama:      w.User.Name,
		Email:     w.User.Email,
		Hubungan:  w.Hubungan,
		Telepon:   w.Telepon,
		Pekerjaan: w.Pekerjaan,
		Alamat:    w.Alamat,
		Santri:    santriList,
	}
}

func toJadwalResponses(list []jadwal.JadwalPelajaran) []JadwalResponse {
	result := make([]JadwalResponse, 0, len(list))
	for _, j := range list {
		resp := JadwalResponse{
			ID:         j.ID,
			Hari:       string(j.Hari),
			JamMulai:   j.JamMulai,
			JamSelesai: j.JamSelesai,
			NamaMapel:  j.MataPelajaran.Nama,
			NamaKelas:  j.Kelas.NamaKelas,
			KelasID:    j.KelasID,
		}
		if j.Guru.User.Name != "" {
			resp.NamaGuru = j.Guru.User.Name
		}
		result = append(result, resp)
	}
	return result
}
