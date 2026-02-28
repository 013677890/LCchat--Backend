package mq

import (
	"github.com/013677890/LCchat-Backend/pkg/kafka"
)

// Producer 封装 msg-service 向 Kafka msg.push topic 写入 MsgPushEvent 的逻辑
//
// Kafka 分区策略：
// - Partition Key = conv_id（保证同一会话的消息在同一 Partition 内有序消费）
// - Value = MsgPushEvent 序列化后的 Protobuf bytes
type Producer struct {
	kafkaProducer *kafka.Producer
	topic         string
}

// NewProducer 创建 msg.push Kafka 生产者
func NewProducer(kafkaProducer *kafka.Producer, topic string) *Producer {
	return &Producer{
		kafkaProducer: kafkaProducer,
		topic:         topic,
	}
}

// TODO: 实现以下方法
//
// - PublishMsgPushEvent(ctx context.Context, convId string, event *pb.MsgPushEvent) error
//   - convId 作为 Kafka partition key
//   - event 序列化为 Protobuf bytes 后写入 topic
