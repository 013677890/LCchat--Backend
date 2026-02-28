package usecase

import (
	convsvc "github.com/013677890/LCchat-Backend/apps/msg/internal/domain/conversation"
	msgsvc "github.com/013677890/LCchat-Backend/apps/msg/internal/domain/message"
	"github.com/013677890/LCchat-Backend/apps/msg/mq"
)

// SendMessageWorkflow 发送消息用例（协调层）
// 这是整个 msg-service 最复杂的流程，跨越 message 和 conversation 两个领域。
//
// 编排步骤：
//  1. message.Service → 幂等检查（三元组去重）
//  2. message.Service → 计算 conv_id（单聊 p2p-sorted / 群聊 target_uuid）
//  3. message.Service → 分配 seq（Redis INCR）
//  4. message.Service → 生成 msg_id（ULID）+ 落库 Message
//  5. conversation.Service → Upsert 发送方会话
//  6. conversation.Service → Upsert 接收方会话（仅单聊写扩散；群聊跳过）
//  7. mq.Producer → 写 Kafka MsgPushEvent（key=conv_id）
//  8. 返回 {msg_id, seq, conv_id, send_time}
//
// 设计原则：
// - usecase 层是唯一允许协调多个 domain service 的地方
// - usecase 层是唯一允许直接调用 Kafka producer 的地方
// - domain service 之间不互相依赖
type SendMessageWorkflow struct {
	msgService  *msgsvc.Service
	convService *convsvc.Service
	producer    *mq.Producer
}

// NewSendMessageWorkflow 创建发送消息用例
func NewSendMessageWorkflow(
	msgService *msgsvc.Service,
	convService *convsvc.Service,
	producer *mq.Producer,
) *SendMessageWorkflow {
	return &SendMessageWorkflow{
		msgService:  msgService,
		convService: convService,
		producer:    producer,
	}
}

// TODO: 实现 Execute(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error)
