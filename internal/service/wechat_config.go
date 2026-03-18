package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

type wechatConfigService struct {
	wechatConfigRepo repository.WechatConfigRepository
	log              *logrus.Logger
}

// NewWechatConfigService creates a new WechatConfigService.
func NewWechatConfigService(wechatConfigRepo repository.WechatConfigRepository, log *logrus.Logger) WechatConfigService {
	return &wechatConfigService{wechatConfigRepo: wechatConfigRepo, log: log}
}

func (s *wechatConfigService) Get(ctx context.Context) (*entity.WechatConfig, error) {
	cfg, err := s.wechatConfigRepo.Get(ctx)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return &entity.WechatConfig{}, nil
	}
	return cfg, nil
}

func (s *wechatConfigService) Update(ctx context.Context, req *dto.UpdateWechatConfigRequest) error {
	cfg, err := s.wechatConfigRepo.Get(ctx)
	if err != nil {
		return err
	}
	if cfg == nil {
		cfg = &entity.WechatConfig{}
	}
	cfg.AppID = req.AppID
	cfg.AppSecret = req.AppSecret
	if req.APIToken != "" {
		cfg.APIToken = req.APIToken
	}
	return s.wechatConfigRepo.Update(ctx, cfg)
}
