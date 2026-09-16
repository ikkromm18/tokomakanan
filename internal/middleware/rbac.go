package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
)

// RequireRoles returns a Gin middleware restricting access to specific roles
func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("user_role")
		if userRole != "" {
			for _, role := range roles {
				if userRole == role {
					c.Next()
					return
				}
			}
		}

		response.Error(c, http.StatusForbidden, "Forbidden: insufficient permissions")
		c.Abort()
	}
}
