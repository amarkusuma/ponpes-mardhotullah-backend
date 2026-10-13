package service

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/simpeg/backend/internal/model"
	"github.com/simpeg/backend/internal/repository"
)

type AuthService struct {
	db     *gorm.DB
	cache  *redis.Client
	logger *zap.Logger
	repo   *repository.UserRepository
	secret string
}

type loginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type registerRequest struct {
	Name     string `json:"name"     binding:"required"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	// Role ditetapkan lewat endpoint terpisah oleh admin.
	// Default saat register: tidak ada role (atau role "santri" jika dikonfigurasi).
}

type authResponse struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

type userProfileResponse struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Roles []string  `json:"roles"`
}

func NewAuthService(db *gorm.DB, cache *redis.Client, logger *zap.Logger, secret string) *AuthService {
	return &AuthService{
		db:     db,
		cache:  cache,
		logger: logger,
		repo:   repository.NewUserRepository(db),
		secret: secret,
	}
}

func (s *AuthService) LoginHandler(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := s.repo.FindByEmail(strings.ToLower(req.Email))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		s.logger.Error("login error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not process request"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := s.generateToken(user, s.secret)
	if err != nil {
		s.logger.Error("token generation failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create token"})
		return
	}

	c.JSON(http.StatusOK, authResponse{Token: token, User: user})
}

func (s *AuthService) RegisterHandler(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("password hash failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save user"})
		return
	}

	user := &model.User{
		Name:     req.Name,
		Email:    strings.ToLower(req.Email),
		Password: string(hash),
		IsActive: true,
	}

	if err := s.repo.Create(user); err != nil {
		s.logger.Error("register error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
		return
	}

	token, err := s.generateToken(user, s.secret)
	if err != nil {
		s.logger.Error("token generation failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create token"})
		return
	}

	c.JSON(http.StatusCreated, authResponse{Token: token, User: user})
}

func (s *AuthService) ProfileHandler(c *gin.Context) {
	raw, found := c.Get("currentUser")
	if !found {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	current := raw.(*model.User)

	roleNames := make([]string, 0, len(current.Roles))
	for _, r := range current.Roles {
		roleNames = append(roleNames, r.Name)
	}

	c.JSON(http.StatusOK, userProfileResponse{
		ID:    current.ID,
		Name:  current.Name,
		Email: current.Email,
		Roles: roleNames,
	})
}

func (s *AuthService) generateToken(user *model.User, secret string) (string, error) {
	roleNames := make([]string, 0, len(user.Roles))
	for _, r := range user.Roles {
		roleNames = append(roleNames, r.Name)
	}

	claims := jwt.MapClaims{
		"sub":   user.ID.String(),
		"name":  user.Name,
		"email": user.Email,
		"roles": roleNames,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
