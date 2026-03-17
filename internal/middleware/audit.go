package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
)

var auditActions = map[string]string{
	"POST":   "create",
	"PUT":    "update",
	"PATCH":  "update",
	"DELETE": "delete",
}

type auditLogger interface {
	Log(ctx context.Context, log *entity.AuditLog)
}

// AuditLogMiddleware logs successful admin write operations as audit logs.
func AuditLogMiddleware(auditSvc auditLogger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if auditSvc == nil {
			ctx.Next()
			return
		}
		method := strings.ToUpper(strings.TrimSpace(ctx.Request.Method))
		action, shouldAudit := auditActions[method]
		if !shouldAudit || !strings.HasPrefix(ctx.FullPath(), "/v1/admin/") {
			ctx.Next()
			return
		}

		bodyBytes, readErr := io.ReadAll(io.LimitReader(ctx.Request.Body, 1<<20))
		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		ctx.Next()

		status := ctx.Writer.Status()
		if status < 200 || status >= 300 {
			return
		}

		userID, _ := GetCurrentUserID(ctx)
		module := extractAuditModule(ctx.FullPath())
		payloadMap := gin.H{
			"path":   ctx.FullPath(),
			"query":  ctx.Request.URL.RawQuery,
			"method": method,
			"body":   string(bodyBytes),
		}
		if readErr != nil {
			payloadMap["body_read_error"] = fmt.Sprintf("%v", readErr)
		}
		payload, _ := json.Marshal(payloadMap)
		auditSvc.Log(ctx.Request.Context(), &entity.AuditLog{
			UserID:      userID,
			Username:    ctx.GetString("username"),
			Action:      action,
			Module:      module,
			Description: method + " " + ctx.FullPath(),
			IPAddress:   ctx.ClientIP(),
			UserAgent:   ctx.Request.UserAgent(),
			RequestData: string(payload),
		})
	}
}

func extractAuditModule(fullPath string) string {
	trimmed := strings.Trim(fullPath, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) < 3 {
		return "admin"
	}
	return parts[2]
}
