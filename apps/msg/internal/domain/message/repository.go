package message

import (
	"context"

	"github.com/013677890/LCchat-Backend/model"
)

// Repository 消息领域仓储接口
// 职责：消息表 CRUD + Redis seq 分配 + 幂等缓存
type Repository interface {
	// ---- seq 分配 ----

	// AllocSeq 原子分配会话内递增序号
	// 实现：Redis INCR msg:seq:{conv_id}
	AllocSeq(ctx context.Context, convId string) (int64, error)

	// ---- 幂等 ----

	// CheckIdempotent 检查消息是否已存在（三元组去重）
	// 命中返回已存在的 Message，未命中返回 nil
	CheckIdempotent(ctx context.Context, fromUuid, deviceId, clientMsgId string) (*model.Message, error)

	// SetIdempotentCache 写入幂等缓存（Redis，10min TTL）
	SetIdempotentCache(ctx context.Context, fromUuid, deviceId, clientMsgId string, msg *model.Message) error

	// ---- 消息 CRUD ----

	// Create 插入一条消息
	Create(ctx context.Context, msg *model.Message) error

	// GetBySeqRange 按 seq 范围拉取消息（支持双向 + clear_seq 过滤）
	// direction: 1=FORWARD(seq > anchor), 2=BACKWARD(seq < anchor)
	GetBySeqRange(ctx context.Context, convId string, anchorSeq int64, direction int, limit int, clearSeq int64) ([]*model.Message, error)

	// GetByIds 批量按消息 ID 查询
	GetByIds(ctx context.Context, convId string, msgIds []string) ([]*model.Message, error)

	// GetById 查单条消息
	GetById(ctx context.Context, convId string, msgId string) (*model.Message, error)

	// ---- 撤回 ----

	// UpdateStatus 更新消息状态和内容（撤回场景：status=1, content=提示JSON）
	UpdateStatus(ctx context.Context, convId string, msgId string, status int8, content string) error
}
