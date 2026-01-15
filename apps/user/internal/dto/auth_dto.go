package dto

import (
	"ChatServer/model"
	pb "ChatServer/apps/user/pb"
)

// LoginRequest 登录请求DTO
type LoginRequest struct {
	Telephone  string           // 手机号
	Password   string           // 密码
	DeviceInfo *pb.DeviceInfo   // 设备信息
}

// LoginResponse 登录响应DTO
type LoginResponse struct {
	UserInfo *UserInfo // 用户信息
}

// RegisterRequest 注册请求DTO
type RegisterRequest struct {
	Telephone  string           // 手机号
	Password   string           // 密码
	VerifyCode string           // 验证码
	Nickname   string           // 昵称（可选）
	DeviceInfo *pb.DeviceInfo   // 设备信息
}

// RegisterResponse 注册响应DTO
type RegisterResponse struct {
	UserUUID  string // 用户UUID
	Telephone string // 手机号
	Nickname  string // 昵称
}

// RefreshTokenRequest 刷新Token请求DTO
type RefreshTokenRequest struct {
	UserUUID string // 用户UUID
	DeviceID string // 设备ID
}

// RefreshTokenResponse 刷新Token响应DTO
type RefreshTokenResponse struct {
	UserInfo *UserInfo // 用户信息
}

// LogoutRequest 登出请求DTO
type LogoutRequest struct {
	UserUUID string // 用户UUID
	DeviceID string // 设备ID
}

// ConvertModelToUserInfo 将数据库模型转换为DTO
func ConvertModelToUserInfo(user *model.UserInfo) *UserInfo {
	if user == nil {
		return nil
	}
	return &UserInfo{
		UUID:      user.Uuid,
		Nickname:  user.Nickname,
		Telephone: user.Telephone,
		Email:     user.Email,
		Avatar:    user.Avatar,
		Gender:    int(user.Gender),
		Signature: user.Signature,
		Birthday:  user.Birthday,
		Status:    int(user.Status),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// ConvertToProtoUserInfo 将DTO转换为Protobuf消息
func ConvertToProtoUserInfo(user *UserInfo) *pb.UserInfo {
	if user == nil {
		return nil
	}
	return &pb.UserInfo{
		Uuid:      user.UUID,
		Nickname:  user.Nickname,
		Telephone: user.Telephone,
		Email:     user.Email,
		Avatar:    user.Avatar,
		Gender:    int32(user.Gender),
		Signature: user.Signature,
		Birthday:  user.Birthday,
		Status:    int32(user.Status),
		CreatedAt: user.CreatedAt.Unix() * 1000, // 转换为毫秒时间戳
		UpdatedAt: user.UpdatedAt.Unix() * 1000,
	}
}
