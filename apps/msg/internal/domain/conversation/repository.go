package conversation

import (
	"context"

	"github.com/013677890/LCchat-Backend/model"
)

// Repository 会话领域仓储接口
// 职责：conversation 表 + group_conversation 表 CRUD
type Repository interface {
	// ==================== 个人会话 (conversation 表) ====================

	// GetByOwnerAndConvId 查询单个会话
	GetByOwnerAndConvId(ctx context.Context, ownerUuid, convId string) (*model.Conversation, error)

	// List 分页查询会话列表（支持全量/增量同步）
	//   - updatedSince=0 全量，>0 增量（只返回 updated_at > updatedSince 的记录）
	//   - cursor: 复合游标 "{updated_at}_{id}"，首页传 ""
	//   - 返回：会话列表, hasMore, error
	List(ctx context.Context, ownerUuid string, updatedSince int64, cursor string, pageSize int) ([]*model.Conversation, bool, error)

	// Upsert 创建或更新个人会话（发消息时调用）
	//   - 按 (owner_uuid, conv_id) 唯一键 upsert
	//   - isSender: 发送方不增加未读数；接收方在 DB 层面 unread_count + 1
	//   - 只更新核心字段 (max_seq, last_msg_*, status)，绝不碰 mute/pin/read_seq/clear_seq
	Upsert(ctx context.Context, conv *model.Conversation, isSender bool) error

	// UpdateReadSeq 更新已读位点（单调递增）
	//   - 实现：UPDATE SET read_seq = GREATEST(read_seq, ?),
	//           unread_count = GREATEST(0, max_seq - GREATEST(read_seq, ?))
	UpdateReadSeq(ctx context.Context, ownerUuid, convId string, readSeq int64) error

	// Delete 逻辑删除会话
	//   - 实现：status=1, clear_seq=max_seq, read_seq=max_seq, unread_count=0
	Delete(ctx context.Context, ownerUuid, convId string) error

	// UpdateSettings 更新会话设置（免打扰/置顶）
	//   - optional 语义：nil 表示不修改该字段
	UpdateSettings(ctx context.Context, ownerUuid, convId string, mute *bool, pin *bool) error

	// ==================== 群会话热数据 (group_conversation 表) ====================

	// UpsertGroupConv 创建或更新群会话热数据
	//   - 每发一条群消息就 UPDATE 一次 max_seq, last_msg_*
	UpsertGroupConv(ctx context.Context, gc *model.GroupConversation) error

	// GetGroupConv 查询单个群的热数据
	GetGroupConv(ctx context.Context, groupUuid string) (*model.GroupConversation, error)

	// BatchGetGroupConvs 批量查询群会话热数据
	//   - 用于 GetConversations 时，将群聊的真实 max_seq / last_msg_* 拼装回去
	//   - 返回 map[group_uuid]*GroupConversation
	BatchGetGroupConvs(ctx context.Context, groupUuids []string) (map[string]*model.GroupConversation, error)
}
