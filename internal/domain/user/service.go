package user

import (
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/simpeg/backend/internal/model"
	"github.com/simpeg/backend/pkg/apperror"
	"github.com/simpeg/backend/pkg/pagination"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ── Dashboard ────────────────────────────────────────────────────────────────

func (s *Service) Dashboard() (*DashboardResponse, error) {
	santri, _ := s.repo.CountSantri()
	guru, _ := s.repo.CountGuru()
	wali, _ := s.repo.CountWali()
	total, lunas, belum, _ := s.repo.TagihanStatsBulanIni()

	return &DashboardResponse{
		TotalSantri:          santri,
		TotalGuru:            guru,
		TotalWali:            wali,
		TotalTagihanBulanIni: total,
		TagihanLunas:         lunas,
		TagihanBelumBayar:    belum,
	}, nil
}

// ── User CRUD ────────────────────────────────────────────────────────────────

func (s *Service) List(params pagination.Params, filter ListUsersFilter) ([]UserResponse, int64, error) {
	users, total, err := s.repo.List(params, filter)
	if err != nil {
		return nil, 0, apperror.ErrInternal
	}

	result := make([]UserResponse, 0, len(users))
	for _, u := range users {
		u := u
		result = append(result, toUserResponse(&u))
	}
	return result, total, nil
}

func (s *Service) Get(id uuid.UUID) (*UserResponse, error) {
	u, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.ErrNotFoundMsg("user")
		}
		return nil, apperror.ErrInternal
	}
	r := toUserResponse(u)
	return &r, nil
}

func (s *Service) Create(req CreateUserRequest) (*UserResponse, error) {
	// Cek duplikat email
	if _, err := s.repo.FindByEmail(strings.ToLower(req.Email)); err == nil {
		return nil, apperror.ErrConflictMsg("email")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.ErrInternal
	}

	u := &model.User{
		Name:     req.Name,
		Email:    strings.ToLower(req.Email),
		Password: string(hash),
		Avatar:   req.Avatar,
		IsActive: true,
	}

	if err := s.repo.Create(u); err != nil {
		return nil, apperror.ErrInternal
	}

	// Assign roles jika ada
	if len(req.RoleIDs) > 0 {
		roleUUIDs := make([]uuid.UUID, 0, len(req.RoleIDs))
		for _, id := range req.RoleIDs {
			uid, err := uuid.Parse(id)
			if err == nil {
				roleUUIDs = append(roleUUIDs, uid)
			}
		}
		s.repo.AssignRoles(u.ID, roleUUIDs) //nolint:errcheck
	}

	return s.Get(u.ID)
}

func (s *Service) Update(id uuid.UUID, req UpdateUserRequest) (*UserResponse, error) {
	u, err := s.repo.FindByID(id)
	if err != nil {
		return nil, apperror.ErrNotFoundMsg("user")
	}

	if req.Name != "" {
		u.Name = req.Name
	}
	if req.Avatar != "" {
		u.Avatar = req.Avatar
	}
	if req.IsActive != nil {
		u.IsActive = *req.IsActive
	}

	if err := s.repo.Update(u); err != nil {
		return nil, apperror.ErrInternal
	}

	r := toUserResponse(u)
	return &r, nil
}

func (s *Service) Delete(id uuid.UUID) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return apperror.ErrNotFoundMsg("user")
	}
	return s.repo.Delete(id)
}

func (s *Service) AssignRoles(userID uuid.UUID, req AssignRoleRequest) error {
	roleUUIDs := make([]uuid.UUID, 0, len(req.RoleIDs))
	for _, id := range req.RoleIDs {
		uid, err := uuid.Parse(id)
		if err != nil {
			return apperror.ErrValidation("role_id tidak valid: " + id)
		}
		roleUUIDs = append(roleUUIDs, uid)
	}
	return s.repo.AssignRoles(userID, roleUUIDs)
}

func (s *Service) ListRoles() ([]model.Role, error) {
	return s.repo.ListRoles()
}
