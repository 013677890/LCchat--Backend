package service

import (
	"ChatServer/apps/user/internal/repository"
	pb "ChatServer/apps/user/pb"
	"ChatServer/consts"
	"ChatServer/pkg/logger"
	"context"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// blacklistServiceImpl 黑名单服务实现
type blacklistServiceImpl struct {
	blacklistRepo repository.IBlacklistRepository
}

// NewBlacklistService 创建黑名单服务实例
func NewBlacklistService(blacklistRepo repository.IBlacklistRepository) BlacklistService {
	return &blacklistServiceImpl{
		blacklistRepo: blacklistRepo,
	}
}

// AddBlacklist 拉黑用户
func (s *blacklistServiceImpl) AddBlacklist(ctx context.Context, req *pb.AddBlacklistRequest) error {
	// 1. 从context中获取当前用户UUID
	currentUserUUID, ok := ctx.Value("user_uuid").(string)
	if !ok || currentUserUUID == "" {
		logger.Error(ctx, "获取用户UUID失败")
		return status.Error(codes.Unauthenticated, strconv.Itoa(consts.CodeUnauthorized))
	}

	// 2. 参数校验
	if req == nil || req.TargetUuid == "" {
		return status.Error(codes.InvalidArgument, strconv.Itoa(consts.CodeParamError))
	}

	// 3. 不能拉黑自己
	if req.TargetUuid == currentUserUUID {
		return status.Error(codes.InvalidArgument, strconv.Itoa(consts.CodeCannotBlacklistSelf))
	}

	// 4. 判断是否已在黑名单中
	isBlocked, err := s.blacklistRepo.IsBlocked(ctx, currentUserUUID, req.TargetUuid)
	if err != nil {
		logger.Error(ctx, "检查黑名单失败",
			logger.String("user_uuid", currentUserUUID),
			logger.String("target_uuid", req.TargetUuid),
			logger.ErrorField("error", err),
		)
		return status.Error(codes.Internal, strconv.Itoa(consts.CodeInternalError))
	}
	if isBlocked {
		return status.Error(codes.AlreadyExists, strconv.Itoa(consts.CodeAlreadyInBlacklist))
	}

	// 5. 拉黑用户
	if err := s.blacklistRepo.AddBlacklist(ctx, currentUserUUID, req.TargetUuid); err != nil {
		logger.Error(ctx, "拉黑用户失败",
			logger.String("user_uuid", currentUserUUID),
			logger.String("target_uuid", req.TargetUuid),
			logger.ErrorField("error", err),
		)
		return status.Error(codes.Internal, strconv.Itoa(consts.CodeInternalError))
	}

	logger.Info(ctx, "拉黑用户成功",
		logger.String("user_uuid", currentUserUUID),
		logger.String("target_uuid", req.TargetUuid),
	)

	return nil
}

// RemoveBlacklist 取消拉黑
func (s *blacklistServiceImpl) RemoveBlacklist(ctx context.Context, req *pb.RemoveBlacklistRequest) error {
	return status.Error(codes.Unimplemented, "取消拉黑功能暂未实现")
}

// GetBlacklistList 获取黑名单列表
func (s *blacklistServiceImpl) GetBlacklistList(ctx context.Context, req *pb.GetBlacklistListRequest) (*pb.GetBlacklistListResponse, error) {
	return nil, status.Error(codes.Unimplemented, "获取黑名单列表功能暂未实现")
}

// CheckIsBlacklist 判断是否拉黑
func (s *blacklistServiceImpl) CheckIsBlacklist(ctx context.Context, req *pb.CheckIsBlacklistRequest) (*pb.CheckIsBlacklistResponse, error) {
	return nil, status.Error(codes.Unimplemented, "判断是否拉黑功能暂未实现")
}
