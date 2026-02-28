package conversation

// Service 会话领域服务
// 职责：纯会话领域逻辑
//
// 核心能力：
// - 会话列表拉取（全量/增量，游标分页）
// - 标记已读（更新 read_seq，计算剩余未读数）
// - 删除会话（逻辑删除 + 记录 clear_seq）
// - 更新会话设置（免打扰/置顶）
// - Upsert 会话（发消息时由 usecase 层调用）
//
// 设计原则：
// - 不依赖 message 领域
// - 不直接写 Kafka（MarkRead 的多端同步由 usecase 层协调）
// - 通过 Repository 接口隔离存储实现
type Service struct {
	repo Repository
}

// NewService 创建会话领域服务
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// TODO: 实现以下方法
//
// - GetConversations(ctx, req) — 按 owner_uuid + updated_since 分页查询
// - MarkRead(ctx, req) — 更新 read_seq = max(DB.read_seq, req.read_seq)，计算 unread_count
// - DeleteConversation(ctx, req) — status=1, clear_seq=max_seq, read_seq=max_seq
// - UpdateSettings(ctx, req) — optional bool 语义，只更新传入的字段
// - UpsertForMessage(ctx, ...) — 发消息时更新/创建会话（供 usecase 调用）
