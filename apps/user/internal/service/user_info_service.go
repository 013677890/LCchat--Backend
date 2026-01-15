package service

import (
	"ChatServer/apps/user/internal/dto"
	"ChatServer/apps/user/internal/repository"
	"ChatServer/apps/user/internal/utils"
	"ChatServer/pkg/logger"
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// userInfoServiceImpl 用户信息服务实现
type userInfoServiceImpl struct {
	userRepo repository.UserRepository
}

// NewUserInfoService 创建用户信息服务实例
func NewUserInfoService(userRepo repository.UserRepository) UserQueryService {
	return &userInfoServiceImpl{
		userRepo: userRepo,
	}
}

// GetUserInfo 获取用户信息
// 支持通过UUID或手机号查询，UUID优先
func (s *userInfoServiceImpl) GetUserInfo(ctx context.Context, req *dto.GetUserInfoRequest) (*dto.GetUserInfoResponse, error) {
	var user *dto.UserInfo
	var err error

	if req.UserUUID != "" {
		// 优先使用UUID查询
		logger.Debug(ctx, "根据UUID查询用户信息",
			logger.String("user_uuid", req.UserUUID),
		)
		
		userModel, queryErr := s.userRepo.GetByUUID(ctx, req.UserUUID)
		if queryErr != nil {
			if errors.Is(queryErr, gorm.ErrRecordNotFound) {
				logger.Warn(ctx, "用户不存在",
					logger.String("user_uuid", req.UserUUID),
				)
				return nil, status.Error(codes.NotFound, "用户不存在")
			}
			logger.Error(ctx, "查询用户失败",
				logger.String("user_uuid", req.UserUUID),
				logger.ErrorField("error", queryErr),
			)
			return nil, status.Error(codes.Internal, "数据库查询失败")
		}
		user = dto.ConvertModelToUserInfo(userModel)
	} else if req.Telephone != "" {
		// 使用手机号查询
		logger.Debug(ctx, "根据手机号查询用户信息",
			logger.String("telephone", utils.MaskPhone(req.Telephone)),
		)
		
		userModel, queryErr := s.userRepo.GetByPhone(ctx, req.Telephone)
		if queryErr != nil {
			if errors.Is(queryErr, gorm.ErrRecordNotFound) {
				logger.Warn(ctx, "用户不存在",
					logger.String("telephone", utils.MaskPhone(req.Telephone)),
				)
				return nil, status.Error(codes.NotFound, "用户不存在")
			}
			logger.Error(ctx, "查询用户失败",
				logger.String("telephone", utils.MaskPhone(req.Telephone)),
				logger.ErrorField("error", queryErr),
			)
			return nil, status.Error(codes.Internal, "数据库查询失败")
		}
		user = dto.ConvertModelToUserInfo(userModel)
	} else {
		return nil, status.Error(codes.InvalidArgument, "UUID和手机号不能同时为空")
	}

	return &dto.GetUserInfoResponse{
		UserInfo: user,
	}, err
}

// UpdateUserInfo 更新用户信息
// 注意：此方法暂未实现，预留接口
func (s *userInfoServiceImpl) UpdateUserInfo(ctx context.Context, req *dto.UpdateUserInfoRequest) (*dto.UpdateUserInfoResponse, error) {
	return nil, status.Error(codes.Unimplemented, "更新用户信息功能暂未实现")
}

// UpdateAvatar 更新用户头像
// 注意：此方法暂未实现，预留接口
func (s *userInfoServiceImpl) UpdateAvatar(ctx context.Context, req *dto.UpdateAvatarRequest) (*dto.UpdateAvatarResponse, error) {
	return nil, status.Error(codes.Unimplemented, "更新头像功能暂未实现")
}

// ChangePassword 修改密码
// 注意：此方法暂未实现，预留接口
func (s *userInfoServiceImpl) ChangePassword(ctx context.Context, req *dto.ChangePasswordRequest) error {
	return status.Error(codes.Unimplemented, "修改密码功能暂未实现")
}

// BindEmail 绑定邮箱
// 注意：此方法暂未实现，预留接口
func (s *userInfoServiceImpl) BindEmail(ctx context.Context, req *dto.BindEmailRequest) error {
	return status.Error(codes.Unimplemented, "绑定邮箱功能暂未实现")
}

// BatchGetUsers 批量获取用户信息
// 用于内部服务调用，如消息服务查询发送者信息、群服务查询成员信息
func (s *userInfoServiceImpl) BatchGetUsers(ctx context.Context, req *dto.BatchGetUsersRequest) (*dto.BatchGetUsersResponse, error) {
	if len(req.UserUUIDs) == 0 {
		return &dto.BatchGetUsersResponse{
			Users: make(map[string]*dto.UserInfo),
		}, nil
	}

	logger.Debug(ctx, "批量查询用户信息",
		logger.Int("count", len(req.UserUUIDs)),
	)

	// 批量查询用户
	users, err := s.userRepo.BatchGetByUUIDs(ctx, req.UserUUIDs)
	if err != nil {
		logger.Error(ctx, "批量查询用户失败",
			logger.Int("count", len(req.UserUUIDs)),
			logger.ErrorField("error", err),
		)
		return nil, status.Error(codes.Internal, "数据库查询失败")
	}

	// 转换为DTO并构建map
	result := make(map[string]*dto.UserInfo)
	for _, user := range users {
		result[user.Uuid] = dto.ConvertModelToUserInfo(user)
	}

	logger.Debug(ctx, "批量查询用户成功",
		logger.Int("requested", len(req.UserUUIDs)),
		logger.Int("found", len(result)),
	)

	return &dto.BatchGetUsersResponse{
		Users: result,
	}, nil
}
