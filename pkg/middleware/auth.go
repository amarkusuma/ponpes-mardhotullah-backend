package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/simpeg/backend/internal/model"
	"github.com/simpeg/backend/pkg/apperror"
	"github.com/simpeg/backend/pkg/response"
)

const ctxUserKey = "currentUser"

// JWTAuth memvalidasi Bearer token dan menyimpan User ke dalam context.
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			response.Error(c, apperror.ErrUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, apperror.ErrUnauthorized
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			response.Error(c, apperror.ErrUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Error(c, apperror.ErrUnauthorized)
			return
		}

		userID, err := uuid.Parse(claims["sub"].(string))
		if err != nil {
			response.Error(c, apperror.ErrUnauthorized)
			return
		}

		// Rekonstruksi roles dari claim
		var roles []model.Role
		if rawRoles, ok := claims["roles"].([]interface{}); ok {
			for _, r := range rawRoles {
				roles = append(roles, model.Role{Name: r.(string)})
			}
		}

		// Rekonstruksi permissions dari claim
		var permissions []model.Permission
		if rawPerms, ok := claims["permissions"].([]interface{}); ok {
			for _, p := range rawPerms {
				permissions = append(permissions, model.Permission{Name: p.(string)})
			}
		}
		// Tempel permissions ke setiap role agar RBAC middleware bisa mengecek
		for i := range roles {
			roles[i].Permissions = permissions
		}

		currentUser := &model.User{
			Base:     model.Base{ID: userID},
			Name:     safeStr(claims["name"]),
			Email:    safeStr(claims["email"]),
			IsActive: true,
			Roles:    roles,
		}

		c.Set(ctxUserKey, currentUser)
		c.Next()
	}
}

// CurrentUser mengambil user yang sudah terautentikasi dari context.
// Mengembalikan nil jika tidak ada (endpoint publik).
func CurrentUser(c *gin.Context) *model.User {
	v, _ := c.Get(ctxUserKey)
	u, _ := v.(*model.User)
	return u
}

// MustCurrentUser sama dengan CurrentUser tapi panic jika tidak ada
// (digunakan di dalam handler yang sudah dijaga JWTAuth).
func MustCurrentUser(c *gin.Context) *model.User {
	u := CurrentUser(c)
	if u == nil {
		panic("MustCurrentUser called outside authenticated route")
	}
	return u
}

func safeStr(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
