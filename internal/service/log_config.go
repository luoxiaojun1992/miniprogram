package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

// ==================== Study Record Service ====================

type logConfigService struct {
	logConfigRepo repository.LogConfigRepository
	log           *logrus.Logger
}

// NewLogConfigService creates a new LogConfigService.
func NewLogConfigService(logConfigRepo repository.LogConfigRepository, log *logrus.Logger) LogConfigService {
	return &logConfigService{logConfigRepo: logConfigRepo, log: log}
}

func (s *logConfigService) Get(ctx context.Context) (*entity.LogConfig, error) {
	cfg, err := s.logConfigRepo.Get(ctx)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return &entity.LogConfig{RetentionDays: 90}, nil
	}
	return cfg, nil
}

func (s *logConfigService) Update(ctx context.Context, req *dto.UpdateLogConfigRequest) error {
	cfg, err := s.logConfigRepo.Get(ctx)
	if err != nil {
		return err
	}
	if cfg == nil {
		cfg = &entity.LogConfig{}
	}
	cfg.RetentionDays = req.RetentionDays
	return s.logConfigRepo.Update(ctx, cfg)
}
