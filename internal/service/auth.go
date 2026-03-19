package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"github.com/luoxiaojun1992/miniprogram/internal/middleware"
	"github.com/luoxiaojun1992/miniprogram/internal/model/dto"
	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/errors"
	"github.com/luoxiaojun1992/miniprogram/internal/pkg/wechat"
	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

type authService struct {
	userRepo      repository.UserRepository
	adminUserRepo repository.AdminUserRepository
	wechatClient  wechat.Client
	jwtSecret     string
	jwtExpiry     int
	log           *logrus.Logger
	signingMethod jwt.SigningMethod // nil means use jwt.SigningMethodHS256 (default)
}

// NewAuthService creates a new AuthService.
func NewAuthService(
	userRepo repository.UserRepository,
	adminUserRepo repository.AdminUserRepository,
	wechatClient wechat.Client,
	jwtSecret string,
	jwtExpiry int,
	log *logrus.Logger,
) AuthService {
	return &authService{
		userRepo:      userRepo,
		adminUserRepo: adminUserRepo,
		wechatClient:  wechatClient,
		jwtSecret:     jwtSecret,
		jwtExpiry:     jwtExpiry,
		log:           log,
	}
}

func (s *authService) WechatLogin(ctx context.Context, req *dto.WechatLoginRequest) (*dto.LoginResponseData, error) {
	openID, err := s.wechatClient.Code2Session(ctx, req.Code)
	if err != nil {
		return nil, errors.NewUnauthorized("微信认证失败", err)
	}

	user, err := s.userRepo.GetByOpenID(ctx, openID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		user = &entity.User{
			OpenID:   openID,
			Nickname: "微信用户",
			UserType: 1,
		}
		if err = s.userRepo.Create(ctx, user); err != nil {
			return nil, err
		}
	}

	tokenStr, err := s.generateToken(user.ID, user.UserType)
	if err != nil {
		return nil, errors.NewInternal("生成Token失败", err)
	}

	return &dto.LoginResponseData{
		AccessToken: tokenStr,
		TokenType:   "Bearer",
		ExpiresIn:   s.jwtExpiry,
		UserInfo:    user,
	}, nil
}

func (s *authService) AdminLogin(ctx context.Context, req *dto.AdminLoginRequest) (*dto.LoginResponseData, error) {
	admin, err := s.adminUserRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, errors.NewUnauthorized("账号或密码错误", nil)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.NewUnauthorized("账号或密码错误", err)
	}

	user, err := s.userRepo.GetByID(ctx, admin.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.NewUnauthorized("用户不存在", nil)
	}

	_ = s.adminUserRepo.UpdateLastLogin(ctx, admin.ID)

	tokenStr, err := s.generateToken(user.ID, user.UserType)
	if err != nil {
		return nil, errors.NewInternal("生成Token失败", err)
	}

	return &dto.LoginResponseData{
		AccessToken: tokenStr,
		TokenType:   "Bearer",
		ExpiresIn:   s.jwtExpiry,
		UserInfo:    user,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, userID uint64, userType int8) (*dto.LoginResponseData, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.NewUnauthorized("用户不存在", nil)
	}

	tokenStr, err := s.generateToken(user.ID, user.UserType)
	if err != nil {
		return nil, errors.NewInternal("生成Token失败", err)
	}

	return &dto.LoginResponseData{
		AccessToken: tokenStr,
		TokenType:   "Bearer",
		ExpiresIn:   s.jwtExpiry,
		UserInfo:    user,
	}, nil
}

func (s *authService) generateToken(userID uint64, userType int8) (string, error) {
	claims := &middleware.JWTClaims{
		UserID:   userID,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.jwtExpiry) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", userID),
		},
	}
	method := s.signingMethod
	if method == nil {
		method = jwt.SigningMethodHS256
	}
	token := jwt.NewWithClaims(method, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
