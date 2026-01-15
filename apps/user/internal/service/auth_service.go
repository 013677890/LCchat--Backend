package service

import (
	"ChatServer/apps/user/internal/dto"
	"ChatServer/apps/user/internal/repository"
	"ChatServer/apps/user/internal/utils"
	"ChatServer/pkg/logger"
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// authServiceImpl 认证服务实现
type authServiceImpl struct {
	userRepo   repository.UserRepository
	deviceRepo repository.DeviceSessionRepository
}

// NewAuthService 创建认证服务实例
func NewAuthService(
	userRepo repository.UserRepository,
	deviceRepo repository.DeviceSessionRepository,
) AuthService {
	return &authServiceImpl{
		userRepo:   userRepo,
		deviceRepo: deviceRepo,
	}
}

// Login 用户登录
// 业务流程：
//   1. 根据手机号查询用户
//   2. 校验用户状态（是否被禁用）
//   3. 校验密码
//   4. 返回用户信息（供Gateway生成Token）
// 
// 错误码映射：
//   - codes.NotFound: 用户不存在
//   - codes.Unauthenticated: 密码错误
//   - codes.PermissionDenied: 用户被禁用
//   - codes.Internal: 系统内部错误
func (s *authServiceImpl) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// 记录登录请求（手机号脱敏）
	logger.Info(ctx, "用户登录请求",
		logger.String("telephone", utils.MaskPhone(req.Telephone)),
		logger.String("device_id", req.DeviceInfo.GetDeviceId()),
		logger.String("platform", req.DeviceInfo.GetPlatform()),
	)

	// 1. 根据手机号查询用户
	user, err := s.userRepo.GetByPhone(ctx, req.Telephone)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn(ctx, "用户不存在",
				logger.String("telephone", utils.MaskPhone(req.Telephone)),
			)
			return nil, status.Error(codes.NotFound, "用户不存在")
		}
		logger.Error(ctx, "查询用户失败",
			logger.String("telephone", utils.MaskPhone(req.Telephone)),
			logger.ErrorField("error", err),
		)
		return nil, status.Error(codes.Internal, "数据库查询失败")
	}

	// 2. 校验用户状态
	if user.Status == 1 {
		logger.Warn(ctx, "用户已被禁用",
			logger.String("user_uuid", user.Uuid),
			logger.String("telephone", utils.MaskPhone(req.Telephone)),
		)
		return nil, status.Error(codes.PermissionDenied, "用户已被禁用")
	}

	// 3. 校验密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		logger.Warn(ctx, "密码错误",
			logger.String("user_uuid", user.Uuid),
			logger.String("telephone", utils.MaskPhone(req.Telephone)),
		)
		return nil, status.Error(codes.Unauthenticated, "密码错误")
	}

	// 4. 登录成功
	logger.Info(ctx, "用户登录成功",
		logger.String("user_uuid", user.Uuid),
		logger.String("telephone", utils.MaskPhone(req.Telephone)),
		logger.String("device_id", req.DeviceInfo.GetDeviceId()),
	)

	// 返回用户信息
	return &dto.LoginResponse{
		UserInfo: dto.ConvertModelToUserInfo(user),
	}, nil
}

// Register 用户注册
// 注意：此方法暂未实现，预留接口
func (s *authServiceImpl) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	return nil, status.Error(codes.Unimplemented, "注册功能暂未实现")
}

// RefreshToken 刷新Token
// 注意：此方法暂未实现，预留接口
func (s *authServiceImpl) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error) {
	return nil, status.Error(codes.Unimplemented, "刷新Token功能暂未实现")
}

// Logout 用户登出
// 注意：此方法暂未实现，预留接口
func (s *authServiceImpl) Logout(ctx context.Context, req *dto.LogoutRequest) error {
	return status.Error(codes.Unimplemented, "登出功能暂未实现")
}

// SendSmsCode 发送短信验证码
// 注意：此方法暂未实现，预留接口
func (s *authServiceImpl) SendSmsCode(ctx context.Context, telephone string, codeType int32) error {
	return status.Error(codes.Unimplemented, "发送验证码功能暂未实现")
}

// ValidateSmsCode 验证短信验证码
// 注意：此方法暂未实现，预留接口
func (s *authServiceImpl) ValidateSmsCode(ctx context.Context, telephone, code string, codeType int32) (bool, error) {
	return false, status.Error(codes.Unimplemented, "验证码校验功能暂未实现")
}
