package domain

import (
	"ChatServer/apps/user/internal/dto"
	"ChatServer/apps/user/internal/service"
	"ChatServer/pkg/logger"
	"context"
)

// LoginDomain 登录业务领域
// 职责：
//   - 协调多个Service完成登录相关的复杂业务流程
//   - 例如：登录后创建设备会话、记录登录日志等
type LoginDomain struct {
	authService   service.AuthService
	deviceService service.DeviceService
}

// NewLoginDomain 创建登录领域实例
func NewLoginDomain(
	authService service.AuthService,
	deviceService service.DeviceService,
) *LoginDomain {
	return &LoginDomain{
		authService:   authService,
		deviceService: deviceService,
	}
}

// Login 执行登录业务流程
// 业务流程：
//   1. 调用认证服务进行用户登录验证
//   2. 登录成功后，创建设备会话记录（可选，取决于业务需求）
//   3. 返回登录结果
func (d *LoginDomain) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// 1. 执行登录认证
	resp, err := d.authService.Login(ctx, req)
	if err != nil {
		return nil, err
	}

	// 2. 登录成功后的后续操作
	// 注意：创建设备会话等操作可以异步执行，或者放在Gateway层处理
	logger.Info(ctx, "登录业务流程完成",
		logger.String("user_uuid", resp.UserInfo.UUID),
		logger.String("device_id", req.DeviceInfo.GetDeviceId()),
	)

	// TODO: 如需创建设备会话，可以在这里调用
	// deviceReq := &dto.CreateDeviceSessionRequest{
	// 	UserUUID:   resp.UserInfo.UUID,
	// 	DeviceInfo: req.DeviceInfo,
	// 	...
	// }
	// _ = d.deviceService.CreateDeviceSession(ctx, deviceReq)

	return resp, nil
}
