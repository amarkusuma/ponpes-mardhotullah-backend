package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/simpeg/backend/internal/model"
)

const refreshTokenPrefix = "rt:"

type Repository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewRepository(db *gorm.DB, redis *redis.Client) *Repository {
	return &Repository{db: db, redis: redis}
}

// FindUserByEmail mencari user beserta roles dan permissions-nya.
func (r *Repository) FindUserByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.
		Preload("Roles.Permissions").
		Where("email = ? AND is_active = true", email).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUserByID mencari user berdasarkan UUID.
func (r *Repository) FindUserByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.
		Preload("Roles.Permissions").
		Where("id = ? AND is_active = true", id).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// SaveRefreshToken menyimpan refresh token ke Redis dengan TTL.
func (r *Repository) SaveRefreshToken(token string, userID uuid.UUID, ttl time.Duration) error {
	key := refreshTokenPrefix + token
	return r.redis.Set(context.Background(), key, userID.String(), ttl).Err()
}

// GetRefreshToken mengambil userID yang terkait dengan refresh token.
func (r *Repository) GetRefreshToken(token string) (uuid.UUID, error) {
	key := refreshTokenPrefix + token
	val, err := r.redis.Get(context.Background(), key).Result()
	if err != nil {
		return uuid.Nil, fmt.Errorf("refresh token tidak valid atau sudah kadaluarsa")
	}
	return uuid.Parse(val)
}

// DeleteRefreshToken menghapus refresh token dari Redis (logout / rotation).
func (r *Repository) DeleteRefreshToken(token string) error {
	key := refreshTokenPrefix + token
	return r.redis.Del(context.Background(), key).Err()
}
