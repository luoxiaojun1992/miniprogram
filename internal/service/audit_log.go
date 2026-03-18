package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

type auditLogService struct {
	auditLogRepo repository.AuditLogRepository
	log          *logrus.Logger
}

// NewAuditLogService creates a new AuditLogService.
func NewAuditLogService(auditLogRepo repository.AuditLogRepository, log *logrus.Logger) AuditLogService {
	return &auditLogService{auditLogRepo: auditLogRepo, log: log}
}

func (s *auditLogService) List(ctx context.Context, page, pageSize int, module, action string, startTime, endTime *string) ([]*entity.AuditLog, int64, error) {
	return s.auditLogRepo.List(ctx, page, pageSize, module, action, startTime, endTime)
}

func (s *auditLogService) Log(ctx context.Context, log *entity.AuditLog) {
	if err := s.auditLogRepo.Create(ctx, log); err != nil {
		s.log.WithError(err).Warn("记录审计日志失败")
	}
}
