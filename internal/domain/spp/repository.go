package spp

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/simpeg/backend/internal/model/akademik"
	modelkeuangan "github.com/simpeg/backend/internal/model/keuangan"
	"github.com/simpeg/backend/pkg/pagination"
)

// tagScan adalah flat struct untuk hasil raw query JOIN tagihan + kelas.
type tagScan struct {
	ID           uuid.UUID                   `gorm:"column:id"`
	SantriID     uuid.UUID                   `gorm:"column:santri_id"`
	NamaSantri   string                      `gorm:"column:nama_santri"`
	NIK          string                      `gorm:"column:nik"`
	KelasNama    string                      `gorm:"column:kelas_nama"`
	Jenis        string                      `gorm:"column:jenis"`
	Nominal      float64                     `gorm:"column:nominal"`
	Bulan        int                         `gorm:"column:bulan"`
	Tahun        int                         `gorm:"column:tahun"`
	Status       modelkeuangan.StatusTagihan `gorm:"column:status"`
	TotalDibayar float64                     `gorm:"column:total_dibayar"`
	CreatedAt    interface{}                 `gorm:"column:created_at"`
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// BeginTx memulai database transaction.
func (r *Repository) BeginTx() *gorm.DB {
	return r.db.Begin()
}

// ListTagihanSPP mengembalikan daftar tagihan SPP dengan JOIN kelas info,
// mendukung filter dan paginasi.
func (r *Repository) ListTagihanSPP(params pagination.Params, filter ListSPPFilter) ([]tagScan, int64, error) {
	baseQuery := r.db.Table("tagihans t").
		Select(`
			t.id,
			t.santri_id,
			u.name        AS nama_santri,
			s.nik         AS nik,
			COALESCE(k.nama_kelas, '') AS kelas_nama,
			t.jenis,
			t.nominal,
			t.bulan,
			t.tahun,
			t.status,
			COALESCE((
				SELECT SUM(p.jumlah_bayar)
				FROM pembayarans p
				WHERE p.tagihan_id = t.id
				  AND p.deleted_at IS NULL
			), 0) AS total_dibayar,
			t.created_at
		`).
		Joins("JOIN santris s ON s.id = t.santri_id AND s.deleted_at IS NULL").
		Joins("JOIN users u ON u.id = s.user_id AND u.deleted_at IS NULL").
		Joins(`LEFT JOIN LATERAL (
			SELECT k2.nama_kelas
			FROM kelas_santris ks
			JOIN kelas k2 ON k2.id = ks.kelas_id AND k2.deleted_at IS NULL
			WHERE ks.santri_id = t.santri_id
			  AND ks.deleted_at IS NULL
			ORDER BY ks.created_at DESC
			LIMIT 1
		) k ON true`).
		Where("t.deleted_at IS NULL")

	// Apply filters
	if filter.Jenis != "" {
		baseQuery = baseQuery.Where("t.jenis = ?", filter.Jenis)
	}
	if filter.Status != "" {
		baseQuery = baseQuery.Where("t.status = ?", filter.Status)
	}
	if filter.Bulan > 0 {
		baseQuery = baseQuery.Where("t.bulan = ?", filter.Bulan)
	}
	if filter.Tahun > 0 {
		baseQuery = baseQuery.Where("t.tahun = ?", filter.Tahun)
	}
	if filter.KelasID != nil {
		baseQuery = baseQuery.Where(`EXISTS (
			SELECT 1 FROM kelas_santris ks2
			WHERE ks2.santri_id = t.santri_id
			  AND ks2.kelas_id  = ?
			  AND ks2.deleted_at IS NULL
		)`, *filter.KelasID)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		baseQuery = baseQuery.Where("(u.name ILIKE ? OR s.nis ILIKE ?)", like, like)
	}

	// Count total (clone query without order/limit/offset)
	var total int64
	countQ := r.db.Table("(?) AS sub", baseQuery)
	if err := countQ.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch page
	var rows []tagScan
	orderCol := "t." + params.SortBy
	// Allowlist the sort column to prevent injection
	allowed := map[string]bool{
		"t.created_at": true, "t.updated_at": true, "t.nominal": true,
		"t.tahun": true, "t.bulan": true, "t.status": true,
	}
	if !allowed[orderCol] {
		orderCol = "t.created_at"
	}
	orderClause := fmt.Sprintf("%s %s", orderCol, params.Order)

	err := baseQuery.
		Order(orderClause).
		Limit(params.Limit).
		Offset(params.Offset()).
		Scan(&rows).Error

	return rows, total, err
}

// FindTagihanByID memuat satu tagihan beserta daftar pembayarannya.
func (r *Repository) FindTagihanByID(id uuid.UUID) (*modelkeuangan.Tagihan, error) {
	var t modelkeuangan.Tagihan
	err := r.db.
		Preload("Pembayaran").
		Preload("Santri.User").
		First(&t, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// CheckExists memeriksa apakah tagihan dengan kombinasi santriID+bulan+tahun+jenis sudah ada.
func (r *Repository) CheckExists(santriID uuid.UUID, bulan, tahun int, jenis string) (bool, error) {
	var count int64
	err := r.db.Model(&modelkeuangan.Tagihan{}).
		Where("santri_id = ? AND bulan = ? AND tahun = ? AND jenis = ? AND deleted_at IS NULL",
			santriID, bulan, tahun, jenis).
		Count(&count).Error
	return count > 0, err
}

// ListAktifSantri mengembalikan semua santri dengan status 'aktif'.
func (r *Repository) ListAktifSantri() ([]akademik.Santri, error) {
	var list []akademik.Santri
	err := r.db.
		Preload("User").
		Where("status = ?", akademik.StatusAktif).
		Find(&list).Error
	return list, err
}

// CreateTagihanBatch membuat banyak tagihan sekaligus dalam satu INSERT.
func (r *Repository) CreateTagihanBatch(tagihans []modelkeuangan.Tagihan) error {
	if len(tagihans) == 0 {
		return nil
	}
	return r.db.CreateInBatches(&tagihans, 100).Error
}

// SumPembayaran menghitung total pembayaran yang sudah dibayar untuk satu tagihan.
func (r *Repository) SumPembayaran(tagihanID uuid.UUID) (float64, error) {
	var sum float64
	err := r.db.Model(&modelkeuangan.Pembayaran{}).
		Where("tagihan_id = ? AND deleted_at IS NULL", tagihanID).
		Select("COALESCE(SUM(jumlah_bayar), 0)").
		Scan(&sum).Error
	return sum, err
}

// CreatePembayaranTx membuat pembayaran baru dalam sebuah transaksi.
func (r *Repository) CreatePembayaranTx(tx *gorm.DB, p *modelkeuangan.Pembayaran) error {
	return tx.Create(p).Error
}

// UpdateStatusTx memperbarui status tagihan dalam sebuah transaksi.
func (r *Repository) UpdateStatusTx(tx *gorm.DB, tagihanID uuid.UUID, status modelkeuangan.StatusTagihan) error {
	return tx.Model(&modelkeuangan.Tagihan{}).
		Where("id = ?", tagihanID).
		Update("status", status).Error
}

// GetRekapBulanan mengembalikan rekapitulasi tagihan per bulan untuk tahun dan jenis tertentu.
func (r *Repository) GetRekapBulanan(jenis string, tahun int) ([]RekapBulanan, error) {
	type rekapRow struct {
		Bulan         int     `gorm:"column:bulan"`
		Tahun         int     `gorm:"column:tahun"`
		TotalTagihan  float64 `gorm:"column:total_tagihan"`
		TotalMasuk    float64 `gorm:"column:total_masuk"`
		JumlahTagihan int64   `gorm:"column:jumlah_tagihan"`
		JumlahLunas   int64   `gorm:"column:jumlah_lunas"`
		JumlahBelum   int64   `gorm:"column:jumlah_belum"`
		JumlahCicilan int64   `gorm:"column:jumlah_cicilan"`
	}

	q := r.db.Table("tagihans t").
		Select(`
			t.bulan,
			t.tahun,
			SUM(t.nominal) AS total_tagihan,
			COALESCE(SUM(
				(SELECT SUM(p.jumlah_bayar)
				 FROM pembayarans p
				 WHERE p.tagihan_id = t.id AND p.deleted_at IS NULL)
			), 0) AS total_masuk,
			COUNT(*) AS jumlah_tagihan,
			COUNT(*) FILTER (WHERE t.status = 'lunas')       AS jumlah_lunas,
			COUNT(*) FILTER (WHERE t.status = 'belum_bayar') AS jumlah_belum,
			COUNT(*) FILTER (WHERE t.status = 'sebagian')    AS jumlah_cicilan
		`).
		Where("t.deleted_at IS NULL AND t.tahun = ?", tahun).
		Group("t.bulan, t.tahun").
		Order("t.bulan ASC")

	if jenis != "" {
		q = q.Where("t.jenis = ?", jenis)
	}

	var rows []rekapRow
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]RekapBulanan, 0, len(rows))
	for _, row := range rows {
		result = append(result, RekapBulanan{
			Bulan:         row.Bulan,
			Tahun:         row.Tahun,
			TotalTagihan:  row.TotalTagihan,
			TotalMasuk:    row.TotalMasuk,
			JumlahTagihan: row.JumlahTagihan,
			JumlahLunas:   row.JumlahLunas,
			JumlahBelum:   row.JumlahBelum,
			JumlahCicilan: row.JumlahCicilan,
		})
	}
	return result, nil
}

// ListUnpaid mengembalikan tagihan yang belum atau sebagian bayar untuk bulan/tahun tertentu.
func (r *Repository) ListUnpaid(jenis string, bulan, tahun int) ([]UnpaidItem, error) {
	type unpaidRow struct {
		TagihanID  uuid.UUID `gorm:"column:tagihan_id"`
		SantriID   uuid.UUID `gorm:"column:santri_id"`
		NamaSantri string    `gorm:"column:nama_santri"`
		NIK        string    `gorm:"column:nik"`
		KelasNama  string    `gorm:"column:kelas_nama"`
		Bulan      int       `gorm:"column:bulan"`
		Tahun      int       `gorm:"column:tahun"`
		Nominal    float64   `gorm:"column:nominal"`
		TotalBayar float64   `gorm:"column:total_bayar"`
		Status     string    `gorm:"column:status"`
	}

	q := r.db.Table("tagihans t").
		Select(`
			t.id           AS tagihan_id,
			t.santri_id,
			u.name         AS nama_santri,
			s.nis          AS nis,
			COALESCE(k.nama_kelas, '') AS kelas_nama,
			t.bulan,
			t.tahun,
			t.nominal,
			COALESCE((
				SELECT SUM(p.jumlah_bayar)
				FROM pembayarans p
				WHERE p.tagihan_id = t.id AND p.deleted_at IS NULL
			), 0) AS total_bayar,
			t.status
		`).
		Joins("JOIN santris s ON s.id = t.santri_id AND s.deleted_at IS NULL").
		Joins("JOIN users u ON u.id = s.user_id AND u.deleted_at IS NULL").
		Joins(`LEFT JOIN LATERAL (
			SELECT k2.nama_kelas
			FROM kelas_santris ks
			JOIN kelas k2 ON k2.id = ks.kelas_id AND k2.deleted_at IS NULL
			WHERE ks.santri_id = t.santri_id
			  AND ks.deleted_at IS NULL
			ORDER BY ks.created_at DESC
			LIMIT 1
		) k ON true`).
		Where("t.deleted_at IS NULL AND t.status IN ('belum_bayar','sebagian')")

	if jenis != "" {
		q = q.Where("t.jenis = ?", jenis)
	}
	if bulan > 0 {
		q = q.Where("t.bulan = ?", bulan)
	}
	if tahun > 0 {
		q = q.Where("t.tahun = ?", tahun)
	}
	q = q.Order("t.tahun ASC, t.bulan ASC, u.name ASC")

	var rows []unpaidRow
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]UnpaidItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, UnpaidItem{
			TagihanID:  row.TagihanID,
			SantriID:   row.SantriID,
			NamaSantri: row.NamaSantri,
			NIK:        row.NIK,
			KelasNama:  row.KelasNama,
			Bulan:      row.Bulan,
			Tahun:      row.Tahun,
			Nominal:    row.Nominal,
			Sisa:       row.Nominal - row.TotalBayar,
			Status:     row.Status,
		})
	}
	return result, nil
}

// GetKelasNama mengembalikan nama kelas terbaru untuk santri tertentu.
func (r *Repository) GetKelasNama(santriID uuid.UUID) (string, error) {
	var nama string
	err := r.db.Table("kelas_santris ks").
		Select("k.nama_kelas").
		Joins("JOIN kelas k ON k.id = ks.kelas_id AND k.deleted_at IS NULL").
		Where("ks.santri_id = ? AND ks.deleted_at IS NULL", santriID).
		Order("ks.created_at DESC").
		Limit(1).
		Scan(&nama).Error
	return nama, err
}

// ListTagihanSPPAll mengembalikan semua tagihan sesuai filter tanpa paginasi (untuk ekspor).
func (r *Repository) ListTagihanSPPAll(filter ListSPPFilter) ([]tagScan, error) {
	q := r.db.Table("tagihans t").
		Select(`
			t.id,
			t.santri_id,
			u.name        AS nama_santri,
			s.nik         AS nik,
			COALESCE(k.nama_kelas, '') AS kelas_nama,
			t.jenis,
			t.nominal,
			t.bulan,
			t.tahun,
			t.status,
			COALESCE((
				SELECT SUM(p.jumlah_bayar)
				FROM pembayarans p
				WHERE p.tagihan_id = t.id
				  AND p.deleted_at IS NULL
			), 0) AS total_dibayar,
			t.created_at
		`).
		Joins("JOIN santris s ON s.id = t.santri_id AND s.deleted_at IS NULL").
		Joins("JOIN users u ON u.id = s.user_id AND u.deleted_at IS NULL").
		Joins(`LEFT JOIN LATERAL (
			SELECT k2.nama_kelas
			FROM kelas_santris ks
			JOIN kelas k2 ON k2.id = ks.kelas_id AND k2.deleted_at IS NULL
			WHERE ks.santri_id = t.santri_id
			  AND ks.deleted_at IS NULL
			ORDER BY ks.created_at DESC
			LIMIT 1
		) k ON true`).
		Where("t.deleted_at IS NULL")

	if filter.Jenis != "" {
		q = q.Where("t.jenis = ?", filter.Jenis)
	}
	if filter.Status != "" {
		q = q.Where("t.status = ?", filter.Status)
	}
	if filter.Bulan > 0 {
		q = q.Where("t.bulan = ?", filter.Bulan)
	}
	if filter.Tahun > 0 {
		q = q.Where("t.tahun = ?", filter.Tahun)
	}
	if filter.KelasID != nil {
		q = q.Where(`EXISTS (
			SELECT 1 FROM kelas_santris ks2
			WHERE ks2.santri_id = t.santri_id
			  AND ks2.kelas_id  = ?
			  AND ks2.deleted_at IS NULL
		)`, *filter.KelasID)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		q = q.Where("(u.name ILIKE ? OR s.nis ILIKE ?)", like, like)
	}

	var rows []tagScan
	err := q.Order("t.tahun ASC, t.bulan ASC, u.name ASC").
		Limit(10000).
		Scan(&rows).Error
	return rows, err
}
