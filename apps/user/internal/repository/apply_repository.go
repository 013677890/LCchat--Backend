package repository

import (
	"ChatServer/model"
	"context"

	"gorm.io/gorm"
)

// applyRequestRepositoryImpl 好友申请数据访问层实现
type applyRequestRepositoryImpl struct {
	db *gorm.DB
}

// NewApplyRequestRepository 创建好友申请仓储实例
func NewApplyRequestRepository(db *gorm.DB) ApplyRequestRepository {
	return &applyRequestRepositoryImpl{db: db}
}

// Create 创建好友申请
func (r *applyRequestRepositoryImpl) Create(ctx context.Context, apply *model.ApplyRequest) (*model.ApplyRequest, error) {
	err := r.db.WithContext(ctx).Create(apply).Error
	if err != nil {
		return nil, err
	}
	return apply, nil
}

// GetByID 根据ID获取好友申请
func (r *applyRequestRepositoryImpl) GetByID(ctx context.Context, id int64) (*model.ApplyRequest, error) {
	var apply model.ApplyRequest
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&apply).Error
	if err != nil {
		return nil, err
	}
	return &apply, nil
}

// GetPendingList 获取待处理的好友申请列表
func (r *applyRequestRepositoryImpl) GetPendingList(ctx context.Context, targetUUID string, page, pageSize int) ([]*model.ApplyRequest, int64, error) {
	var applies []*model.ApplyRequest
	var total int64
	
	query := r.db.WithContext(ctx).
		Model(&model.ApplyRequest{}).
		Where("target_uuid = ? AND apply_type = 0 AND status = 0", targetUUID)
	
	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&applies).Error
	
	if err != nil {
		return nil, 0, err
	}
	
	return applies, total, nil
}

// UpdateStatus 更新申请状态
func (r *applyRequestRepositoryImpl) UpdateStatus(ctx context.Context, id int64, status int, remark string) error {
	updates := map[string]interface{}{
		"status": status,
		"is_read": true,
	}
	
	if remark != "" {
		updates["handle_remark"] = remark
	}
	
	return r.db.WithContext(ctx).
		Model(&model.ApplyRequest{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// ExistsPendingRequest 检查是否存在待处理的申请
func (r *applyRequestRepositoryImpl) ExistsPendingRequest(ctx context.Context, applicantUUID, targetUUID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.ApplyRequest{}).
		Where("applicant_uuid = ? AND target_uuid = ? AND apply_type = 0 AND status = 0", applicantUUID, targetUUID).
		Count(&count).Error
	
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}
