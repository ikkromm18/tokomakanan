package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS returns a configured CORS middleware
func CORS(origins []string) gin.HandlerFunc {
	cfg := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	if len(origins) == 0 {
		origins = []string{"http://localhost:3000", "http://localhost:5173"}
	}

	hasWildcard := false
	for _, o := range origins {
		if o == "*" {
			hasWildcard = true
			break
		}
	}

	if hasWildcard {
		cfg.AllowOriginFunc = func(origin string) bool { return true }
	} else {
		cfg.AllowOrigins = origins
	}

	return cors.New(cfg)
}
