package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/simpeg/backend/pkg/apperror"
	"github.com/simpeg/backend/pkg/response"
)

// RequireRole memastikan user memiliki salah satu dari role yang diberikan.
// Harus dipasang setelah JWTAuth.
func RequireRole(roles ...string) gin.HandlerFunc {
	set := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		set[r] = struct{}{}
	}

	return func(c *gin.Context) {
		user := CurrentUser(c)
		if user == nil {
			response.Error(c, apperror.ErrUnauthorized)
			return
		}

		for _, role := range user.Roles {
			if _, ok := set[role.Name]; ok {
				c.Next()
				return
			}
		}

		response.Error(c, apperror.ErrForbidden)
	}
}

// RequirePermission memastikan user memiliki permission tertentu.
// Permission di-embed dalam claims JWT pada kolom "permissions".
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := CurrentUser(c)
		if user == nil {
			response.Error(c, apperror.ErrUnauthorized)
			return
		}

		for _, role := range user.Roles {
			for _, perm := range role.Permissions {
				if perm.Name == permission {
					c.Next()
					return
				}
			}
		}

		response.Error(c, apperror.ErrForbidden)
	}
}
