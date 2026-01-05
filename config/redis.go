package config

import "time"

// RedisConfig 单实例 Redis 配置。
// 仅需一套连接，容器场景建议仍走 stdout 观测，Redis 保持轻量连接池。
type RedisConfig struct {
	Addr         string        `json:"addr" yaml:"addr"`                 // host:port
	Password     string        `json:"password" yaml:"password"`         // 可空
	DB           int           `json:"db" yaml:"db"`                     // DB 索引，默认 0
	PoolSize     int           `json:"poolSize" yaml:"poolSize"`         // 连接池大小
	MinIdleConns int           `json:"minIdleConns" yaml:"minIdleConns"` // 最小空闲连接
	DialTimeout  time.Duration `json:"dialTimeout" yaml:"dialTimeout"`   // 建连超时
	ReadTimeout  time.Duration `json:"readTimeout" yaml:"readTimeout"`   // 读超时
	WriteTimeout time.Duration `json:"writeTimeout" yaml:"writeTimeout"` // 写超时
	PoolTimeout  time.Duration `json:"poolTimeout" yaml:"poolTimeout"`   // 从池获取连接超时
	ConnMaxIdle  time.Duration `json:"connMaxIdle" yaml:"connMaxIdle"`   // 连接最大空闲时间（对应 go-redis ConnMaxIdleTime）
}

// DefaultRedisConfig 返回本地开发的默认配置。
func DefaultRedisConfig() RedisConfig {
	return RedisConfig{
		// 与 docker-compose.yml 对齐：host redis，默认无密码
		Addr:         "redis:6379",
		Password:     "",
		DB:           0,
		PoolSize:     20,
		MinIdleConns: 4,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
		PoolTimeout:  4 * time.Second,
		ConnMaxIdle:  5 * time.Minute,
	}
}
