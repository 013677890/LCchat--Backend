package repository

import (
	"ChatServer/model"
	"context"
	"time"

	"gorm.io/gorm"
)

// relationRepositoryImpl 好友关系数据访问层实现
type relationRepositoryImpl struct {
	db *gorm.DB
}

// NewRelationRepository 创建好友关系仓储实例
func NewRelationRepository(db *gorm.DB) RelationRepository {
	return &relationRepositoryImpl{db: db}
}

// GetFriendRelation 获取好友关系
func (r *relationRepositoryImpl) GetFriendRelation(ctx context.Context, userUUID, friendUUID string) (*model.UserRelation, error) {
	var relation model.UserRelation
	err := r.db.WithContext(ctx).
		Where("user_uuid = ? AND peer_uuid = ? AND status = 0", userUUID, friendUUID).
		First(&relation).Error
	if err != nil {
		return nil, err
	}
	return &relation, nil
}

// GetFriendList 获取好友列表
func (r *relationRepositoryImpl) GetFriendList(ctx context.Context, userUUID string, page, pageSize int) ([]*model.UserRelation, int64, error) {
	var relations []*model.UserRelation
	var total int64
	
	query := r.db.WithContext(ctx).
		Model(&model.UserRelation{}).
		Where("user_uuid = ? AND status = 0", userUUID)
	
	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&relations).Error
	
	if err != nil {
		return nil, 0, err
	}
	
	return relations, total, nil
}

// CreateFriendRelation 创建好友关系（双向）
func (r *relationRepositoryImpl) CreateFriendRelation(ctx context.Context, userUUID, friendUUID string) error {
	now := time.Now()
	
	relations := []*model.UserRelation{
		{
			UserUuid:  userUUID,
			PeerUuid:  friendUUID,
			Status:    0,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			UserUuid:  friendUUID,
			PeerUuid:  userUUID,
			Status:    0,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	
	return r.db.WithContext(ctx).Create(&relations).Error
}

// DeleteFriendRelation 删除好友关系（单向）
func (r *relationRepositoryImpl) DeleteFriendRelation(ctx context.Context, userUUID, friendUUID string) error {
	return r.db.WithContext(ctx).
		Model(&model.UserRelation{}).
		Where("user_uuid = ? AND peer_uuid = ?", userUUID, friendUUID).
		Update("status", 2).Error
}

// SetFriendRemark 设置好友备注
func (r *relationRepositoryImpl) SetFriendRemark(ctx context.Context, userUUID, friendUUID, remark string) error {
	return r.db.WithContext(ctx).
		Model(&model.UserRelation{}).
		Where("user_uuid = ? AND peer_uuid = ?", userUUID, friendUUID).
		Update("remark", remark).Error
}

// BlockUser 拉黑用户
func (r *relationRepositoryImpl) BlockUser(ctx context.Context, userUUID, targetUUID string) error {
	// 先查询是否存在关系
	var relation model.UserRelation
	err := r.db.WithContext(ctx).
		Where("user_uuid = ? AND peer_uuid = ?", userUUID, targetUUID).
		First(&relation).Error
	
	if err == gorm.ErrRecordNotFound {
		// 不存在关系，创建新的拉黑关系
		relation = model.UserRelation{
			UserUuid:  userUUID,
			PeerUuid:  targetUUID,
			Status:    1, // 拉黑状态
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		return r.db.WithContext(ctx).Create(&relation).Error
	}
	
	if err != nil {
		return err
	}
	
	// 存在关系，更新为拉黑状态
	return r.db.WithContext(ctx).
		Model(&relation).
		Update("status", 1).Error
}

// UnblockUser 解除拉黑
func (r *relationRepositoryImpl) UnblockUser(ctx context.Context, userUUID, targetUUID string) error {
	return r.db.WithContext(ctx).
		Model(&model.UserRelation{}).
		Where("user_uuid = ? AND peer_uuid = ? AND status = 1", userUUID, targetUUID).
		Update("status", 2).Error
}

// GetBlacklist 获取黑名单列表
func (r *relationRepositoryImpl) GetBlacklist(ctx context.Context, userUUID string, page, pageSize int) ([]*model.UserRelation, int64, error) {
	var relations []*model.UserRelation
	var total int64
	
	query := r.db.WithContext(ctx).
		Model(&model.UserRelation{}).
		Where("user_uuid = ? AND status = 1", userUUID)
	
	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("updated_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&relations).Error
	
	if err != nil {
		return nil, 0, err
	}
	
	return relations, total, nil
}

// IsBlocked 检查是否被拉黑
func (r *relationRepositoryImpl) IsBlocked(ctx context.Context, userUUID, targetUUID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserRelation{}).
		Where("user_uuid = ? AND peer_uuid = ? AND status = 1", targetUUID, userUUID).
		Count(&count).Error
	
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// IsFriend 检查是否是好友
func (r *relationRepositoryImpl) IsFriend(ctx context.Context, userUUID, friendUUID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserRelation{}).
		Where("user_uuid = ? AND peer_uuid = ? AND status = 0", userUUID, friendUUID).
		Count(&count).Error
	
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}
