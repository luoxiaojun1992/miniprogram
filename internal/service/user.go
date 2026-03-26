package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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
	attrRepo      repository.AttributeRepository
	uaRepo        repository.UserAttributeRepository
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
	deps ...interface{},
) UserService {
	var attrRepo repository.AttributeRepository
	var uaRepo repository.UserAttributeRepository
	for _, dep := range deps {
		switch v := dep.(type) {
		case repository.AttributeRepository:
			attrRepo = v
		case repository.UserAttributeRepository:
			uaRepo = v
		}
	}
	return &userService{
		userRepo:      userRepo,
		adminUserRepo: adminUserRepo,
		tagRepo:       tagRepo,
		roleRepo:      roleRepo,
		permRepo:      permRepo,
		attrRepo:      attrRepo,
		uaRepo:        uaRepo,
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
	s.bindAvatarFileIDFromUserAttributes(ctx, user)
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
	if req.AvatarFileID > 0 {
		if err = s.upsertAvatarFileIDAttribute(ctx, userID, req.AvatarFileID); err != nil {
			return err
		}
	}
	return s.userRepo.Update(ctx, user)
}

func (s *userService) bindAvatarFileIDFromUserAttributes(ctx context.Context, user *entity.User) {
	if user == nil || s.uaRepo == nil {
		return
	}
	uas, err := s.uaRepo.ListByUserID(ctx, user.ID)
	if err != nil {
		return
	}
	for _, ua := range uas {
		if ua == nil || ua.Attribute == nil || ua.Attribute.Name != "avatar_file_id" || ua.ValueBigint == nil || *ua.ValueBigint <= 0 {
			continue
		}
		v := uint64(*ua.ValueBigint)
		user.AvatarFileID = &v
		return
	}
}

func (s *userService) upsertAvatarFileIDAttribute(ctx context.Context, userID uint64, fileID uint64) error {
	if s.attrRepo == nil || s.uaRepo == nil {
		return nil
	}
	attr, err := s.attrRepo.GetByName(ctx, "avatar_file_id")
	if err != nil {
		return err
	}
	if attr == nil {
		attr = &entity.Attribute{Name: "avatar_file_id", Type: entity.AttributeTypeBigInt}
		if err = s.attrRepo.Create(ctx, attr); err != nil {
			return err
		}
	}
	v := int64(fileID)
	return s.uaRepo.Upsert(ctx, &entity.UserAttribute{
		UserID:      userID,
		AttributeID: attr.ID,
		ValueBigint: &v,
		ValueString: "",
	})
}

func (s *userService) GetPermissions(ctx context.Context, userID uint64) ([]string, []string, error) {
	roles, err := s.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	allRoles, err := s.roleRepo.List(ctx)
	if err != nil {
		return nil, nil, err
	}
	roleMap := map[uint]*entity.Role{}
	for _, role := range allRoles {
		roleMap[role.ID] = role
	}
	expandedRoleIDs := expandRoleIDs(roles, allRoles)
	expandedRoles := make([]*entity.Role, 0, len(expandedRoleIDs))
	directRoleMap := map[uint]*entity.Role{}
	for _, role := range roles {
		directRoleMap[role.ID] = role
	}
	for roleID := range expandedRoleIDs {
		if role, ok := roleMap[roleID]; ok {
			expandedRoles = append(expandedRoles, role)
			continue
		}
		if role, ok := directRoleMap[roleID]; ok {
			expandedRoles = append(expandedRoles, role)
		}
	}
	perms, err := s.permRepo.GetPermissionsByRoleIDs(ctx, toRoleIDSlice(expandedRoleIDs))
	if err != nil {
		return nil, nil, err
	}
	if len(perms) == 0 {
		perms, err = s.permRepo.GetUserPermissions(ctx, userID)
		if err != nil {
			return nil, nil, err
		}
	}
	allPerms, err := s.permRepo.List(ctx)
	if err != nil {
		return nil, nil, err
	}
	expandedPermIDs := expandPermissionIDs(perms, allPerms)
	permMap := map[uint]*entity.Permission{}
	for _, perm := range allPerms {
		permMap[perm.ID] = perm
	}
	expandedPerms := make([]*entity.Permission, 0, len(expandedPermIDs))
	directPermMap := map[uint]*entity.Permission{}
	for _, perm := range perms {
		directPermMap[perm.ID] = perm
	}
	for permID := range expandedPermIDs {
		if perm, ok := permMap[permID]; ok {
			expandedPerms = append(expandedPerms, perm)
			continue
		}
		if perm, ok := directPermMap[permID]; ok {
			expandedPerms = append(expandedPerms, perm)
		}
	}

	roleNames := make([]string, 0, len(expandedRoles))
	for _, r := range expandedRoles {
		roleNames = append(roleNames, r.Name)
	}
	permCodes := make([]string, 0, len(expandedPerms))
	for _, p := range expandedPerms {
		permCodes = append(permCodes, p.Code)
	}
	return roleNames, permCodes, nil
}

func (s *userService) List(ctx context.Context, page, pageSize int, keyword string, userType *int8) ([]*entity.User, int64, error) {
	return s.userRepo.List(ctx, page, pageSize, keyword, userType)
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
		OpenID:   fmt.Sprintf("admin_%s", uuid.NewString()),
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
	hasAssociations, err := s.userRepo.HasAssociations(ctx, id)
	if err != nil {
		return err
	}
	if hasAssociations {
		return errors.NewBadRequest("用户存在关联数据，禁止删除", nil)
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

func (s *userService) DeleteTag(ctx context.Context, userID uint64, tagID uint) error {
	return s.tagRepo.Delete(ctx, tagID)
}

func expandRoleIDs(userRoles []*entity.Role, allRoles []*entity.Role) map[uint]struct{} {
	parentByRole := make(map[uint]uint, len(allRoles))
	childrenByRole := make(map[uint][]uint, len(allRoles))
	for _, role := range allRoles {
		parentByRole[role.ID] = role.ParentID
		if role.ParentID > 0 {
			childrenByRole[role.ParentID] = append(childrenByRole[role.ParentID], role.ID)
		}
	}
	visited := map[uint]struct{}{}
	stack := make([]uint, 0, len(userRoles))
	for _, role := range userRoles {
		stack = append(stack, role.ID)
	}
	for len(stack) > 0 {
		n := len(stack) - 1
		roleID := stack[n]
		stack = stack[:n]
		if _, ok := visited[roleID]; ok {
			continue
		}
		visited[roleID] = struct{}{}
		if parentID, ok := parentByRole[roleID]; ok && parentID > 0 {
			stack = append(stack, parentID)
		}
		stack = append(stack, childrenByRole[roleID]...)
	}
	return visited
}

func toRoleIDSlice(roleIDs map[uint]struct{}) []uint {
	ids := make([]uint, 0, len(roleIDs))
	for roleID := range roleIDs {
		ids = append(ids, roleID)
	}
	return ids
}

func expandPermissionIDs(basePerms []*entity.Permission, allPerms []*entity.Permission) map[uint]struct{} {
	parentByPerm := make(map[uint]uint, len(allPerms))
	childrenByPerm := make(map[uint][]uint, len(allPerms))
	for _, perm := range allPerms {
		parentByPerm[perm.ID] = perm.ParentID
		if perm.ParentID > 0 {
			childrenByPerm[perm.ParentID] = append(childrenByPerm[perm.ParentID], perm.ID)
		}
	}
	visited := map[uint]struct{}{}
	stack := make([]uint, 0, len(basePerms))
	for _, perm := range basePerms {
		stack = append(stack, perm.ID)
	}
	for len(stack) > 0 {
		n := len(stack) - 1
		permID := stack[n]
		stack = stack[:n]
		if _, ok := visited[permID]; ok {
			continue
		}
		visited[permID] = struct{}{}
		if parentID, ok := parentByPerm[permID]; ok && parentID > 0 {
			stack = append(stack, parentID)
		}
		stack = append(stack, childrenByPerm[permID]...)
	}
	return visited
}
