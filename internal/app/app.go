// Package app adalah entry point bootstrap aplikasi Pesantren Management System.
package app

import (
	"context"

	"go.uber.org/zap"

	"github.com/simpeg/backend/internal/config"
	internaldb "github.com/simpeg/backend/internal/database"
	"github.com/simpeg/backend/internal/server"
	"github.com/simpeg/backend/pkg/cache"
	"github.com/simpeg/backend/pkg/database"
	"github.com/simpeg/backend/pkg/logger"
	"github.com/simpeg/backend/pkg/storage"
)

func Run() {
	cfg := config.LoadConfig()

	zapLogger := logger.New(cfg)
	defer zapLogger.Sync() //nolint:errcheck

	db, err := database.NewPostgres(cfg)
	if err != nil {
		zapLogger.Fatal("gagal koneksi database", zap.Error(err))
	}

	if err := internaldb.Migrate(db); err != nil {
		zapLogger.Fatal("migrasi database gagal", zap.Error(err))
	}

	if err := internaldb.Seed(db); err != nil {
		zapLogger.Warn("seeding gagal", zap.Error(err))
	}

	redisClient := cache.NewRedis(cfg)

	store, err := storage.NewMinio(context.Background(), cfg)
	if err != nil {
		zapLogger.Fatal("gagal koneksi minio", zap.Error(err))
	}

	r := server.NewRouter(server.Deps{
		Config:  cfg,
		Logger:  zapLogger,
		DB:      db,
		Redis:   redisClient,
		Storage: store,
	})

	zapLogger.Info("server berjalan", zap.String("addr", cfg.ServerAddress))
	if err := r.Run(cfg.ServerAddress); err != nil {
		zapLogger.Fatal("server berhenti", zap.Error(err))
	}
}
