package service

import (
	"context"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

type userService struct {
	userRepo      repository.UserRepository
	adminUserRepo repository.AdminUserRepository
	tagRepo       repository.UserTagRepository
	roleRepo      repository.RoleRepository
	permRepo      repository.PermissionRepository
	log           *logrus.Logger
}

// NewUserService creates a new UserService.
func NewUserService(
	userRepo repository.UserRepository,
	adminUserRepo repository.AdminUserRepository,
	tagRepo repository.UserTagRepository,
	roleRepo repository.RoleRepository,
	permRepo repository.PermissionRepository,
	log *logrus.Logger,
) UserService {
	return &userService{
		userRepo:      userRepo,
		adminUserRepo: adminUserRepo,
		tagRepo:       tagRepo,
		roleRepo:      roleRepo,
		permRepo:      permRepo,
		log:           log,
	}
}

func (s *userService) GetProfile(ctx context.Context, userID uint64) (*entity.User, error) {
	user, err := s.userRepo.GetWithTags(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.NewNotFound("用户不存在", nil)
	}
	return user, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID uint64, req *dto.UserProfileUpdateRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.NewNotFound("用户不存在", nil)
	}
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}
	return s.userRepo.Update(ctx, user)
}

func (s *userService) GetPermissions(ctx context.Context, userID uint64) ([]string, []string, error) {
	roles, err := s.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	perms, err := s.permRepo.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	roleNames := make([]string, 0, len(roles))
	for _, r := range roles {
		roleNames = append(roleNames, r.Name)
	}
	permCodes := make([]string, 0, len(perms))
	for _, p := range perms {
		permCodes = append(permCodes, p.Code)
	}
	return roleNames, permCodes, nil
}

func (s *userService) List(ctx context.Context, page, pageSize int, keyword string, userType, status *int8) ([]*entity.User, int64, error) {
	return s.userRepo.List(ctx, page, pageSize, keyword, userType, status)
}

func (s *userService) GetByID(ctx context.Context, id uint64) (*entity.User, error) {
	user, err := s.userRepo.GetWithTags(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.NewNotFound("用户不存在", nil)
	}
	return user, nil
}

func (s *userService) CreateAdminUser(ctx context.Context, req *dto.CreateAdminUserRequest) (uint64, error) {
	existing, err := s.adminUserRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return 0, err
	}
	if existing != nil {
		return 0, errors.NewConflict("邮箱已存在", nil)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, errors.NewInternal("密码加密失败", err)
	}

	user := &entity.User{
		Nickname: req.Nickname,
		UserType: req.UserType,
		Status:   1,
	}
	if err = s.userRepo.Create(ctx, user); err != nil {
		return 0, err
	}

	admin := &entity.AdminUser{
		UserID:       user.ID,
		Email:        req.Email,
		PasswordHash: string(hash),
	}
	if err = s.adminUserRepo.Create(ctx, admin); err != nil {
		return 0, err
	}

	return user.ID, nil
}

func (s *userService) UpdateUser(ctx context.Context, id uint64, req *dto.UpdateUserRequest, operatorID uint64) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.NewNotFound("用户不存在", nil)
	}
	if id == operatorID && req.UserType != 0 && req.UserType != user.UserType {
		return errors.NewForbidden("不能修改自己的用户类型", nil)
	}
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.UserType != 0 {
		user.UserType = req.UserType
	}
	if req.Status != 0 {
		user.Status = req.Status
	}
	if req.FreezeEndTime != nil {
		user.FreezeEndTime = req.FreezeEndTime
	}
	return s.userRepo.Update(ctx, user)
}

func (s *userService) DeleteUser(ctx context.Context, id uint64) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.NewNotFound("用户不存在", nil)
	}
	return s.userRepo.Delete(ctx, id)
}

func (s *userService) AssignRoles(ctx context.Context, userID uint64, req *dto.AssignRolesRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.NewNotFound("用户不存在", nil)
	}
	return s.roleRepo.AssignUserRoles(ctx, userID, req.RoleIDs)
}

func (s *userService) AddTag(ctx context.Context, userID uint64, req *dto.AddTagRequest) (uint, error) {
	tags, err := s.tagRepo.GetByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}
	if len(tags) >= 10 {
		return 0, errors.NewBadRequest("标签数量已达上限", nil)
	}
	tag := &entity.UserTag{
		UserID:  userID,
		TagName: req.TagName,
	}
	if err = s.tagRepo.Create(ctx, tag); err != nil {
		return 0, err
	}
	return tag.ID, nil
}

func (s *userService) DeleteTag(ctx context.Context, userID, tagID uint64) error {
	return s.tagRepo.Delete(ctx, uint(tagID))
}
