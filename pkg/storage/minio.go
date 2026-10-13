// Package storage menyediakan object storage berbasis MinIO (S3-compatible).
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/simpeg/backend/internal/config"
)

// ErrNotFound dikembalikan ketika object tidak ada di bucket.
var ErrNotFound = errors.New("object tidak ditemukan")

type Storage struct {
	client *minio.Client
	bucket string
}

// NewMinio membuat client MinIO dan memastikan bucket sudah tersedia.
func NewMinio(ctx context.Context, cfg *config.Config) (*Storage, error) {
	client, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio client: %w", err)
	}

	exists, err := client.BucketExists(ctx, cfg.MinioBucket)
	if err != nil {
		return nil, fmt.Errorf("cek bucket %q: %w", cfg.MinioBucket, err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.MinioBucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("buat bucket %q: %w", cfg.MinioBucket, err)
		}
	}

	return &Storage{client: client, bucket: cfg.MinioBucket}, nil
}

// Put mengunggah object. size -1 jika ukuran tidak diketahui.
func (s *Storage) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

// Get mengambil object beserta metadata-nya. Caller wajib menutup reader.
func (s *Storage) Get(ctx context.Context, key string) (io.ReadCloser, minio.ObjectInfo, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, minio.ObjectInfo{}, err
	}

	info, err := obj.Stat()
	if err != nil {
		obj.Close()
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return nil, minio.ObjectInfo{}, ErrNotFound
		}
		return nil, minio.ObjectInfo{}, err
	}

	return obj, info, nil
}

// Remove menghapus object.
func (s *Storage) Remove(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}
