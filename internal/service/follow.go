package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

type followService struct {
	followRepo repository.FollowRepository
	userRepo   repository.UserRepository
	notifRepo  repository.NotificationRepository
	attrRepo   repository.AttributeRepository
	uaRepo     repository.UserAttributeRepository
	log        *logrus.Logger
}

// NewFollowService creates a new FollowService.
func NewFollowService(
	followRepo repository.FollowRepository,
	userRepo repository.UserRepository,
	notifRepo repository.NotificationRepository,
	log *logrus.Logger,
	attrRepo repository.AttributeRepository,
	uaRepo repository.UserAttributeRepository,
) FollowService {
	return &followService{
		followRepo: followRepo,
		userRepo:   userRepo,
		notifRepo:  notifRepo,
		attrRepo:   attrRepo,
		uaRepo:     uaRepo,
		log:        log,
	}
}

func (s *followService) Add(ctx context.Context, followerID, followedID uint64) error {
	if followerID == followedID {
		return errors.NewBadRequest("不能关注自己", nil)
	}
	if s.userRepo != nil {
		u, err := s.userRepo.GetByID(ctx, followedID)
		if err != nil {
			return err
		}
		if u == nil {
			return errors.NewNotFound("用户不存在", nil)
		}
	}
	existing, err := s.followRepo.Get(ctx, followerID, followedID)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.NewConflict("已关注", nil)
	}
	if err := s.followRepo.Create(ctx, &entity.Follow{
		FollowerID: followerID,
		FollowedID: followedID,
	}); err != nil {
		return err
	}
	if s.notifRepo != nil {
		target := followedID
		if err := s.notifRepo.Create(ctx, &entity.Notification{
			UserID:  &target,
			Type:    5,
			Title:   "收到新的关注",
			Content: "你有一个新的关注者",
			IsRead:  0,
		}); err != nil {
			s.log.WithError(err).Warn("发送关注通知失败")
		}
	}
	if err := s.incrUserFollowerCount(ctx, followedID); err != nil {
		s.log.WithError(err).Warn("更新用户被关注数失败")
	}
	return nil
}

func (s *followService) Remove(ctx context.Context, followerID, followedID uint64) error {
	if err := s.followRepo.Delete(ctx, followerID, followedID); err != nil {
		return err
	}
	if err := s.decrUserFollowerCount(ctx, followedID); err != nil {
		s.log.WithError(err).Warn("更新用户被关注数失败")
	}
	return nil
}

func (s *followService) incrUserFollowerCount(ctx context.Context, userID uint64) error {
	return s.adjustUserFollowerCount(ctx, userID, 1)
}

func (s *followService) decrUserFollowerCount(ctx context.Context, userID uint64) error {
	return s.adjustUserFollowerCount(ctx, userID, -1)
}

func (s *followService) adjustUserFollowerCount(ctx context.Context, userID uint64, delta int64) error {
	if s.attrRepo == nil || s.uaRepo == nil {
		return nil
	}
	attr, err := s.attrRepo.GetByName(ctx, "follower_count")
	if err != nil {
		return err
	}
	if attr == nil {
		attr = &entity.Attribute{Name: "follower_count", Type: entity.AttributeTypeBigInt}
		if err = s.attrRepo.Create(ctx, attr); err != nil {
			return err
		}
	}
	var current int64
	uas, err := s.uaRepo.ListByUserID(ctx, userID)
	if err != nil {
		return err
	}
	for _, ua := range uas {
		if ua == nil || ua.AttributeID != attr.ID || ua.ValueBigint == nil {
			continue
		}
		current = *ua.ValueBigint
		break
	}
	next := current + delta
	if next < 0 {
		next = 0
	}
	return s.uaRepo.Upsert(ctx, &entity.UserAttribute{
		UserID:      userID,
		AttributeID: attr.ID,
		ValueBigint: &next,
	})
}
