package repository

import (
	"ChatServer/model"
	"context"

	"gorm.io/gorm"
)

// userRepositoryImpl 用户数据访问层实现
// 职责：只负责GORM的CRUD操作，不含业务逻辑
// 设计原则：
//   - 返回数据库原始错误（如gorm.ErrRecordNotFound）
//   - 不进行业务判断（如密码校验）
//   - 不进行错误转换（错误转换在Service层完成）
type userRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储实例
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepositoryImpl{db: db}
}

// GetByPhone 根据手机号查询用户信息
func (r *userRepositoryImpl) GetByPhone(ctx context.Context, telephone string) (*model.UserInfo, error) {
	var user model.UserInfo
	err := r.db.WithContext(ctx).Where("telephone = ?", telephone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUUID 根据UUID查询用户信息
func (r *userRepositoryImpl) GetByUUID(ctx context.Context, uuid string) (*model.UserInfo, error) {
	var user model.UserInfo
	err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create 创建新用户
func (r *userRepositoryImpl) Create(ctx context.Context, user *model.UserInfo) (*model.UserInfo, error) {
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Update 更新用户信息
func (r *userRepositoryImpl) Update(ctx context.Context, user *model.UserInfo) (*model.UserInfo, error) {
	err := r.db.WithContext(ctx).Save(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

// ExistsByPhone 检查手机号是否已存在
func (r *userRepositoryImpl) ExistsByPhone(ctx context.Context, telephone string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.UserInfo{}).Where("telephone = ?", telephone).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateLastLogin 更新最后登录时间
func (r *userRepositoryImpl) UpdateLastLogin(ctx context.Context, userUUID string) error {
	return r.db.WithContext(ctx).Model(&model.UserInfo{}).
		Where("uuid = ?", userUUID).
		Update("updated_at", gorm.Expr("NOW()")).Error
}

// BatchGetByUUIDs 批量查询用户信息
func (r *userRepositoryImpl) BatchGetByUUIDs(ctx context.Context, uuids []string) ([]*model.UserInfo, error) {
	if len(uuids) == 0 {
		return []*model.UserInfo{}, nil
	}
	
	var users []*model.UserInfo
	err := r.db.WithContext(ctx).Where("uuid IN ?", uuids).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}
