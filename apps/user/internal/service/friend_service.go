package service

import (
	"ChatServer/apps/user/internal/dto"
	"ChatServer/apps/user/internal/repository"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// friendServiceImpl 好友关系服务实现
type friendServiceImpl struct {
	userRepo     repository.UserRepository
	relationRepo repository.RelationRepository
	applyRepo    repository.ApplyRequestRepository
}

// NewFriendService 创建好友服务实例
func NewFriendService(
	userRepo repository.UserRepository,
	relationRepo repository.RelationRepository,
	applyRepo repository.ApplyRequestRepository,
) FriendService {
	return &friendServiceImpl{
		userRepo:     userRepo,
		relationRepo: relationRepo,
		applyRepo:    applyRepo,
	}
}

// SearchUser 搜索用户
// 注意：此方法暂未实现，预留接口
func (s *friendServiceImpl) SearchUser(ctx context.Context, req *dto.SearchUserRequest) (*dto.SearchUserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "搜索用户功能暂未实现")
}

// SendFriendRequest 发送好友申请
// 注意：此方法暂未实现，预留接口
func (s *friendServiceImpl) SendFriendRequest(ctx context.Context, req *dto.SendFriendRequestRequest) (*dto.SendFriendRequestResponse, error) {
	return nil, status.Error(codes.Unimplemented, "发送好友申请功能暂未实现")
}

// HandleFriendRequest 处理好友申请
// 注意：此方法暂未实现，预留接口
func (s *friendServiceImpl) HandleFriendRequest(ctx context.Context, req *dto.HandleFriendRequestRequest) error {
	return status.Error(codes.Unimplemented, "处理好友申请功能暂未实现")
}

// GetFriendRequests 获取好友申请列表
// 注意：此方法暂未实现，预留接口
func (s *friendServiceImpl) GetFriendRequests(ctx context.Context, req *dto.GetFriendRequestsRequest) (*dto.GetFriendRequestsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "获取好友申请列表功能暂未实现")
}

// GetFriendList 获取好友列表
// 注意：此方法暂未实现，预留接口
func (s *friendServiceImpl) GetFriendList(ctx context.Context, req *dto.GetFriendListRequest) (*dto.GetFriendListResponse, error) {
	return nil, status.Error(codes.Unimplemented, "获取好友列表功能暂未实现")
}

// DeleteFriend 删除好友
// 注意：此方法暂未实现，预留接口
func (s *friendServiceImpl) DeleteFriend(ctx context.Context, req *dto.DeleteFriendRequest) error {
	return status.Error(codes.Unimplemented, "删除好友功能暂未实现")
}

// SetFriendRemark 设置好友备注
// 注意：此方法暂未实现，预留接口
func (s *friendServiceImpl) SetFriendRemark(ctx context.Context, req *dto.SetFriendRemarkRequest) error {
	return status.Error(codes.Unimplemented, "设置好友备注功能暂未实现")
}

// BlockUser 拉黑用户
// 注意：此方法暂未实现，预留接口
func (s *friendServiceImpl) BlockUser(ctx context.Context, req *dto.BlockUserRequest) error {
	return status.Error(codes.Unimplemented, "拉黑用户功能暂未实现")
}

// UnblockUser 解除拉黑
// 注意：此方法暂未实现，预留接口
func (s *friendServiceImpl) UnblockUser(ctx context.Context, req *dto.UnblockUserRequest) error {
	return status.Error(codes.Unimplemented, "解除拉黑功能暂未实现")
}

// GetBlacklist 获取黑名单
// 注意：此方法暂未实现，预留接口
func (s *friendServiceImpl) GetBlacklist(ctx context.Context, req *dto.GetBlacklistRequest) (*dto.GetBlacklistResponse, error) {
	return nil, status.Error(codes.Unimplemented, "获取黑名单功能暂未实现")
}
