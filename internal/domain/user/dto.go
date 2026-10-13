package user

import (
	"github.com/google/uuid"
	"github.com/simpeg/backend/internal/model"
)

// ── Request DTOs ────────────────────────────────────────────────────────────

type CreateUserRequest struct {
	Name     string   `json:"name"     binding:"required,min=2"`
	Email    string   `json:"email"    binding:"required,email"`
	Password string   `json:"password" binding:"required,min=8"`
	Avatar   string   `json:"avatar"`
	RoleIDs  []string `json:"role_ids"`
}

type UpdateUserRequest struct {
	Name     string `json:"name"   binding:"omitempty,min=2"`
	Avatar   string `json:"avatar"`
	IsActive *bool  `json:"is_active"`
}

type AssignRoleRequest struct {
	RoleIDs []string `json:"role_ids" binding:"required,min=1"`
}

type ListUsersFilter struct {
	Search   string
	RoleName string
	IsActive *bool
}

// ── Response DTOs ───────────────────────────────────────────────────────────

type UserResponse struct {
	ID       uuid.UUID      `json:"id"`
	Name     string         `json:"name"`
	Email    string         `json:"email"`
	Avatar   string         `json:"avatar,omitempty"`
	IsActive bool           `json:"is_active"`
	Roles    []RoleResponse `json:"roles"`
}

type RoleResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type DashboardResponse struct {
	TotalSantri          int64   `json:"total_santri"`
	TotalGuru            int64   `json:"total_guru"`
	TotalWali            int64   `json:"total_wali"`
	TotalTagihanBulanIni float64 `json:"total_tagihan_bulan_ini"`
	TagihanLunas         int64   `json:"tagihan_lunas"`
	TagihanBelumBayar    int64   `json:"tagihan_belum_bayar"`
}

func toUserResponse(u *model.User) UserResponse {
	roles := make([]RoleResponse, 0, len(u.Roles))
	for _, r := range u.Roles {
		roles = append(roles, RoleResponse{ID: r.ID, Name: r.Name})
	}
	return UserResponse{
		ID:       u.ID,
		Name:     u.Name,
		Email:    u.Email,
		Avatar:   u.Avatar,
		IsActive: u.IsActive,
		Roles:    roles,
	}
}
