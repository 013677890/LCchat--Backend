package handler

import (
	pb "github.com/013677890/LCchat-Backend/apps/msg/pb"

	convsvc "github.com/013677890/LCchat-Backend/apps/msg/internal/domain/conversation"
	msgsvc "github.com/013677890/LCchat-Backend/apps/msg/internal/domain/message"
	"github.com/013677890/LCchat-Backend/apps/msg/internal/usecase"
)

// MsgHandler 消息服务 gRPC Handler（薄层）
// 职责：接收 gRPC 请求 → 委托给 domain service 或 usecase workflow
//
// 路由规则：
// - 跨领域操作 → usecase workflow（SendMessage, RecallMessage, MarkRead）
// - 单领域操作 → 直接调用 domain service（PullMessages, GetMessagesByIds, GetConversations, DeleteConv, UpdateSettings）
type MsgHandler struct {
	pb.UnimplementedMsgServiceServer

	// domain services
	msgService  *msgsvc.Service
	convService *convsvc.Service

	// usecase workflows
	sendMessageWorkflow   *usecase.SendMessageWorkflow
	recallMessageWorkflow *usecase.RecallMessageWorkflow
	markReadWorkflow      *usecase.MarkReadWorkflow
}

// NewMsgHandler 创建 MsgHandler
func NewMsgHandler(
	msgService *msgsvc.Service,
	convService *convsvc.Service,
	sendWf *usecase.SendMessageWorkflow,
	recallWf *usecase.RecallMessageWorkflow,
	markReadWf *usecase.MarkReadWorkflow,
) *MsgHandler {
	return &MsgHandler{
		msgService:            msgService,
		convService:           convService,
		sendMessageWorkflow:   sendWf,
		recallMessageWorkflow: recallWf,
		markReadWorkflow:      markReadWf,
	}
}

// TODO: 实现 gRPC 方法
//
// 跨领域（走 usecase）：
// - SendMessage        → sendMessageWorkflow.Execute(ctx, req)
// - RecallMessage      → recallMessageWorkflow.Execute(ctx, req)
// - MarkRead           → markReadWorkflow.Execute(ctx, req)
//
// 单领域（直调 domain service）：
// - PullMessages       → msgService.PullMessages(ctx, req)
// - GetMessagesByIds   → msgService.GetMessagesByIds(ctx, req)
// - GetConversations   → convService.GetConversations(ctx, req)
// - DeleteConversation → convService.DeleteConversation(ctx, req)
// - UpdateConvSettings → convService.UpdateSettings(ctx, req)
