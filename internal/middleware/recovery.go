package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RecoveryMiddleware recovers from panics and returns a 500 error.
func RecoveryMiddleware(log *logrus.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.WithField("panic", r).Error("recovered from panic")
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"code":    500000,
					"message": "服务器内部错误",
					"data":    nil,
				})
				ctx.Abort()
			}
		}()
		ctx.Next()
	}
}
