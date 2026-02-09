package manager

import (
	"hash/fnv"
	"sync"
	"sync/atomic"
)

const (
	// defaultConnectionBuckets 连接管理默认分桶数量。
	// 默认 32 桶可在多数场景下显著降低单把大锁竞争。
	defaultConnectionBuckets = 32
)

type userBucket struct {
	mu     sync.RWMutex
	byUser map[string]map[string]*Client
}

// ConnectionManager 管理所有在线 WebSocket 连接。
// 维护按用户分桶索引：
// - byUser(user_uuid -> device_id -> client) 用于设备定位与按用户广播。
type ConnectionManager struct {
	userBuckets []userBucket
	shutdown    atomic.Bool
}

// NewConnectionManager 创建连接管理器实例。
// 默认按 32 桶初始化。
func NewConnectionManager() *ConnectionManager {
	return NewConnectionManagerWithBuckets(defaultConnectionBuckets)
}

// NewConnectionManagerWithBuckets 创建指定分桶数的连接管理器。
// bucketCount <= 0 时回退到默认值 32。
func NewConnectionManagerWithBuckets(bucketCount int) *ConnectionManager {
	if bucketCount <= 0 {
		bucketCount = defaultConnectionBuckets
	}

	m := &ConnectionManager{
		userBuckets: make([]userBucket, bucketCount),
	}

	for i := 0; i < bucketCount; i++ {
		m.userBuckets[i] = userBucket{
			byUser: make(map[string]map[string]*Client),
		}
	}

	return m
}

// Register 注册一个设备连接。
// 返回值 replaced 表示被新连接替换掉的旧连接（如果存在）。
// 调用方通常应主动关闭 replaced，确保同设备最多一个活跃连接。
func (m *ConnectionManager) Register(client *Client) (replaced *Client) {
	if client == nil || m.shutdown.Load() {
		return nil
	}

	userUUID := client.UserUUID()
	deviceID := client.DeviceID()

	userBucket := m.userBucketFor(userUUID)

	userBucket.mu.Lock()
	defer userBucket.mu.Unlock()

	// 加锁后再次判断，避免 Shutdown 与 Register 并发交错。
	if m.shutdown.Load() {
		return nil
	}

	userConns, ok := userBucket.byUser[userUUID]
	if !ok {
		userConns = make(map[string]*Client)
		userBucket.byUser[userUUID] = userConns
	}
	if old, ok := userConns[deviceID]; ok && old != client {
		replaced = old
	}
	userConns[deviceID] = client

	return replaced
}

// Unregister 注销一个连接。
// 只有当 map 中当前连接与入参完全一致时才删除，防止并发替换时误删新连接。
func (m *ConnectionManager) Unregister(client *Client) {
	if client == nil {
		return
	}

	userUUID := client.UserUUID()
	deviceID := client.DeviceID()

	userBucket := m.userBucketFor(userUUID)

	userBucket.mu.Lock()
	defer userBucket.mu.Unlock()

	if userConns, ok := userBucket.byUser[userUUID]; ok {
		// 防御并发替换：仅当指针一致时才删除，避免误删新连接。
		if existed, ok := userConns[deviceID]; ok && existed == client {
			delete(userConns, deviceID)
		}
		if len(userConns) == 0 {
			delete(userBucket.byUser, userUUID)
		}
	}
}

// SendToDevice 向指定用户的指定设备发送消息。
// 返回 false 表示目标连接不存在或写队列不可用。
func (m *ConnectionManager) SendToDevice(userUUID, deviceID string, msg []byte) bool {
	userBucket := m.userBucketFor(userUUID)

	userBucket.mu.RLock()
	var client *Client
	if userConns, ok := userBucket.byUser[userUUID]; ok {
		client = userConns[deviceID]
	}
	userBucket.mu.RUnlock()
	if client == nil {
		return false
	}
	return client.Enqueue(msg)
}

// SendToUser 向用户的所有在线设备广播消息。
// 返回成功入队的设备数量，可用于统计下行投递率。
func (m *ConnectionManager) SendToUser(userUUID string, msg []byte) int {
	userBucket := m.userBucketFor(userUUID)

	userBucket.mu.RLock()
	userConns, ok := userBucket.byUser[userUUID]
	if !ok || len(userConns) == 0 {
		userBucket.mu.RUnlock()
		return 0
	}
	clients := make([]*Client, 0, len(userConns))
	for _, client := range userConns {
		clients = append(clients, client)
	}
	userBucket.mu.RUnlock()

	sent := 0
	for _, client := range clients {
		if client.Enqueue(msg) {
			sent++
		}
	}
	return sent
}

// Count 返回当前在线连接数（按 user_uuid+device_id 去重后）。
func (m *ConnectionManager) Count() int {
	total := 0
	for i := range m.userBuckets {
		b := &m.userBuckets[i]
		b.mu.RLock()
		for _, userConns := range b.byUser {
			total += len(userConns)
		}
		b.mu.RUnlock()
	}
	return total
}

// Shutdown 关闭全部连接并阻止后续注册。
// 用于进程优雅退出阶段，确保不再接收新连接并尽快释放资源。
func (m *ConnectionManager) Shutdown() {
	if !m.shutdown.CompareAndSwap(false, true) {
		return
	}

	clients := make([]*Client, 0)
	for i := range m.userBuckets {
		b := &m.userBuckets[i]
		b.mu.Lock()
		for _, userConns := range b.byUser {
			for _, client := range userConns {
				clients = append(clients, client)
			}
		}
		b.byUser = make(map[string]map[string]*Client)
		b.mu.Unlock()
	}

	for _, client := range clients {
		client.Close()
	}
}

func (m *ConnectionManager) userBucketFor(userUUID string) *userBucket {
	return &m.userBuckets[m.bucketIndex(userUUID)]
}

func (m *ConnectionManager) bucketIndex(value string) int {
	if len(m.userBuckets) == 1 {
		return 0
	}
	return int(hashString(value) % uint32(len(m.userBuckets)))
}

func hashString(value string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(value))
	return h.Sum32()
}
