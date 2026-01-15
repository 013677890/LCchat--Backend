package service

import (
	"ChatServer/apps/user/internal/dto"
	"ChatServer/apps/user/internal/repository"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// deviceServiceImpl 设备会话服务实现
type deviceServiceImpl struct {
	deviceRepo repository.DeviceSessionRepository
}

// NewDeviceService 创建设备服务实例
func NewDeviceService(deviceRepo repository.DeviceSessionRepository) DeviceService {
	return &deviceServiceImpl{
		deviceRepo: deviceRepo,
	}
}

// CreateDeviceSession 创建设备会话
// 注意：此方法暂未实现，预留接口
func (s *deviceServiceImpl) CreateDeviceSession(ctx context.Context, req *dto.CreateDeviceSessionRequest) error {
	return status.Error(codes.Unimplemented, "创建设备会话功能暂未实现")
}

// GetDeviceSessions 获取设备列表
// 注意：此方法暂未实现，预留接口
func (s *deviceServiceImpl) GetDeviceSessions(ctx context.Context, req *dto.GetDeviceSessionsRequest) (*dto.GetDeviceSessionsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "获取设备列表功能暂未实现")
}

// UpdateDeviceOnlineState 更新设备在线状态
// 注意：此方法暂未实现，预留接口
func (s *deviceServiceImpl) UpdateDeviceOnlineState(ctx context.Context, req *dto.UpdateDeviceOnlineStateRequest) error {
	return status.Error(codes.Unimplemented, "更新设备在线状态功能暂未实现")
}

// KickDevice 踢出设备
// 注意：此方法暂未实现，预留接口
func (s *deviceServiceImpl) KickDevice(ctx context.Context, req *dto.KickDeviceRequest) error {
	return status.Error(codes.Unimplemented, "踢出设备功能暂未实现")
}

// GetUsersOnlineState 批量获取用户在线状态
// 注意：此方法暂未实现，预留接口
func (s *deviceServiceImpl) GetUsersOnlineState(ctx context.Context, req *dto.GetUsersOnlineStateRequest) (*dto.GetUsersOnlineStateResponse, error) {
	return nil, status.Error(codes.Unimplemented, "获取用户在线状态功能暂未实现")
}
