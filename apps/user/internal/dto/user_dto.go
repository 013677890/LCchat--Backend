package dto

import (
	"time"
)

// UserInfo 用户信息DTO
type UserInfo struct {
	UUID      string    // 用户UUID
	Nickname  string    // 昵称
	Telephone string    // 手机号
	Email     string    // 邮箱
	Avatar    string    // 头像URL
	Gender    int       // 性别(0:男 1:女 2:未知)
	Signature string    // 个性签名
	Birthday  string    // 生日(YYYY-MM-DD)
	Status    int       // 状态(0:正常 1:禁用)
	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 更新时间
}

// GetUserInfoRequest 获取用户信息请求DTO
type GetUserInfoRequest struct {
	UserUUID  string // 用户UUID（优先）
	Telephone string // 手机号（备选）
}

// GetUserInfoResponse 获取用户信息响应DTO
type GetUserInfoResponse struct {
	UserInfo *UserInfo // 用户信息
}

// UpdateUserInfoRequest 更新用户信息请求DTO
type UpdateUserInfoRequest struct {
	UserUUID  string // 用户UUID
	Nickname  string // 昵称（可选）
	Signature string // 个性签名（可选）
	Birthday  string // 生日(YYYY-MM-DD,可选)
	Gender    int    // 性别(0:男 1:女 2:未知,可选)
}

// UpdateUserInfoResponse 更新用户信息响应DTO
type UpdateUserInfoResponse struct {
	UserInfo *UserInfo // 更新后的用户信息
}

// UpdateAvatarRequest 更新头像请求DTO
type UpdateAvatarRequest struct {
	UserUUID  string // 用户UUID
	AvatarURL string // 头像URL
}

// UpdateAvatarResponse 更新头像响应DTO
type UpdateAvatarResponse struct {
	AvatarURL string // 更新后的头像URL
}

// ChangePasswordRequest 修改密码请求DTO
type ChangePasswordRequest struct {
	UserUUID    string // 用户UUID
	OldPassword string // 旧密码
	NewPassword string // 新密码
}

// BindEmailRequest 绑定邮箱请求DTO
type BindEmailRequest struct {
	UserUUID   string // 用户UUID
	Email      string // 邮箱
	VerifyCode string // 验证码
}
