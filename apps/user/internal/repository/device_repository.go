package repository

import (
	"ChatServer/model"
	"context"
	"time"

	"gorm.io/gorm"
)

// deviceSessionRepositoryImpl 设备会话数据访问层实现
type deviceSessionRepositoryImpl struct {
	db *gorm.DB
}

// NewDeviceSessionRepository 创建设备会话仓储实例
func NewDeviceSessionRepository(db *gorm.DB) DeviceSessionRepository {
	return &deviceSessionRepositoryImpl{db: db}
}

// Create 创建设备会话
func (r *deviceSessionRepositoryImpl) Create(ctx context.Context, session *model.DeviceSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetByUserUUID 获取用户的所有设备会话
func (r *deviceSessionRepositoryImpl) GetByUserUUID(ctx context.Context, userUUID string) ([]*model.DeviceSession, error) {
	var sessions []*model.DeviceSession
	err := r.db.WithContext(ctx).
		Where("user_uuid = ?", userUUID).
		Order("last_seen_at DESC").
		Find(&sessions).Error
	
	if err != nil {
		return nil, err
	}
	
	return sessions, nil
}

// GetByDeviceID 根据设备ID获取会话
func (r *deviceSessionRepositoryImpl) GetByDeviceID(ctx context.Context, userUUID, deviceID string) (*model.DeviceSession, error) {
	var session model.DeviceSession
	err := r.db.WithContext(ctx).
		Where("user_uuid = ? AND device_id = ?", userUUID, deviceID).
		First(&session).Error
	
	if err != nil {
		return nil, err
	}
	
	return &session, nil
}

// UpdateOnlineStatus 更新在线状态
func (r *deviceSessionRepositoryImpl) UpdateOnlineStatus(ctx context.Context, userUUID, deviceID string, status int) error {
	return r.db.WithContext(ctx).
		Model(&model.DeviceSession{}).
		Where("user_uuid = ? AND device_id = ?", userUUID, deviceID).
		Updates(map[string]interface{}{
			"status":       status,
			"last_seen_at": time.Now(),
		}).Error
}

// UpdateLastSeen 更新最后活跃时间
func (r *deviceSessionRepositoryImpl) UpdateLastSeen(ctx context.Context, userUUID, deviceID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.DeviceSession{}).
		Where("user_uuid = ? AND device_id = ?", userUUID, deviceID).
		Update("last_seen_at", &now).Error
}

// Delete 删除设备会话
func (r *deviceSessionRepositoryImpl) Delete(ctx context.Context, userUUID, deviceID string) error {
	return r.db.WithContext(ctx).
		Where("user_uuid = ? AND device_id = ?", userUUID, deviceID).
		Delete(&model.DeviceSession{}).Error
}

// GetOnlineDevices 获取在线设备列表
func (r *deviceSessionRepositoryImpl) GetOnlineDevices(ctx context.Context, userUUID string) ([]*model.DeviceSession, error) {
	var sessions []*model.DeviceSession
	err := r.db.WithContext(ctx).
		Where("user_uuid = ? AND status = 0", userUUID).
		Order("last_seen_at DESC").
		Find(&sessions).Error
	
	if err != nil {
		return nil, err
	}
	
	return sessions, nil
}

// BatchGetOnlineStatus 批量获取用户在线状态
func (r *deviceSessionRepositoryImpl) BatchGetOnlineStatus(ctx context.Context, userUUIDs []string) (map[string][]*model.DeviceSession, error) {
	if len(userUUIDs) == 0 {
		return make(map[string][]*model.DeviceSession), nil
	}
	
	var sessions []*model.DeviceSession
	err := r.db.WithContext(ctx).
		Where("user_uuid IN ? AND status = 0", userUUIDs).
		Order("user_uuid, last_seen_at DESC").
		Find(&sessions).Error
	
	if err != nil {
		return nil, err
	}
	
	// 按用户UUID分组
	result := make(map[string][]*model.DeviceSession)
	for _, session := range sessions {
		result[session.UserUuid] = append(result[session.UserUuid], session)
	}
	
	return result, nil
}
