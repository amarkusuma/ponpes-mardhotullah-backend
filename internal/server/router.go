// Package server menyiapkan seluruh routing Gin dengan middleware yang sesuai.
package server

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/simpeg/backend/internal/config"
	authDomain "github.com/simpeg/backend/internal/domain/auth"
	akademikDomain "github.com/simpeg/backend/internal/domain/akademik"
	keuanganDomain "github.com/simpeg/backend/internal/domain/keuangan"
	kontenDomain "github.com/simpeg/backend/internal/domain/konten"
	sppDomain "github.com/simpeg/backend/internal/domain/spp"
	transkripDomain "github.com/simpeg/backend/internal/domain/transkrip"
	userDomain "github.com/simpeg/backend/internal/domain/user"
	mw "github.com/simpeg/backend/pkg/middleware"
	"github.com/simpeg/backend/pkg/storage"
)

// Deps adalah kumpulan dependency yang dibutuhkan router.
type Deps struct {
	Config   *config.Config
	Logger   *zap.Logger
	DB       *gorm.DB
	Redis    *redis.Client
	Storage  *storage.Storage
}

// NewRouter membangun gin.Engine dengan semua route terdaftar.
func NewRouter(deps Deps) *gin.Engine {
	// ── Bootstrap handlers ─────────────────────────────────────────────────
	// Auth
	authRepo    := authDomain.NewRepository(deps.DB, deps.Redis)
	authSvc     := authDomain.NewService(authRepo, deps.Config.JWTSecret, deps.Config.AccessTokenTTL, deps.Config.RefreshTokenTTL)
	authHandler := authDomain.NewHandler(authSvc)

	// User / Admin
	userRepo    := userDomain.NewRepository(deps.DB)
	userSvc     := userDomain.NewService(userRepo)
	userHandler := userDomain.NewHandler(userSvc)

	// Akademik
	akademikRepo    := akademikDomain.NewRepository(deps.DB)
	akademikSvc     := akademikDomain.NewService(akademikRepo)
	akademikHandler := akademikDomain.NewHandler(akademikSvc)

	// Keuangan
	keuanganRepo    := keuanganDomain.NewRepository(deps.DB)
	keuanganSvc     := keuanganDomain.NewService(keuanganRepo)
	keuanganHandler := keuanganDomain.NewHandler(keuanganSvc)

	// SPP
	sppRepo    := sppDomain.NewRepository(deps.DB)
	sppSvc     := sppDomain.NewService(sppRepo)
	sppHandler := sppDomain.NewHandler(sppSvc)

	// Konten
	kontenRepo    := kontenDomain.NewRepository(deps.DB)
	kontenSvc     := kontenDomain.NewService(kontenRepo)
	kontenHandler := kontenDomain.NewHandler(kontenSvc, deps.Storage, deps.Redis)

	// Transkrip
	transkripSvc     := transkripDomain.NewService(deps.DB, deps.Storage)
	transkripHandler := transkripDomain.NewHandler(transkripSvc)

	// ── Middleware ─────────────────────────────────────────────────────────
	authMW := mw.JWTAuth(deps.Config.JWTSecret)

	// ── Engine ─────────────────────────────────────────────────────────────
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(mw.RequestLogger(deps.Logger, deps.Redis))

	api := r.Group("/api")

	// ── Publik (tanpa auth) ────────────────────────────────────────────────
	publik := api.Group("/publik")
	kontenHandler.RegisterPublikRoutes(publik)

	// ── Auth ───────────────────────────────────────────────────────────────
	authHandler.RegisterRoutes(api, authMW)

	// ── Admin ──────────────────────────────────────────────────────────────
	admin := api.Group("/admin",
		authMW,
		mw.RequireRole("admin"),
	)
	userHandler.RegisterRoutes(admin)
	akademikHandler.RegisterAdminRoutes(admin)
	keuanganHandler.RegisterAdminRoutes(admin)
	kontenHandler.RegisterAdminRoutes(admin)
	sppHandler.RegisterRoutes(admin.Group("/spp"))

	// ── Guru ───────────────────────────────────────────────────────────────
	guru := api.Group("/guru",
		authMW,
		mw.RequireRole("admin", "guru"),
	)
	akademikHandler.RegisterGuruRoutes(guru)

	// ── Santri ─────────────────────────────────────────────────────────────
	santri := api.Group("/santri",
		authMW,
		mw.RequireRole("admin", "santri"),
	)
	akademikHandler.RegisterSantriRoutes(santri)
	keuanganHandler.RegisterSantriRoutes(santri)

	// ── Wali ───────────────────────────────────────────────────────────────
	wali := api.Group("/wali",
		authMW,
		mw.RequireRole("admin", "wali"),
	)
	akademikHandler.RegisterWaliRoutes(wali)
	keuanganHandler.RegisterWaliRoutes(wali)

	// ── Transkrip ──────────────────────────────────────────────────────────
	transkrip := api.Group("/transkrip",
		authMW,
		mw.RequireRole("admin", "guru"),
	)
	transkripHandler.RegisterRoutes(transkrip)

	// File upload (di-stream dari MinIO, contoh: /uploads/galeri/xxx.png)
	r.GET("/uploads/*key", serveUpload(deps.Storage))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}
