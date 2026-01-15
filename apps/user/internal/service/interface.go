package service

import (
	"ChatServer/apps/user/internal/dto"
	"context"
)

// AuthService 认证服务接口
type AuthService interface {
	// Login 用户登录
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
	
	// Register 用户注册
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error)
	
	// RefreshToken 刷新Token
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error)
	
	// Logout 用户登出
	Logout(ctx context.Context, req *dto.LogoutRequest) error
	
	// SendSmsCode 发送短信验证码
	SendSmsCode(ctx context.Context, telephone string, codeType int32) error
	
	// ValidateSmsCode 验证短信验证码
	ValidateSmsCode(ctx context.Context, telephone, code string, codeType int32) (bool, error)
}

// UserQueryService 用户信息服务接口
type UserQueryService interface {
	// GetUserInfo 获取用户信息
	GetUserInfo(ctx context.Context, req *dto.GetUserInfoRequest) (*dto.GetUserInfoResponse, error)
	
	// UpdateUserInfo 更新用户信息
	UpdateUserInfo(ctx context.Context, req *dto.UpdateUserInfoRequest) (*dto.UpdateUserInfoResponse, error)
	
	// UpdateAvatar 更新用户头像
	UpdateAvatar(ctx context.Context, req *dto.UpdateAvatarRequest) (*dto.UpdateAvatarResponse, error)
	
	// ChangePassword 修改密码
	ChangePassword(ctx context.Context, req *dto.ChangePasswordRequest) error
	
	// BindEmail 绑定邮箱
	BindEmail(ctx context.Context, req *dto.BindEmailRequest) error
	
	// BatchGetUsers 批量获取用户信息
	BatchGetUsers(ctx context.Context, req *dto.BatchGetUsersRequest) (*dto.BatchGetUsersResponse, error)
}

// FriendService 好友关系服务接口
type FriendService interface {
	// SearchUser 搜索用户
	SearchUser(ctx context.Context, req *dto.SearchUserRequest) (*dto.SearchUserResponse, error)
	
	// SendFriendRequest 发送好友申请
	SendFriendRequest(ctx context.Context, req *dto.SendFriendRequestRequest) (*dto.SendFriendRequestResponse, error)
	
	// HandleFriendRequest 处理好友申请
	HandleFriendRequest(ctx context.Context, req *dto.HandleFriendRequestRequest) error
	
	// GetFriendRequests 获取好友申请列表
	GetFriendRequests(ctx context.Context, req *dto.GetFriendRequestsRequest) (*dto.GetFriendRequestsResponse, error)
	
	// GetFriendList 获取好友列表
	GetFriendList(ctx context.Context, req *dto.GetFriendListRequest) (*dto.GetFriendListResponse, error)
	
	// DeleteFriend 删除好友
	DeleteFriend(ctx context.Context, req *dto.DeleteFriendRequest) error
	
	// SetFriendRemark 设置好友备注
	SetFriendRemark(ctx context.Context, req *dto.SetFriendRemarkRequest) error
	
	// BlockUser 拉黑用户
	BlockUser(ctx context.Context, req *dto.BlockUserRequest) error
	
	// UnblockUser 解除拉黑
	UnblockUser(ctx context.Context, req *dto.UnblockUserRequest) error
	
	// GetBlacklist 获取黑名单
	GetBlacklist(ctx context.Context, req *dto.GetBlacklistRequest) (*dto.GetBlacklistResponse, error)
}

// DeviceService 设备会话服务接口
type DeviceService interface {
	// CreateDeviceSession 创建设备会话
	CreateDeviceSession(ctx context.Context, req *dto.CreateDeviceSessionRequest) error
	
	// GetDeviceSessions 获取设备列表
	GetDeviceSessions(ctx context.Context, req *dto.GetDeviceSessionsRequest) (*dto.GetDeviceSessionsResponse, error)
	
	// UpdateDeviceOnlineState 更新设备在线状态
	UpdateDeviceOnlineState(ctx context.Context, req *dto.UpdateDeviceOnlineStateRequest) error
	
	// KickDevice 踢出设备
	KickDevice(ctx context.Context, req *dto.KickDeviceRequest) error
	
	// GetUsersOnlineState 批量获取用户在线状态
	GetUsersOnlineState(ctx context.Context, req *dto.GetUsersOnlineStateRequest) (*dto.GetUsersOnlineStateResponse, error)
}
