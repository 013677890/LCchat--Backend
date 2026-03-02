package conversation

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/013677890/LCchat-Backend/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// repositoryImpl 会话领域仓储实现
type repositoryImpl struct {
	db *gorm.DB
}

// NewRepository 创建会话仓储实例
func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

// ==================== 个人会话 (conversation 表) ====================

// GetByOwnerAndConvId 查询单个会话
func (r *repositoryImpl) GetByOwnerAndConvId(ctx context.Context, ownerUuid, convId string) (*model.Conversation, error) {
	var conv model.Conversation
	err := r.db.WithContext(ctx).
		Where("owner_uuid = ? AND conv_id = ?", ownerUuid, convId).
		First(&conv).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrConversationNotFound
		}
		return nil, fmt.Errorf("GetByOwnerAndConvId: db query failed: %w", err)
	}
	return &conv, nil
}

// List 分页查询会话列表
//
// 分页方式：游标分页（基于 updated_at_id），降序排列
// hasMore 判断：N+1 Trick
func (r *repositoryImpl) List(ctx context.Context, ownerUuid string, updatedSince int64, cursor string, pageSize int) ([]*model.Conversation, bool, error) {
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 50
	}

	query := r.db.WithContext(ctx).Where("owner_uuid = ?", ownerUuid)

	if updatedSince > 0 {
		// 增量同步：返回所有变更记录（包括 status=1 已删除的，用于多端同步删除状态）
		sinceTime := time.UnixMilli(updatedSince)
		query = query.Where("updated_at > ?", sinceTime)
	} else {
		// 全量拉取：只返回活跃会话
		query = query.Where("status = 0")
	}

	// 解析复合游标
	if cursor != "" {
		var curUpdatedAtStr, curIdStr string
		parts := strings.SplitN(cursor, "_", 2)
		if len(parts) == 2 {
			curUpdatedAtStr = parts[0]
			curIdStr = parts[1]

			curUpdatedAt, err1 := strconv.ParseInt(curUpdatedAtStr, 10, 64)
			curId, err2 := strconv.ParseInt(curIdStr, 10, 64)

			if err1 == nil && err2 == nil {
				curTime := time.UnixMilli(curUpdatedAt)
				// 核心游标逻辑：严格小于上一页最后一条的 (updated_at, id)
				query = query.Where("(updated_at < ?) OR (updated_at = ? AND id < ?)", curTime, curTime, curId)
			}
		}
	}

	var convs []*model.Conversation
	err := query.Order("updated_at DESC, id DESC").
		Limit(pageSize + 1).
		Find(&convs).Error
	if err != nil {
		return nil, false, fmt.Errorf("List: db query failed: %w", err)
	}

	hasMore := len(convs) > pageSize
	if hasMore {
		convs = convs[:pageSize]
	}

	return convs, hasMore, nil
}

// Upsert 创建或更新个人会话
//
// 【Bug1 修复】
// - 只更新核心字段 (max_seq, last_msg_*, status)
// - 绝不碰 mute / pin / read_seq / clear_seq（这些由专门的方法维护）
// - 接收方 unread_count 在 DB 层面 +1，而非 Go 层覆盖
// - 发送方 read_seq = max_seq，unread_count 不变
func (r *repositoryImpl) Upsert(ctx context.Context, conv *model.Conversation, isSender bool) error {
	// 构造更新 map，只更新发消息时需要变更的字段
	updates := map[string]interface{}{
		"max_seq":          conv.MaxSeq,
		"last_msg_id":      conv.LastMsgId,
		"last_msg_at":      conv.LastMsgAt,
		"last_msg_preview": conv.LastMsgPrev,
		"status":           0,          // 重新激活已删除会话
		"updated_at":       time.Now(), // 强制更新，用于增量拉取排序
	}

	if isSender {
		// 发送方：read_seq 追平到当前消息（自己发的消息不算未读）
		updates["read_seq"] = conv.MaxSeq
	} else {
		// 接收方：在数据库层面 unread_count + 1（极端并发下也安全）
		updates["unread_count"] = gorm.Expr("unread_count + 1")
	}

	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "owner_uuid"}, {Name: "conv_id"}},
			DoUpdates: clause.Assignments(updates),
		}).Create(conv).Error
	if err != nil {
		return fmt.Errorf("Upsert: db upsert failed: %w", err)
	}
	return nil
}

// UpdateReadSeq 更新已读位点（单调递增）
//
// GREATEST 保证 read_seq 只增不减，防止旧设备覆盖新设备的已读位点
func (r *repositoryImpl) UpdateReadSeq(ctx context.Context, ownerUuid, convId string, readSeq int64) error {
	result := r.db.WithContext(ctx).
		Model(&model.Conversation{}).
		Where("owner_uuid = ? AND conv_id = ?", ownerUuid, convId).
		Updates(map[string]interface{}{
			"read_seq":     gorm.Expr("GREATEST(read_seq, ?)", readSeq),
			"unread_count": gorm.Expr("GREATEST(0, max_seq - GREATEST(read_seq, ?))", readSeq),
		})
	if result.Error != nil {
		return fmt.Errorf("UpdateReadSeq: db update failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrConversationNotFound
	}
	return nil
}

// Delete 逻辑删除会话
func (r *repositoryImpl) Delete(ctx context.Context, ownerUuid, convId string) error {
	result := r.db.WithContext(ctx).
		Model(&model.Conversation{}).
		Where("owner_uuid = ? AND conv_id = ?", ownerUuid, convId).
		Updates(map[string]interface{}{
			"status":       1,
			"clear_seq":    gorm.Expr("max_seq"),
			"read_seq":     gorm.Expr("max_seq"),
			"unread_count": 0,
		})
	if result.Error != nil {
		return fmt.Errorf("Delete: db update failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrConversationNotFound
	}
	return nil
}

// UpdateSettings 更新会话设置（optional bool 语义）
func (r *repositoryImpl) UpdateSettings(ctx context.Context, ownerUuid, convId string, mute *bool, pin *bool) error {
	updates := map[string]interface{}{}
	if mute != nil {
		updates["mute"] = *mute
	}
	if pin != nil {
		updates["pin"] = *pin
	}
	if len(updates) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&model.Conversation{}).
		Where("owner_uuid = ? AND conv_id = ?", ownerUuid, convId).
		Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("UpdateSettings: db update failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrConversationNotFound
	}
	return nil
}

// ==================== 群会话热数据 (group_conversation 表) ====================

// UpsertGroupConv 创建或更新群会话热数据
func (r *repositoryImpl) UpsertGroupConv(ctx context.Context, gc *model.GroupConversation) error {
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "group_uuid"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"max_seq", "last_msg_id", "last_msg_preview", "last_msg_at",
			}),
		}).Create(gc).Error
	if err != nil {
		return fmt.Errorf("UpsertGroupConv: db upsert failed: %w", err)
	}
	return nil
}

// GetGroupConv 查询单个群的热数据
func (r *repositoryImpl) GetGroupConv(ctx context.Context, groupUuid string) (*model.GroupConversation, error) {
	var gc model.GroupConversation
	err := r.db.WithContext(ctx).
		Where("group_uuid = ?", groupUuid).
		First(&gc).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrConversationNotFound
		}
		return nil, fmt.Errorf("GetGroupConv: db query failed: %w", err)
	}
	return &gc, nil
}

// BatchGetGroupConvs 批量查询群会话热数据
//
// 【Bug2 修复】用于 GetConversations 拼装群聊真实 max_seq / last_msg_*
// 返回 map[group_uuid]*GroupConversation，调用方按 target_uuid 匹配
func (r *repositoryImpl) BatchGetGroupConvs(ctx context.Context, groupUuids []string) (map[string]*model.GroupConversation, error) {
	if len(groupUuids) == 0 {
		return map[string]*model.GroupConversation{}, nil
	}

	var gcs []*model.GroupConversation
	err := r.db.WithContext(ctx).
		Where("group_uuid IN ?", groupUuids).
		Find(&gcs).Error
	if err != nil {
		return nil, fmt.Errorf("BatchGetGroupConvs: db query failed: %w", err)
	}

	result := make(map[string]*model.GroupConversation, len(gcs))
	for _, gc := range gcs {
		result[gc.GroupUuid] = gc
	}
	return result, nil
}
