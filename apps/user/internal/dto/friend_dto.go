package dto

import (
	pb "ChatServer/apps/user/pb"
	"time"
)

// SearchUserRequest 搜索用户请求DTO
type SearchUserRequest struct {
	Keyword     string        // 搜索关键字(手机号或昵称)
	CurrentUUID string        // 当前用户UUID(用于过滤)
	Pagination  *pb.Pagination // 分页参数
}

// SearchUserResponse 搜索用户响应DTO
type SearchUserResponse struct {
	Users      []*UserInfo              // 用户列表(手机号已脱敏)
	Pagination *pb.PaginationResponse    // 分页信息
}

// SendFriendRequestRequest 发送好友申请请求DTO
type SendFriendRequestRequest struct {
	ApplicantUUID string // 申请人UUID
	TargetUUID    string // 目标用户UUID
	Reason        string // 申请理由
}

// SendFriendRequestResponse 发送好友申请响应DTO
type SendFriendRequestResponse struct {
	RequestID string // 申请ID
}

// HandleFriendRequestRequest 处理好友申请请求DTO
type HandleFriendRequestRequest struct {
	RequestID   int64  // 申请ID
	HandlerUUID string // 处理人UUID
	Action      int32  // 操作(1:同意 2:拒绝)
	Remark      string // 处理备注(可选)
}

// GetFriendRequestsRequest 获取好友申请列表请求DTO
type GetFriendRequestsRequest struct {
	UserUUID   string        // 用户UUID
	Pagination *pb.Pagination // 分页参数
}

// FriendRequestInfo 好友申请信息DTO
type FriendRequestInfo struct {
	ID                int64     // 申请ID
	ApplicantUUID     string    // 申请人UUID
	ApplicantNickname string    // 申请人昵称
	ApplicantAvatar   string    // 申请人头像
	Reason            string    // 申请理由
	Status            int32     // 状态(0:待处理 1:已通过 2:已拒绝 3:已过期)
	IsRead            bool      // 是否已读
	CreatedAt         time.Time // 申请时间
}

// GetFriendRequestsResponse 获取好友申请列表响应DTO
type GetFriendRequestsResponse struct {
	Requests   []*FriendRequestInfo     // 好友申请列表
	Pagination *pb.PaginationResponse    // 分页信息
}

// GetFriendListRequest 获取好友列表请求DTO
type GetFriendListRequest struct {
	UserUUID   string        // 用户UUID
	Pagination *pb.Pagination // 分页参数
}

// FriendInfo 好友信息DTO
type FriendInfo struct {
	UUID      string    // 好友UUID
	Nickname  string    // 昵称
	Avatar    string    // 头像
	Remark    string    // 备注名
	Gender    int32     // 性别
	Signature string    // 个性签名
	CreatedAt time.Time // 添加好友时间
	Telephone string    // 手机号(已脱敏)
}

// GetFriendListResponse 获取好友列表响应DTO
type GetFriendListResponse struct {
	Friends    []*FriendInfo            // 好友列表
	Pagination *pb.PaginationResponse    // 分页信息
}

// DeleteFriendRequest 删除好友请求DTO
type DeleteFriendRequest struct {
	UserUUID   string // 当前用户UUID
	TargetUUID string // 目标用户UUID
}

// SetFriendRemarkRequest 设置好友备注请求DTO
type SetFriendRemarkRequest struct {
	UserUUID   string // 当前用户UUID
	TargetUUID string // 目标用户UUID
	Remark     string // 备注名
}

// BlockUserRequest 拉黑用户请求DTO
type BlockUserRequest struct {
	UserUUID   string // 当前用户UUID
	TargetUUID string // 目标用户UUID
}

// UnblockUserRequest 解除拉黑请求DTO
type UnblockUserRequest struct {
	UserUUID   string // 当前用户UUID
	TargetUUID string // 目标用户UUID
}

// GetBlacklistRequest 获取黑名单请求DTO
type GetBlacklistRequest struct {
	UserUUID   string        // 用户UUID
	Pagination *pb.Pagination // 分页参数
}

// BlacklistUser 黑名单用户信息DTO
type BlacklistUser struct {
	UUID      string    // 用户UUID
	Nickname  string    // 昵称
	Avatar    string    // 头像
	BlockedAt time.Time // 拉黑时间
}

// GetBlacklistResponse 获取黑名单响应DTO
type GetBlacklistResponse struct {
	Blacklist  []*BlacklistUser         // 黑名单列表
	Pagination *pb.PaginationResponse    // 分页信息
}
