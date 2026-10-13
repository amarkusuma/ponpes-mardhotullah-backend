package auth

import "github.com/google/uuid"

// ── Request DTOs ────────────────────────────────────────────────────────────

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ── Response DTOs ───────────────────────────────────────────────────────────

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // detik
}

type LoginResponse struct {
	User  UserInfo  `json:"user"`
	Token TokenPair `json:"token"`
}

type UserInfo struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Avatar   string    `json:"avatar,omitempty"`
	IsActive bool      `json:"is_active"`
	Roles    []string  `json:"roles"`
}
