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

// friendServiceImpl 好友关系服务实现
type friendServiceImpl struct {
	userRepo   repository.IUserRepository
	friendRepo repository.IFriendRepository
	applyRepo  repository.IApplyRepository
}

// NewFriendService 创建好友服务实例
func NewFriendService(
	userRepo repository.IUserRepository,
	friendRepo repository.IFriendRepository,
	applyRepo repository.IApplyRepository,
) FriendService {
	return &friendServiceImpl{
		userRepo:   userRepo,
		friendRepo: friendRepo,
		applyRepo:  applyRepo,
	}
}

// SearchUser 搜索用户
func (s *friendServiceImpl) SearchUser(ctx context.Context, req *pb.SearchUserRequest) (*pb.SearchUserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "搜索用户功能暂未实现")
}

// SendFriendApply 发送好友申请
func (s *friendServiceImpl) SendFriendApply(ctx context.Context, req *pb.SendFriendApplyRequest) (*pb.SendFriendApplyResponse, error) {
	return nil, status.Error(codes.Unimplemented, "发送好友申请功能暂未实现")
}

// GetFriendApplyList 获取好友申请列表
func (s *friendServiceImpl) GetFriendApplyList(ctx context.Context, req *pb.GetFriendApplyListRequest) (*pb.GetFriendApplyListResponse, error) {
	return nil, status.Error(codes.Unimplemented, "获取好友申请列表功能暂未实现")
}

// GetSentApplyList 获取发出的申请列表
func (s *friendServiceImpl) GetSentApplyList(ctx context.Context, req *pb.GetSentApplyListRequest) (*pb.GetSentApplyListResponse, error) {
	return nil, status.Error(codes.Unimplemented, "获取发出的申请列表功能暂未实现")
}

// HandleFriendApply 处理好友申请
func (s *friendServiceImpl) HandleFriendApply(ctx context.Context, req *pb.HandleFriendApplyRequest) error {
	return status.Error(codes.Unimplemented, "处理好友申请功能暂未实现")
}

// GetUnreadApplyCount 获取未读申请数量
func (s *friendServiceImpl) GetUnreadApplyCount(ctx context.Context, req *pb.GetUnreadApplyCountRequest) (*pb.GetUnreadApplyCountResponse, error) {
	return nil, status.Error(codes.Unimplemented, "获取未读申请数量功能暂未实现")
}

// MarkApplyAsRead 标记申请已读
func (s *friendServiceImpl) MarkApplyAsRead(ctx context.Context, req *pb.MarkApplyAsReadRequest) error {
	return status.Error(codes.Unimplemented, "标记申请已读功能暂未实现")
}

// GetFriendList 获取好友列表
func (s *friendServiceImpl) GetFriendList(ctx context.Context, req *pb.GetFriendListRequest) (*pb.GetFriendListResponse, error) {
	return nil, status.Error(codes.Unimplemented, "获取好友列表功能暂未实现")
}

// SyncFriendList 好友增量同步
func (s *friendServiceImpl) SyncFriendList(ctx context.Context, req *pb.SyncFriendListRequest) (*pb.SyncFriendListResponse, error) {
	return nil, status.Error(codes.Unimplemented, "好友增量同步功能暂未实现")
}

// DeleteFriend 删除好友
func (s *friendServiceImpl) DeleteFriend(ctx context.Context, req *pb.DeleteFriendRequest) error {
	return status.Error(codes.Unimplemented, "删除好友功能暂未实现")
}

// SetFriendRemark 设置好友备注
func (s *friendServiceImpl) SetFriendRemark(ctx context.Context, req *pb.SetFriendRemarkRequest) error {
	return status.Error(codes.Unimplemented, "设置好友备注功能暂未实现")
}

// SetFriendTag 设置好友标签
func (s *friendServiceImpl) SetFriendTag(ctx context.Context, req *pb.SetFriendTagRequest) error {
	return status.Error(codes.Unimplemented, "设置好友标签功能暂未实现")
}

// GetTagList 获取标签列表
func (s *friendServiceImpl) GetTagList(ctx context.Context, req *pb.GetTagListRequest) (*pb.GetTagListResponse, error) {
	return nil, status.Error(codes.Unimplemented, "获取标签列表功能暂未实现")
}

// CheckIsFriend 判断是否好友
func (s *friendServiceImpl) CheckIsFriend(ctx context.Context, req *pb.CheckIsFriendRequest) (*pb.CheckIsFriendResponse, error) {
	isFriend, err := s.friendRepo.IsFriend(ctx, req.UserUuid, req.PeerUuid)
	if err != nil {
		logger.Error(ctx, "判断是否好友失败",
			logger.String("user_uuid", req.UserUuid),
			logger.String("peer_uuid", req.PeerUuid),
			logger.ErrorField("error", err),
		)
		return nil, status.Error(codes.Internal, strconv.Itoa(consts.CodeInternalError))
	}
	return &pb.CheckIsFriendResponse{
		IsFriend: isFriend,
	}, nil
}

// GetRelationStatus 获取关系状态
func (s *friendServiceImpl) GetRelationStatus(ctx context.Context, req *pb.GetRelationStatusRequest) (*pb.GetRelationStatusResponse, error) {
	return nil, status.Error(codes.Unimplemented, "获取关系状态功能暂未实现")
}
