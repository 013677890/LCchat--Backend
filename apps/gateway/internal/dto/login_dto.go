package dto

import (
	userpb "ChatServer/apps/user/pb"
)

// LoginRequest 登录请求 DTO
type LoginRequest struct {
	Telephone  string     `json:"telephone" binding:"required,len=11"`      // 手机号
	Password   string     `json:"password" binding:"required,min=8,max=16"` // 密码
	DeviceInfo DeviceInfo `json:"deviceInfo"`                               // 设备信息
}

// DeviceInfo 设备信息 DTO
type DeviceInfo struct {
	Platform    string `json:"platform"`    // 平台(iOS/Android/Web)
	OSVersion   string `json:"osVersion"`   // 系统版本
	AppVersion  string `json:"appVersion"`  // 应用版本
	DeviceModel string `json:"deviceModel"` // 设备型号
}

// UserInfo 用户信息 DTO
type UserInfo struct {
	UUID      string `json:"uuid"`      // 用户UUID
	Nickname  string `json:"nickname"`  // 昵称
	Telephone string `json:"telephone"` // 手机号
	Email     string `json:"email"`     // 邮箱
	Avatar    string `json:"avatar"`    // 头像
	Gender    int8   `json:"gender"`    // 性别
	Signature string `json:"signature"` // 个性签名
	Birthday  string `json:"birthday"`  // 生日
}

// LoginResponse 登录响应 DTO
type LoginResponse struct {
	AccessToken  string   `json:"accessToken"`  // 访问令牌
	RefreshToken string   `json:"refreshToken"` // 刷新令牌
	TokenType    string   `json:"tokenType"`    // 令牌类型
	ExpiresIn    int64    `json:"expiresIn"`    // 过期时间(秒)
	UserInfo     UserInfo `json:"userInfo"`     // 用户信息
}

// ==================== DTO 转换函数 ====================

// ConvertToProtoLoginRequest 将 DTO 登录请求转换为 Protobuf 请求
func ConvertToProtoLoginRequest(dto *LoginRequest) *userpb.LoginRequest {
	if dto == nil {
		return nil
	}

	return &userpb.LoginRequest{
		Telephone: dto.Telephone,
		Password:  dto.Password,
		DeviceInfo: &userpb.DeviceInfo{
			Platform:    dto.DeviceInfo.Platform,
			OsVersion:   dto.DeviceInfo.OSVersion,
			AppVersion:  dto.DeviceInfo.AppVersion,
			DeviceModel: dto.DeviceInfo.DeviceModel,
		},
	}
}

// ConvertUserInfoFromProto 将 Protobuf 用户信息转换为 DTO
func ConvertUserInfoFromProto(pb *userpb.UserInfo) UserInfo {
	if pb == nil {
		return UserInfo{}
	}

	return UserInfo{
		UUID:      pb.Uuid,
		Nickname:  pb.Nickname,
		Telephone: pb.Telephone,
		Email:     pb.Email,
		Avatar:    pb.Avatar,
		Gender:    int8(pb.Gender),
		Signature: pb.Signature,
		Birthday:  pb.Birthday,
	}
}
