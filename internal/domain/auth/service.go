package auth

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/simpeg/backend/internal/model"
	"github.com/simpeg/backend/pkg/apperror"
)

type Service struct {
	repo            *Repository
	jwtSecret       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewService(repo *Repository, jwtSecret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		repo:            repo,
		jwtSecret:       jwtSecret,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

// Login memvalidasi kredensial dan mengembalikan token pair.
func (s *Service) Login(req LoginRequest) (*LoginResponse, error) {
	user, err := s.repo.FindUserByEmail(req.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.ErrUnauthorized
		}
		return nil, apperror.ErrInternal
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, apperror.ErrUnauthorized
	}

	pair, err := s.generateTokenPair(user)
	if err != nil {
		return nil, apperror.ErrInternal
	}

	return &LoginResponse{
		User:  toUserInfo(user),
		Token: *pair,
	}, nil
}

// Logout menginvalidasi refresh token.
func (s *Service) Logout(refreshToken string) error {
	if err := s.repo.DeleteRefreshToken(refreshToken); err != nil {
		return apperror.ErrUnauthorized
	}
	return nil
}

// Refresh melakukan token rotation: hapus token lama, terbitkan token baru.
func (s *Service) Refresh(refreshToken string) (*TokenPair, error) {
	userID, err := s.repo.GetRefreshToken(refreshToken)
	if err != nil {
		return nil, apperror.ErrUnauthorized
	}

	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return nil, apperror.ErrUnauthorized
	}

	// Hapus token lama (rotation)
	s.repo.DeleteRefreshToken(refreshToken) //nolint:errcheck

	pair, err := s.generateTokenPair(user)
	if err != nil {
		return nil, apperror.ErrInternal
	}

	return pair, nil
}

// Me mengembalikan profil user berdasarkan ID dari JWT.
func (s *Service) Me(userID uuid.UUID) (*model.User, error) {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return nil, apperror.ErrNotFoundMsg("user")
	}
	return user, nil
}

// ── private helpers ──────────────────────────────────────────────────────────

func (s *Service) generateTokenPair(user *model.User) (*TokenPair, error) {
	accessToken, err := s.issueAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	if err := s.repo.SaveRefreshToken(refreshToken, user.ID, s.refreshTokenTTL); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.accessTokenTTL.Seconds()),
	}, nil
}

func (s *Service) issueAccessToken(user *model.User) (string, error) {
	roleNames := make([]string, 0, len(user.Roles))
	permNames := make([]string, 0)
	seen := make(map[string]bool)

	for _, role := range user.Roles {
		roleNames = append(roleNames, role.Name)
		for _, perm := range role.Permissions {
			if !seen[perm.Name] {
				permNames = append(permNames, perm.Name)
				seen[perm.Name] = true
			}
		}
	}

	claims := jwt.MapClaims{
		"sub":         user.ID.String(),
		"name":        user.Name,
		"email":       user.Email,
		"roles":       roleNames,
		"permissions": permNames,
		"exp":         time.Now().Add(s.accessTokenTTL).Unix(),
		"iat":         time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func toUserInfo(u *model.User) UserInfo {
	roles := make([]string, 0, len(u.Roles))
	for _, r := range u.Roles {
		roles = append(roles, r.Name)
	}
	return UserInfo{
		ID:       u.ID,
		Name:     u.Name,
		Email:    u.Email,
		Avatar:   u.Avatar,
		IsActive: u.IsActive,
		Roles:    roles,
	}
}
