package pb

import (
	userpb "ChatServer/apps/user/pb"
	"context"
)

// UserServiceClient 用户服务 gRPC 客户端接口
// 职责：封装对用户服务的 gRPC 调用
type UserServiceClient interface {
	// Login 用户登录
	Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginResponse, error)

	// TODO: 后续扩展其他用户服务方法
	// GetUserInfo(ctx context.Context, req *userpb.GetUserInfoRequest) (*userpb.GetUserInfoResponse, error)
	// Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.RegisterResponse, error)
}
