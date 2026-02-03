package service

import (
	"ChatServer/apps/gateway/internal/dto"
	"ChatServer/apps/gateway/internal/pb"
	"ChatServer/apps/gateway/internal/utils"
	userpb "ChatServer/apps/user/pb"
	"ChatServer/consts"
	"ChatServer/pkg/logger"
	"context"
	"time"
)

// BlacklistServiceImpl 黑名单服务实现
type BlacklistServiceImpl struct {
	userClient pb.UserServiceClient
}

// NewBlacklistService 创建黑名单服务实例
// userClient: 用户服务 gRPC 客户端
func NewBlacklistService(userClient pb.UserServiceClient) BlacklistService {
	return &BlacklistServiceImpl{
		userClient: userClient,
	}
}

// AddBlacklist 拉黑用户
func (s *BlacklistServiceImpl) AddBlacklist(ctx context.Context, req *dto.AddBlacklistRequest) (*dto.AddBlacklistResponse, error) {
	startTime := time.Now()

	grpcReq := dto.ConvertToProtoAddBlacklistRequest(req)
	grpcResp, err := s.userClient.AddBlacklist(ctx, grpcReq)
	if err != nil {
		code := utils.ExtractErrorCode(err)
		logger.Error(ctx, "调用用户服务 gRPC 失败",
			logger.ErrorField("error", err),
			logger.Int("business_code", code),
			logger.String("business_message", consts.GetMessage(code)),
			logger.Duration("duration", time.Since(startTime)),
		)
		return nil, err
	}

	return dto.ConvertAddBlacklistResponseFromProto(grpcResp), nil
}

// RemoveBlacklist 取消拉黑
func (s *BlacklistServiceImpl) RemoveBlacklist(ctx context.Context, req *dto.RemoveBlacklistRequest) (*dto.RemoveBlacklistResponse, error) {
	startTime := time.Now()

	grpcReq := dto.ConvertToProtoRemoveBlacklistRequest(req)
	grpcResp, err := s.userClient.RemoveBlacklist(ctx, grpcReq)
	if err != nil {
		code := utils.ExtractErrorCode(err)
		logger.Error(ctx, "调用用户服务 gRPC 失败",
			logger.ErrorField("error", err),
			logger.Int("business_code", code),
			logger.String("business_message", consts.GetMessage(code)),
			logger.Duration("duration", time.Since(startTime)),
		)
		return nil, err
	}

	return dto.ConvertRemoveBlacklistResponseFromProto(grpcResp), nil
}

// GetBlacklistList 获取黑名单列表
func (s *BlacklistServiceImpl) GetBlacklistList(ctx context.Context, req *dto.GetBlacklistListRequest) (*dto.GetBlacklistListResponse, error) {
	startTime := time.Now()

	grpcReq := &userpb.GetBlacklistListRequest{
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	grpcResp, err := s.userClient.GetBlacklistList(ctx, grpcReq)
	if err != nil {
		code := utils.ExtractErrorCode(err)
		logger.Error(ctx, "调用用户服务 gRPC 失败",
			logger.ErrorField("error", err),
			logger.Int("business_code", code),
			logger.String("business_message", consts.GetMessage(code)),
			logger.Duration("duration", time.Since(startTime)),
		)
		return nil, err
	}

	return dto.ConvertGetBlacklistListResponseFromProto(grpcResp), nil
}

// CheckIsBlacklist 判断是否拉黑
func (s *BlacklistServiceImpl) CheckIsBlacklist(ctx context.Context, req *dto.CheckIsBlacklistRequest) (*dto.CheckIsBlacklistResponse, error) {
	startTime := time.Now()

	grpcReq := dto.ConvertToProtoCheckIsBlacklistRequest(req)
	grpcResp, err := s.userClient.CheckIsBlacklist(ctx, grpcReq)
	if err != nil {
		code := utils.ExtractErrorCode(err)
		logger.Error(ctx, "调用用户服务 gRPC 失败",
			logger.ErrorField("error", err),
			logger.Int("business_code", code),
			logger.String("business_message", consts.GetMessage(code)),
			logger.Duration("duration", time.Since(startTime)),
		)
		return nil, err
	}

	return dto.ConvertCheckIsBlacklistResponseFromProto(grpcResp), nil
}
