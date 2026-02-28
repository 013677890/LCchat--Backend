package usecase

import (
	msgsvc "github.com/013677890/LCchat-Backend/apps/msg/internal/domain/message"
	"github.com/013677890/LCchat-Backend/apps/msg/mq"
)

// RecallMessageWorkflow 撤回消息用例（协调层）
//
// 编排步骤：
//  1. message.Service → 查消息、校验权限（发送者 or 群管理员）、校验 2 分钟窗口
//  2. message.Service → 更新 DB（status=1, content=撤回提示 JSON）
//  3. mq.Producer → 写 Kafka MsgPushEvent{type="MSG_RECALL", data=RecallNotice}
//
// 之所以放在 usecase 而不是 message.Service：
// - 撤回需要写 Kafka 通知（跨领域 side effect）
type RecallMessageWorkflow struct {
	msgService *msgsvc.Service
	producer   *mq.Producer
}

// NewRecallMessageWorkflow 创建撤回消息用例
func NewRecallMessageWorkflow(
	msgService *msgsvc.Service,
	producer *mq.Producer,
) *RecallMessageWorkflow {
	return &RecallMessageWorkflow{
		msgService: msgService,
		producer:   producer,
	}
}

// TODO: 实现 Execute(ctx context.Context, req *pb.RecallMessageRequest) (*pb.RecallMessageResponse, error)
