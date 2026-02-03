package deviceactive

import (
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

const (
	// DefaultCacheSize 默认 LRU 容量
	DefaultCacheSize = 100000
	// DefaultUpdateInterval 默认更新间隔（超过该时间才写 Redis）
	DefaultUpdateInterval = 10 * time.Minute
)

var cache *lru.Cache[string, int64]

// Init 初始化设备活跃时间的本地 LRU 缓存
func Init(cacheSize int) error {
	if cacheSize <= 0 {
		cacheSize = DefaultCacheSize
	}
	c, err := lru.New[string, int64](cacheSize)
	if err != nil {
		return err
	}
	cache = c
	return nil
}

// ShouldUpdate 判断是否需要更新 Redis，并在需要时刷新本地缓存时间戳
func ShouldUpdate(key string, now time.Time) bool {
	if cache == nil {
		return true
	}
	last, ok := cache.Get(key)
	if !ok || now.Sub(time.Unix(last, 0)) >= DefaultUpdateInterval {
		cache.Add(key, now.Unix())
		return true
	}
	return false
}
