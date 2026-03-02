package conversation

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	pb "github.com/013677890/LCchat-Backend/apps/msg/pb"
	"github.com/013677890/LCchat-Backend/model"
)

const (
	previewMaxRunes = 20
)

type lastMsgPreviewPayload struct {
	SenderUUID string `json:"sender_uuid"`
	Preview    string `json:"preview"`
}

// Service 会话领域服务
//
// 职责边界：
//   - ✅ Upsert 个人会话 + 群会话热数据（供 usecase 调用）
//   - ✅ 会话列表拉取（全量/增量，含群聊热数据拼装）
//   - ✅ 标记已读 / 删除会话 / 更新设置
//   - ❌ 不依赖 message 领域
//   - ❌ 不直接写 Kafka
type Service struct {
	repo Repository
}

// NewService 创建会话领域服务
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ==================== UpsertForMessage (供 usecase 编排调用) ====================

// UpsertForMessage 发消息时更新/创建个人会话
//
// 参数说明：
//   - ownerUuid:  会话归属用户（发送方或接收方）
//   - msg:        刚落库的消息实体
//   - convType:   会话类型（P2P / GROUP）
//   - targetUuid: 单聊为对端 UUID，群聊为群 UUID
//   - isSender:   是否为发送方（控制未读数逻辑：发送方不加未读，接收方 DB 层面 +1）
func (s *Service) UpsertForMessage(
	ctx context.Context,
	ownerUuid string,
	msg *model.Message,
	convType pb.ConvType,
	targetUuid string,
	isSender bool,
) error {
	preview := buildLastMsgPreview(msg)
	sendTime := msg.SendTime

	conv := &model.Conversation{
		ConvId:      msg.ConvId,
		Type:        int8(convType),
		OwnerUuid:   ownerUuid,
		TargetUuid:  targetUuid,
		LastMsgId:   msg.MsgId,
		LastMsgAt:   &sendTime,
		LastMsgPrev: preview,
		MaxSeq:      msg.Seq,
		Status:      0, // 重新激活已删除的会话
	}

	// 发送方初始化时 read_seq 追平
	if isSender {
		conv.ReadSeq = msg.Seq
		conv.UnreadCount = 0
	} else {
		// INSERT 场景（首次创建会话）时 unread_count=1
		// UPDATE 场景会被 repo.Upsert 里的 gorm.Expr("unread_count + 1") 覆盖
		conv.UnreadCount = 1
	}

	// isSender 透传给 repository，控制 ON DUPLICATE KEY UPDATE 中的 unread 逻辑
	return s.repo.Upsert(ctx, conv, isSender)
}

// UpsertGroupConv 发群消息时更新群会话热数据
//
// 每发一条群消息 UPDATE 一次 max_seq + last_msg_*
func (s *Service) UpsertGroupConv(ctx context.Context, msg *model.Message) error {
	preview := buildLastMsgPreview(msg)
	sendTime := msg.SendTime

	gc := &model.GroupConversation{
		GroupUuid:   msg.ConvId,
		MaxSeq:      msg.Seq,
		LastMsgId:   msg.MsgId,
		LastMsgPrev: preview,
		LastMsgAt:   &sendTime,
	}

	return s.repo.UpsertGroupConv(ctx, gc)
}

// GetByOwnerAndConvId 获取单个个人会话记录
func (s *Service) GetByOwnerAndConvId(ctx context.Context, ownerUuid, convId string) (*model.Conversation, error) {
	return s.repo.GetByOwnerAndConvId(ctx, ownerUuid, convId)
}

// ==================== GetConversations ====================

// GetConversations 查询会话列表
//
// 完整流程：
//  1. 查出个人的所有会话 (conversation 表)
//  2. 收集 Type=2 (群聊) 的 target_uuid
//  3. 批量查 group_conversation 表拿到群的真实 max_seq / last_msg_*
//  4. 内存替换：把个人会话中群聊记录的 max_seq / last_msg 替换成群热数据的真实值
//  5. 重新计算群聊未读数：real_max_seq - read_seq
//  6. 转 proto 返回
func (s *Service) GetConversations(ctx context.Context, ownerUuid string, updatedSince int64, cursor string, pageSize int) ([]*pb.ConversationItem, bool, string, error) {
	convs, hasMore, err := s.repo.List(ctx, ownerUuid, updatedSince, cursor, pageSize)
	if err != nil {
		return nil, false, "", fmt.Errorf("GetConversations: query failed: %w", err)
	}

	// ---- 收集群聊 ID，批量查群热数据 ----
	var groupIds []string
	for _, c := range convs {
		if c.Type == 2 { // GROUP
			groupIds = append(groupIds, c.TargetUuid)
		}
	}

	var groupMap map[string]*model.GroupConversation
	if len(groupIds) > 0 {
		groupMap, err = s.repo.BatchGetGroupConvs(ctx, groupIds)
		if err != nil {
			// 群热数据查询失败不阻断，降级使用个人会话里的（可能过时的）数据
			groupMap = map[string]*model.GroupConversation{}
		}
	}

	// ---- 转换 + 拼装群热数据 ----
	items := make([]*pb.ConversationItem, 0, len(convs))
	if len(convs) == 0 {
		// 结果为空时返回 updated_since，便于客户端断点续传
		return items, hasMore, strconv.FormatInt(updatedSince, 10), nil
	}

	var nextCursorStr string
	for _, conv := range convs {
		// 如果是群聊，用群热数据替换 max_seq / last_msg_*，并重新计算未读数
		if conv.Type == 2 && groupMap != nil {
			if gc, ok := groupMap[conv.TargetUuid]; ok {
				conv.MaxSeq = gc.MaxSeq
				conv.LastMsgId = gc.LastMsgId
				conv.LastMsgPrev = gc.LastMsgPrev
				conv.LastMsgAt = gc.LastMsgAt
				// 动态计算未读数 = 群真实 max_seq - 个人 read_seq
				unread := int(gc.MaxSeq - conv.ReadSeq)
				if unread < 0 {
					unread = 0
				}
				conv.UnreadCount = unread
			}
		}

		items = append(items, modelToConvItem(conv))
		// 最后一条的 updated_at 和 id 组合成联合游标，防止毫秒级时间冲突导致丢数据
		nextCursorStr = fmt.Sprintf("%d_%d", conv.UpdatedAt.UnixMilli(), conv.Id)
	}

	return items, hasMore, nextCursorStr, nil
}

// ==================== MarkRead ====================

// MarkRead 标记会话已读
//
// DB 层面 read_seq = GREATEST(read_seq, readSeq)，同步计算 unread_count
// 返回最新计算得到的 unread_count
func (s *Service) MarkRead(ctx context.Context, ownerUuid, convId string, readSeq int64) (int32, error) {
	err := s.repo.UpdateReadSeq(ctx, ownerUuid, convId, readSeq)
	if err != nil {
		return 0, err
	}
	// 查询最新的会话状态获取 unread_count
	conv, err := s.repo.GetByOwnerAndConvId(ctx, ownerUuid, convId)
	if err != nil {
		return 0, err
	}
	return int32(conv.UnreadCount), nil
}

// ==================== DeleteConversation ====================

// DeleteConversation 逻辑删除会话
//
// status=1 + clear_seq=max_seq + read_seq=max_seq + unread=0
// 收到新消息时 Upsert 自动 status=0 重新激活
func (s *Service) DeleteConversation(ctx context.Context, ownerUuid, convId string) error {
	return s.repo.Delete(ctx, ownerUuid, convId)
}

// ==================== UpdateSettings ====================

// UpdateSettings 更新会话设置（免打扰/置顶）
func (s *Service) UpdateSettings(ctx context.Context, ownerUuid, convId string, mute *bool, pin *bool) error {
	return s.repo.UpdateSettings(ctx, ownerUuid, convId, mute, pin)
}

// ==================== 辅助方法 ====================

// truncatePreview 截取消息预览（超过 maxLen 截断 + "..."）。
func truncatePreview(text string, maxLen int) string {
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	return string(runes[:maxLen]) + "..."
}

func buildLastMsgPreview(msg *model.Message) string {
	payload := lastMsgPreviewPayload{
		SenderUUID: msg.FromUuid,
		Preview:    buildPreviewText(msg),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return `{"sender_uuid":"","preview":""}`
	}
	return string(data)
}

func buildPreviewText(msg *model.Message) string {
	switch msg.MsgType {
	case 2:
		return "[图片]"
	case 3:
		return "[语音]"
	case 4:
		return "[视频]"
	case 5:
		return "[文件]"
	case 6:
		return "[位置]"
	}

	type textContent struct {
		Text string `json:"text"`
	}
	var content textContent
	if err := json.Unmarshal([]byte(msg.Content), &content); err == nil && content.Text != "" {
		return truncatePreview(content.Text, previewMaxRunes)
	}
	return truncatePreview(msg.Content, previewMaxRunes)
}

// modelToConvItem 将 model.Conversation 转换为 pb.ConversationItem
func modelToConvItem(conv *model.Conversation) *pb.ConversationItem {
	item := &pb.ConversationItem{
		ConvId:      conv.ConvId,
		ConvType:    pb.ConvType(conv.Type),
		TargetUuid:  conv.TargetUuid,
		UnreadCount: int32(conv.UnreadCount),
		Mute:        conv.Mute,
		Pin:         conv.Pin,
		UpdatedAt:   conv.UpdatedAt.UnixMilli(),
	}

	if conv.LastMsgId != "" {
		var sendTimeMs int64
		if conv.LastMsgAt != nil {
			sendTimeMs = conv.LastMsgAt.UnixMilli()
		}
		item.LastMsg = &pb.LastMsgPreview{
			MsgId:       conv.LastMsgId,
			PreviewJson: conv.LastMsgPrev,
			SendTime:    sendTimeMs,
		}
	}

	return item
}

// ComputeUnreadCount 计算未读数 (导出供 usecase 复用)
func ComputeUnreadCount(maxSeq, readSeq int64) int {
	count := int(maxSeq - readSeq)
	if count < 0 {
		return 0
	}
	return count
}

// NowPtr 返回当前时间的指针
func NowPtr() *time.Time {
	now := time.Now()
	return &now
}
