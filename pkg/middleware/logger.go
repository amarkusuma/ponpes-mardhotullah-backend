package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// sensitiveFields adalah field yang tidak boleh di-log.
var sensitiveFields = map[string]bool{
	"password": true, "token": true, "refresh_token": true,
	"access_token": true, "secret": true,
}

type logEntry struct {
	Timestamp    string `json:"timestamp"`
	Method       string `json:"method"`
	Path         string `json:"path"`
	UserID       string `json:"user_id,omitempty"`
	RequestBody  string `json:"request_body,omitempty"`
	Status       int    `json:"status"`
	ResponseTime string `json:"response_time"`
	ClientIP     string `json:"client_ip"`
}

// bodyWriter wraps gin.ResponseWriter untuk menangkap status code.
type bodyWriter struct {
	gin.ResponseWriter
	status int
}

func (bw *bodyWriter) WriteHeader(status int) {
	bw.status = status
	bw.ResponseWriter.WriteHeader(status)
}

// RequestLogger adalah middleware structured logging yang menyimpan log ke
// Zap (stdout) dan Redis (list harian dengan TTL 7 hari).
func RequestLogger(logger *zap.Logger, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		bw := &bodyWriter{ResponseWriter: c.Writer}
		c.Writer = bw

		// Baca dan kembalikan request body
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		c.Next()

		duration := time.Since(start)
		status := bw.status
		if status == 0 {
			status = c.Writer.Status()
		}

		userID := ""
		if u := CurrentUser(c); u != nil {
			userID = u.ID.String()
		}

		entry := logEntry{
			Timestamp:    start.UTC().Format(time.RFC3339),
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			UserID:       userID,
			RequestBody:  sanitizeBody(bodyBytes),
			Status:       status,
			ResponseTime: duration.String(),
			ClientIP:     c.ClientIP(),
		}

		// Zap logging
		logger.Info("http request",
			zap.String("method", entry.Method),
			zap.String("path", entry.Path),
			zap.String("user_id", entry.UserID),
			zap.Int("status", entry.Status),
			zap.String("response_time", entry.ResponseTime),
			zap.String("client_ip", entry.ClientIP),
		)

		// Async: simpan ke Redis
		if redisClient != nil {
			go storeLogToRedis(redisClient, entry)
		}
	}
}

// ZapLogger adalah middleware logging sederhana tanpa Redis (fallback).
func ZapLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info("http request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}

func storeLogToRedis(client *redis.Client, entry logEntry) {
	ctx := context.Background()
	key := fmt.Sprintf("reqlog:%s", time.Now().UTC().Format("2006-01-02"))

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	pipe := client.Pipeline()
	pipe.RPush(ctx, key, string(data))
	pipe.Expire(ctx, key, 7*24*time.Hour)
	pipe.Exec(ctx) //nolint:errcheck
}

func sanitizeBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}

	// Hanya proses JSON body
	var m map[string]interface{}
	if err := json.Unmarshal(body, &m); err != nil {
		return "[non-json body]"
	}

	for key := range m {
		if sensitiveFields[strings.ToLower(key)] {
			m[key] = "[REDACTED]"
		}
	}

	sanitized, _ := json.Marshal(m)
	return string(sanitized)
}
