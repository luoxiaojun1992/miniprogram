package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

type bannerService struct {
	bannerRepo repository.BannerRepository
	fileRepo   repository.FileRepository
	log        *logrus.Logger
}

// NewBannerService creates a new BannerService.
func NewBannerService(bannerRepo repository.BannerRepository, log *logrus.Logger, fileRepo ...repository.FileRepository) BannerService {
	var fRepo repository.FileRepository
	if len(fileRepo) > 0 {
		fRepo = fileRepo[0]
	}
	return &bannerService{
		bannerRepo: bannerRepo,
		fileRepo:   fRepo,
		log:        log,
	}
}

func (s *bannerService) List(ctx context.Context) ([]*entity.Banner, error) {
	status := int8(1)
	return s.bannerRepo.List(ctx, &status)
}

func (s *bannerService) AdminList(ctx context.Context, status *int8) ([]*entity.Banner, error) {
	return s.bannerRepo.List(ctx, status)
}

func (s *bannerService) Create(ctx context.Context, req *dto.CreateBannerRequest) (uint64, error) {
	if err := s.validateBannerMediaFile(ctx, req.ImageFileID); err != nil {
		return 0, err
	}
	banner := &entity.Banner{
		Title:       req.Title,
		ImageFileID: toOptionalUint64(req.ImageFileID),
		LinkURL:     req.LinkURL,
		SortOrder:   req.SortOrder,
		Status:      req.Status,
	}
	if banner.Status != 0 && banner.Status != 1 {
		banner.Status = 1
	}
	if err := s.bannerRepo.Create(ctx, banner); err != nil {
		return 0, err
	}
	return banner.ID, nil
}

func (s *bannerService) Update(ctx context.Context, id uint64, req *dto.CreateBannerRequest) error {
	banner, err := s.bannerRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if banner == nil {
		return errors.NewNotFound("轮播图不存在", nil)
	}
	if err := s.validateBannerMediaFile(ctx, req.ImageFileID); err != nil {
		return err
	}
	banner.Title = req.Title
	banner.ImageFileID = toOptionalUint64(req.ImageFileID)
	banner.LinkURL = req.LinkURL
	banner.SortOrder = req.SortOrder
	banner.Status = req.Status
	if banner.Status != 0 && banner.Status != 1 {
		banner.Status = 1
	}
	return s.bannerRepo.Update(ctx, banner)
}

func (s *bannerService) Delete(ctx context.Context, id uint64) error {
	banner, err := s.bannerRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if banner == nil {
		return errors.NewNotFound("轮播图不存在", nil)
	}
	return s.bannerRepo.Delete(ctx, id)
}

func (s *bannerService) validateBannerMediaFile(ctx context.Context, fileID uint64) error {
	if fileID == 0 {
		return errors.NewBadRequest("轮播图素材不能为空", nil)
	}
	if s.fileRepo == nil {
		return nil
	}
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return err
	}
	if file == nil {
		return errors.NewBadRequest("轮播图素材文件不存在", nil)
	}
	if file.Usage != "protected" {
		return errors.NewBadRequest("轮播图素材必须为受保护文件", nil)
	}
	if file.Category != "image" && file.Category != "video" {
		return errors.NewBadRequest("轮播图素材仅支持图片或视频", nil)
	}
	return nil
}
