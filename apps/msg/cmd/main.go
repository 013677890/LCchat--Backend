package main

// msg-service 入口
// 参照 apps/user/cmd/main.go 模式
//
// 初始化顺序：
// 1. Logger
// 2. MySQL（复用 config.DefaultMySQLConfig / pkg/mysql）
// 3. Redis（复用 config.DefaultRedisConfig / pkg/redis）
// 4. Kafka Producer（msg.push topic）
// 5. Async 协程池
// 6. 雪花算法 / ULID 生成器
// 7. Repository 层组装（MsgRepo, ConvRepo, SeqRepo）
// 8. Service 层组装（MsgService, ConvService）
// 9. Handler 层组装（MsgHandler）
// 10. Metrics HTTP Server
// 11. gRPC Server（注册 MsgServiceServer，阻塞）
//
// 环境变量：
// - MSG_GRPC_ADDR (默认 :9092)
// - MSG_METRICS_ADDR (默认 :9093)
//
// TODO: 实现完整初始化链路

func main() {
	// TODO
}
