package dto

import (
	pb "ChatServer/apps/user/pb"
	"time"
)

// CreateDeviceSessionRequest 创建设备会话请求DTO
type CreateDeviceSessionRequest struct {
	UserUUID   string         // 用户UUID
	DeviceInfo *pb.DeviceInfo // 设备信息
	IP         string         // IP地址
	UserAgent  string         // User-Agent
}

// DeviceSessionInfo 设备会话信息DTO
type DeviceSessionInfo struct {
	DeviceID   string    // 设备ID
	DeviceName string    // 设备名称
	Platform   string    // 平台
	OSVersion  string    // 系统版本
	AppVersion string    // 应用版本
	IP         string    // 登录IP
	Status     int32     // 在线状态(0:在线 1:下线)
	LastSeenAt time.Time // 最后活跃时间
	CreatedAt  time.Time // 登录时间
}

// GetDeviceSessionsRequest 获取设备列表请求DTO
type GetDeviceSessionsRequest struct {
	UserUUID string // 用户UUID
}

// GetDeviceSessionsResponse 获取设备列表响应DTO
type GetDeviceSessionsResponse struct {
	Devices []*DeviceSessionInfo // 设备列表
}

// UpdateDeviceOnlineStateRequest 更新设备在线状态请求DTO
type UpdateDeviceOnlineStateRequest struct {
	UserUUID string // 用户UUID
	DeviceID string // 设备ID
	Status   int32  // 在线状态(0:在线 1:下线)
}

// KickDeviceRequest 踢出设备请求DTO
type KickDeviceRequest struct {
	OperatorUUID   string // 操作人UUID
	TargetDeviceID string // 目标设备ID
}

// BatchGetUsersRequest 批量获取用户信息请求DTO
type BatchGetUsersRequest struct {
	UserUUIDs []string // 用户UUID列表
}

// BatchGetUsersResponse 批量获取用户信息响应DTO
type BatchGetUsersResponse struct {
	Users map[string]*UserInfo // 用户信息映射(UUID -> UserInfo)
}

// OnlineState 在线状态DTO
type OnlineState struct {
	IsOnline     bool      // 是否在线
	DeviceIDs    []string  // 在线的设备ID列表
	LastActiveAt time.Time // 最后活跃时间
}

// GetUsersOnlineStateRequest 获取用户在线状态请求DTO
type GetUsersOnlineStateRequest struct {
	UserUUIDs []string // 用户UUID列表
}

// GetUsersOnlineStateResponse 获取用户在线状态响应DTO
type GetUsersOnlineStateResponse struct {
	OnlineStates map[string]*OnlineState // 在线状态映射(UUID -> OnlineState)
}
