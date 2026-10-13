package transkrip

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/simpeg/backend/pkg/apperror"
	"github.com/simpeg/backend/pkg/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// POST /api/transkrip/generate/:santri_id
func (h *Handler) Generate(c *gin.Context) {
	santriID, err := parseUUID(c, "santri_id")
	if err != nil {
		response.Error(c, err)
		return
	}

	filePath, err := h.svc.Generate(santriID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "transkrip berhasil digenerate", gin.H{
		"file_path": filePath,
	})
}

// GET /api/transkrip/download/:id
func (h *Handler) Download(c *gin.Context) {
	raporID, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}

	rapor, err := h.svc.GetRaporByID(raporID)
	if err != nil {
		response.Error(c, err)
		return
	}

	// Pastikan file ada
	if _, statErr := os.Stat(rapor.FilePath); os.IsNotExist(statErr) {
		response.Error(c, apperror.ErrNotFoundMsg("file transkrip"))
		return
	}

	// Set header download
	fileName := "transkrip_" + rapor.Santri.NIK + ".txt"
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Status(http.StatusOK)
	c.File(rapor.FilePath)
}

// RegisterRoutes mendaftarkan route transkrip.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/generate/:santri_id", h.Generate)
	rg.GET("/download/:id", h.Download)
}

func parseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		return uuid.Nil, apperror.ErrValidation(param + " tidak valid")
	}
	return id, nil
}
