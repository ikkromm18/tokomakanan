package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
)

func setSecurityHeaders(c *gin.Context) {
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Frame-Options", "DENY")
	c.Header("X-XSS-Protection", "1; mode=block")
	c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
}

// SecurityHeaders returns a Gin middleware that attaches security headers to responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		setSecurityHeaders(c)
		c.Next()
	}
}

// Recovery returns a Gin middleware that catches unhandled panics, logs stack trace,
// returns a safe 500 error response, and applies security headers.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		setSecurityHeaders(c)

		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				log.Error().
					Interface("error", r).
					Str("stack", string(stack)).
					Msg("panic recovered")

				response.Error(c, http.StatusInternalServerError, "Internal server error")
				c.Abort()
			}
		}()

		c.Next()
	}
}
