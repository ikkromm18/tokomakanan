package middleware

import (
	"bytes"
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// Logger returns a Gin middleware logging HTTP requests to os.Stdout using Zerolog
func Logger() gin.HandlerFunc {
	return LoggerWithWriter(os.Stdout)
}

// LoggerWithWriter returns a Gin middleware logging HTTP requests to the provided io.Writer
func LoggerWithWriter(w io.Writer) gin.HandlerFunc {
	logger := zerolog.New(w).With().Timestamp().Logger()

	return func(c *gin.Context) {
		start := time.Now()

		reqID := c.GetHeader("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		}
		c.Set("request_id", reqID)
		c.Header("X-Request-ID", reqID)

		bodyWriter := &bodyLogWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = bodyWriter

		c.Next()

		latency := time.Since(start).Milliseconds()
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()

		var event *zerolog.Event
		if status >= 500 {
			event = logger.Error()
		} else if status >= 400 {
			event = logger.Warn()
		} else {
			event = logger.Info()
		}

		event = event.
			Str("method", method).
			Str("path", path).
			Int("status", status).
			Int64("latency_ms", latency).
			Str("client_ip", clientIP).
			Str("request_id", reqID)

		if userID, exists := c.Get("user_id"); exists {
			event = event.Any("user_id", userID)
		}

		// Log response body only on error (status >= 400), never log request body
		if status >= 400 && bodyWriter.body.Len() > 0 {
			event = event.Str("error_response", bodyWriter.body.String())
		}

		event.Msg("HTTP request")
	}
}
