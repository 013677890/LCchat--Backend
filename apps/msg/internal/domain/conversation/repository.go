package conversation

import (
	"context"

	"github.com/013677890/LCchat-Backend/model"
)

// Repository 会话领域仓储接口
// 职责：conversation 表 CRUD
type Repository interface {
	// ---- 查询 ----

	// GetByOwnerAndConvId 查询单个会话
	GetByOwnerAndConvId(ctx context.Context, ownerUuid, convId string) (*model.Conversation, error)

	// List 分页查询会话列表（支持全量/增量同步）
	// updatedSince=0 全量，>0 增量
	// 返回：会话列表, hasMore, error
	List(ctx context.Context, ownerUuid string, updatedSince int64, cursor int64, pageSize int) ([]*model.Conversation, bool, error)

	// ---- 写入 ----

	// Upsert 创建或更新会话（发消息时调用）
	// 按 (owner_uuid, conv_id) 唯一键 upsert
	Upsert(ctx context.Context, conv *model.Conversation) error

	// UpdateReadSeq 更新已读位点（单调递增）
	// 实现：UPDATE SET read_seq = GREATEST(read_seq, ?) WHERE owner_uuid=? AND conv_id=?
	UpdateReadSeq(ctx context.Context, ownerUuid, convId string, readSeq int64) error

	// Delete 逻辑删除会话
	// 实现：status=1, clear_seq=max_seq, read_seq=max_seq
	Delete(ctx context.Context, ownerUuid, convId string) error

	// UpdateSettings 更新会话设置（免打扰/置顶）
	// optional 语义：nil 表示不修改
	UpdateSettings(ctx context.Context, ownerUuid, convId string, mute *bool, pin *bool) error
}
