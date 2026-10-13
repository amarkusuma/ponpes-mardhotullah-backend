package akademik

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/simpeg/backend/pkg/apperror"
	mw "github.com/simpeg/backend/pkg/middleware"
	"github.com/simpeg/backend/pkg/pagination"
	"github.com/simpeg/backend/pkg/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ── Admin: TahunAjaran ────────────────────────────────────────────────────────

func (h *Handler) ListTahunAjaran(c *gin.Context) {
	list, err := h.svc.ListTahunAjaran()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "daftar tahun ajaran", list)
}

// ── Admin: Kelas ──────────────────────────────────────────────────────────────

func (h *Handler) ListKelas(c *gin.Context) {
	p := pagination.Parse(c)
	var taID *uuid.UUID
	if raw := c.Query("tahun_ajaran_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err == nil {
			taID = &id
		}
	}
	list, total, err := h.svc.ListKelas(p, taID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.WithPagination(c, "daftar kelas", list, p.Page, p.Limit, total)
}

func (h *Handler) GetKelas(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	k, err := h.svc.GetKelas(id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "detail kelas", k)
}

func (h *Handler) CreateKelas(c *gin.Context) {
	var req CreateKelasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}
	k, err := h.svc.CreateKelas(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, "kelas berhasil dibuat", k)
}

func (h *Handler) UpdateKelas(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req UpdateKelasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}
	k, err := h.svc.UpdateKelas(id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "kelas berhasil diperbarui", k)
}

func (h *Handler) DeleteKelas(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.DeleteKelas(id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// ── Admin: MataPelajaran ──────────────────────────────────────────────────────

func (h *Handler) ListMapel(c *gin.Context) {
	p := pagination.Parse(c)
	list, total, err := h.svc.ListMapel(p)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.WithPagination(c, "daftar mata pelajaran", list, p.Page, p.Limit, total)
}

func (h *Handler) CreateMapel(c *gin.Context) {
	var req CreateMapelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}
	m, err := h.svc.CreateMapel(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, "mata pelajaran berhasil dibuat", m)
}

func (h *Handler) UpdateMapel(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req UpdateMapelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}
	m, err := h.svc.UpdateMapel(id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "mata pelajaran berhasil diperbarui", m)
}

func (h *Handler) DeleteMapel(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.DeleteMapel(id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// ── Admin: JadwalPelajaran ────────────────────────────────────────────────────

func (h *Handler) ListJadwalKelas(c *gin.Context) {
	kelasID, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	list, err := h.svc.GetJadwalKelas(kelasID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "jadwal kelas", list)
}

func (h *Handler) CreateJadwal(c *gin.Context) {
	var req CreateJadwalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}
	j, err := h.svc.CreateJadwal(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, "jadwal berhasil dibuat", j)
}

func (h *Handler) DeleteJadwal(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.DeleteJadwal(id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// ── Admin: Profil Guru ────────────────────────────────────────────────────────

// GET /api/admin/guru
func (h *Handler) AdminListGuru(c *gin.Context) {
	list, err := h.svc.ListGuruProfil()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "daftar profil guru", list)
}

// PUT /api/admin/guru/:user_id
func (h *Handler) AdminUpdateGuru(c *gin.Context) {
	userID, err := parseUUID(c, "user_id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req UpdateGuruRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}
	g, err := h.svc.UpdateGuruProfil(userID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "profil guru berhasil diperbarui", g)
}

// ── Admin: Profil Santri ──────────────────────────────────────────────────────

// GET /api/admin/santri
func (h *Handler) AdminListSantri(c *gin.Context) {
	list, err := h.svc.ListSantriProfil()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "daftar profil santri", list)
}

// PUT /api/admin/santri/:user_id
func (h *Handler) AdminUpdateSantri(c *gin.Context) {
	userID, err := parseUUID(c, "user_id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req UpdateSantriRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}
	s, err := h.svc.UpdateSantriProfil(userID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "profil santri berhasil diperbarui", s)
}

// ── Admin: Profil Wali ────────────────────────────────────────────────────────

// GET /api/admin/wali
func (h *Handler) AdminListWali(c *gin.Context) {
	list, err := h.svc.ListWaliProfil()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "daftar profil wali", list)
}

// PUT /api/admin/wali/:user_id
func (h *Handler) AdminUpdateWali(c *gin.Context) {
	userID, err := parseUUID(c, "user_id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req UpdateWaliRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}
	w, err := h.svc.UpdateWaliProfil(userID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "profil wali berhasil diperbarui", w)
}

// POST /api/admin/wali/:user_id/link-santri  — body: {"nis":"S2024001"}
func (h *Handler) AdminLinkSantri(c *gin.Context) {
	waliUserID, err := parseUUID(c, "user_id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req LinkSantriRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}
	if err := h.svc.LinkSantriToWali(waliUserID, req.NIK); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "santri berhasil dihubungkan ke wali", nil)
}

// DELETE /api/admin/wali/:user_id/santri/:santri_id
func (h *Handler) AdminUnlinkSantri(c *gin.Context) {
	waliUserID, err := parseUUID(c, "user_id")
	if err != nil {
		response.Error(c, err)
		return
	}
	santriID, err := parseUUID(c, "santri_id")
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.UnlinkSantriFromWali(waliUserID, santriID); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// ── Guru Endpoints ────────────────────────────────────────────────────────────

func (h *Handler) GuruJadwalMengajar(c *gin.Context) {
	user := mw.MustCurrentUser(c)
	list, err := h.svc.GetJadwalGuru(user.ID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "jadwal mengajar", list)
}

func (h *Handler) GuruDaftarSantri(c *gin.Context) {
	kelasID, err := parseUUID(c, "kelas_id")
	if err != nil {
		response.Error(c, err)
		return
	}
	list, err := h.svc.GetDaftarSantriByKelas(kelasID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "daftar santri kelas", list)
}

func (h *Handler) GuruCreatePenilaian(c *gin.Context) {
	var req CreatePenilaianRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}
	p, err := h.svc.CreatePenilaian(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, "penilaian berhasil disimpan", p)
}

func (h *Handler) GuruUpdatePenilaian(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req UpdatePenilaianRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.ErrValidation(err.Error()))
		return
	}
	p, err := h.svc.UpdatePenilaian(id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "penilaian berhasil diperbarui", p)
}

func (h *Handler) GuruRekapNilai(c *gin.Context) {
	kelasID, err := parseUUID(c, "kelas_id")
	if err != nil {
		response.Error(c, err)
		return
	}
	mapelID, err := parseUUID(c, "mapel_id")
	if err != nil {
		response.Error(c, err)
		return
	}
	rekap, err := h.svc.RekapNilai(kelasID, mapelID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "rekap nilai", rekap)
}

// ── Santri Endpoints ──────────────────────────────────────────────────────────

func (h *Handler) SantriProfil(c *gin.Context) {
	user := mw.MustCurrentUser(c)
	profil, err := h.svc.GetProfilSantri(user.ID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "profil santri", profil)
}

func (h *Handler) SantriJadwal(c *gin.Context) {
	raw := c.Query("kelas_id")
	if raw == "" {
		response.Error(c, apperror.ErrValidation("kelas_id wajib diisi"))
		return
	}
	kelasID, err := uuid.Parse(raw)
	if err != nil {
		response.Error(c, apperror.ErrValidation("kelas_id tidak valid"))
		return
	}
	list, err := h.svc.GetJadwalKelas(kelasID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "jadwal pelajaran", list)
}

func (h *Handler) SantriNilai(c *gin.Context) {
	user := mw.MustCurrentUser(c)
	list, err := h.svc.GetNilaiSantri(user.ID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "nilai santri", list)
}

// ── Wali Endpoints ────────────────────────────────────────────────────────────

func (h *Handler) WaliProfilAnak(c *gin.Context) {
	user := mw.MustCurrentUser(c)
	list, err := h.svc.GetWaliSantri(user.ID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "profil anak", list)
}

func (h *Handler) WaliNilaiAnak(c *gin.Context) {
	santriUserID, err := parseUUID(c, "santri_id")
	if err != nil {
		response.Error(c, err)
		return
	}
	list, err := h.svc.GetNilaiSantri(santriUserID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "nilai anak", list)
}

// ── Route Registration ────────────────────────────────────────────────────────

func (h *Handler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	rg.GET("/tahun-ajaran", h.ListTahunAjaran)

	kelas := rg.Group("/kelas")
	kelas.GET("", h.ListKelas)
	kelas.GET("/:id", h.GetKelas)
	kelas.POST("", h.CreateKelas)
	kelas.PUT("/:id", h.UpdateKelas)
	kelas.DELETE("/:id", h.DeleteKelas)
	kelas.GET("/:id/jadwal", h.ListJadwalKelas)

	mapel := rg.Group("/mata-pelajaran")
	mapel.GET("", h.ListMapel)
	mapel.POST("", h.CreateMapel)
	mapel.PUT("/:id", h.UpdateMapel)
	mapel.DELETE("/:id", h.DeleteMapel)

	jadwal := rg.Group("/jadwal")
	jadwal.POST("", h.CreateJadwal)
	jadwal.DELETE("/:id", h.DeleteJadwal)

	// Profil endpoints
	guru := rg.Group("/guru")
	guru.GET("", h.AdminListGuru)
	guru.PUT("/:user_id", h.AdminUpdateGuru)

	santri := rg.Group("/santri")
	santri.GET("", h.AdminListSantri)
	santri.PUT("/:user_id", h.AdminUpdateSantri)

	wali := rg.Group("/wali")
	wali.GET("", h.AdminListWali)
	wali.PUT("/:user_id", h.AdminUpdateWali)
	wali.POST("/:user_id/link-santri", h.AdminLinkSantri)
	wali.DELETE("/:user_id/santri/:santri_id", h.AdminUnlinkSantri)
}

func (h *Handler) RegisterGuruRoutes(rg *gin.RouterGroup) {
	rg.GET("/jadwal-mengajar", h.GuruJadwalMengajar)
	rg.GET("/daftar-santri/:kelas_id", h.GuruDaftarSantri)
	rg.POST("/penilaian", h.GuruCreatePenilaian)
	rg.PUT("/penilaian/:id", h.GuruUpdatePenilaian)
	rg.GET("/rekap-nilai/:kelas_id/:mapel_id", h.GuruRekapNilai)
}

func (h *Handler) RegisterSantriRoutes(rg *gin.RouterGroup) {
	rg.GET("/profil", h.SantriProfil)
	rg.GET("/jadwal-pelajaran", h.SantriJadwal)
	rg.GET("/nilai", h.SantriNilai)
}

func (h *Handler) RegisterWaliRoutes(rg *gin.RouterGroup) {
	rg.GET("/profil-anak", h.WaliProfilAnak)
	rg.GET("/nilai-anak/:santri_id", h.WaliNilaiAnak)
}

func parseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		return uuid.Nil, apperror.ErrValidation(param + " tidak valid")
	}
	return id, nil
}
