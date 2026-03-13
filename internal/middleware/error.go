package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
)

// ErrorMiddleware is a Gin middleware that handles errors added via ctx.Error().
func ErrorMiddleware(log *logrus.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		if len(ctx.Errors) > 0 {
			err := ctx.Errors.Last().Err

			if appErr, ok := err.(*errors.AppError); ok && appErr.Cause != nil {
				log.WithError(appErr.Cause).WithField("code", appErr.Code).Error(appErr.Message)
			} else {
				log.WithError(err).Error("request error")
			}

			httpCode, resp := errors.ToResponse(err)
			ctx.JSON(httpCode, resp)
			ctx.Abort()
		}
	}
}
