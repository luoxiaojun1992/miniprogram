package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// LoggerMiddleware logs each request with method, path, status and latency.
func LoggerMiddleware(log *logrus.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		path := ctx.Request.URL.Path

		ctx.Next()

		latency := time.Since(start)
		log.WithFields(logrus.Fields{
			"method":     ctx.Request.Method,
			"path":       path,
			"status":     ctx.Writer.Status(),
			"latency":    latency,
			"ip":         ctx.ClientIP(),
			"request_id": ctx.GetString("request_id"),
		}).Info("request")
	}
}
