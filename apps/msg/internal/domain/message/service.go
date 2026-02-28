package message

// Service 消息领域服务
// 职责：纯消息领域逻辑，不涉及会话更新和 Kafka 投递
//
// 核心能力：
// - 幂等检查 + 分配 seq + 落库（供 usecase 编排调用）
// - 按 seq/ID 拉取消息
// - 撤回（权限校验 + 时间窗口 + 状态更新）
//
// 设计原则：
// - 不依赖 conversation 领域
// - 不直接写 Kafka（由 usecase 层协调）
// - 通过 Repository 接口隔离存储实现
type Service struct {
	repo Repository
}

// NewService 创建消息领域服务
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// TODO: 实现以下方法
//
// - CreateMessage(ctx, req) — 幂等检查 + 生成 msg_id + 计算 conv_id + 分配 seq + 落库
// - PullMessages(ctx, req)  — 按 conv_id + anchor_seq + direction 查询，过滤 clear_seq
// - GetMessagesByIds(ctx, req) — 按 conv_id + msg_ids 批量查
// - RecallMessage(ctx, req) — 权限校验 + 2分钟窗口校验 + 更新 status/content
