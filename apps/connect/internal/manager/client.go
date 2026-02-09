package manager

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	defaultSendQueueSize = 64
	wsWriteTimeout       = 5 * time.Second
)

// MessageHandler 定义上行消息回调。
// 参数 raw 为客户端原始二进制载荷（通常是 JSON 编码后的字节）。
type MessageHandler func(raw []byte)

// CloseHandler 定义连接关闭回调。
// 用于在 read/write 循环退出后执行清理逻辑（例如从 manager 注销）。
type CloseHandler func()

// Client 封装单条 WebSocket 连接。
// 设计要点：
// - send 队列用于削峰，避免业务 goroutine 直接阻塞在网络写；
// - done 用于统一关闭信号，读写循环都监听该信号退出；
// - once 保证 Close 幂等，避免重复 close channel/panic。
type Client struct {
	conn     *websocket.Conn
	userUUID string
	deviceID string
	send     chan []byte
	done     chan struct{}
	once     sync.Once
}

// NewClient 创建连接包装对象。
func NewClient(conn *websocket.Conn, userUUID, deviceID string) *Client {
	return &Client{
		conn:     conn,
		userUUID: userUUID,
		deviceID: deviceID,
		send:     make(chan []byte, defaultSendQueueSize),
		done:     make(chan struct{}),
	}
}

func (c *Client) UserUUID() string {
	return c.userUUID
}

func (c *Client) DeviceID() string {
	return c.deviceID
}

// Done 返回连接关闭信号通道。
// 外部可通过监听该通道感知连接生命周期结束。
func (c *Client) Done() <-chan struct{} {
	return c.done
}

// Enqueue 将待发送消息投递到写队列。
// 返回值语义：
// - true：已成功入队；
// - false：连接已关闭或队列已满（调用方可选择断开连接或丢弃消息）。
func (c *Client) Enqueue(msg []byte) bool {
	if len(msg) == 0 {
		return true
	}
	cloned := append([]byte(nil), msg...)
	select {
	case <-c.done:
		return false
	case c.send <- cloned:
		return true
	default:
		return false
	}
}

// Run 启动读写循环并阻塞等待 readLoop 结束。
// 行为说明：
// - writeLoop 在独立 goroutine 中运行；
// - readLoop 在当前 goroutine 运行，通常由其错误/断连触发整体退出；
// - 退出时保证调用 Close 和 onClose，确保资源回收。
func (c *Client) Run(ctx context.Context, onMessage MessageHandler, onClose CloseHandler) {
	defer func() {
		c.Close()
		if onClose != nil {
			onClose()
		}
	}()

	go c.writeLoop(ctx)
	c.readLoop(ctx, onMessage)
}

// Close 幂等关闭连接。
// 关闭顺序：
// 1. 关闭 done 信号，通知读写循环退出；
// 2. 关闭底层 websocket 连接释放网络资源。
func (c *Client) Close() {
	c.once.Do(func() {
		close(c.done)
		_ = c.conn.Close()
	})
}

// readLoop 持续读取客户端上行帧并交由 onMessage 处理。
// 退出条件：ctx cancel、连接关闭信号、网络读错误。
func (c *Client) readLoop(ctx context.Context, onMessage MessageHandler) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.done:
			return
		default:
		}

		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		if onMessage != nil {
			onMessage(raw)
		}
	}
}

// writeLoop 持续从 send 队列取消息写入客户端。
// 每次写操作设置超时，避免慢连接长期占用写协程。
func (c *Client) writeLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.done:
			return
		case msg := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(wsWriteTimeout))
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				c.Close()
				return
			}
		}
	}
}
