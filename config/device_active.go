package config

// DeviceActiveConfig 设备活跃时间更新的本地缓存配置（Gateway）
type DeviceActiveConfig struct {
	CacheSize int `json:"cacheSize" yaml:"cacheSize"` // LRU 容量
}

// DefaultDeviceActiveConfig 返回默认配置
func DefaultDeviceActiveConfig() DeviceActiveConfig {
	return DeviceActiveConfig{
		CacheSize: 100000,
	}
}
