package logger

import (
  "go.uber.org/zap"
  "go.uber.org/zap/zapcore"

  "github.com/simpeg/backend/internal/config"
)

func New(cfg *config.Config) *zap.Logger {
  level := zapcore.InfoLevel
  switch cfg.LogLevel {
  case "debug":
    level = zapcore.DebugLevel
  case "warn":
    level = zapcore.WarnLevel
  case "error":
    level = zapcore.ErrorLevel
  }

  config := zap.NewProductionConfig()
  config.Level = zap.NewAtomicLevelAt(level)
  config.EncoderConfig.TimeKey = "timestamp"
  config.EncoderConfig.MessageKey = "message"
  config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

  logger, err := config.Build()
  if err != nil {
    panic(err)
  }

  return logger
}
