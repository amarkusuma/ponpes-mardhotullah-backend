package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base menjadi embedded struct untuk semua model,
// menggunakan UUID v4 sebagai primary key.
type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                                            json:"-"`
}

// BeforeCreate di-promote ke semua model yang embed Base.
// Menjamin UUID terisi dari sisi Go tanpa bergantung extension PostgreSQL.
func (b *Base) BeforeCreate(_ *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
