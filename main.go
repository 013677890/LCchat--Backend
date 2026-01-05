package main

import (
	"context"
	"log"
	"time"

	"ChatServer/config"
	"ChatServer/pkg/logger"
	//"ChatServer/pkg/mysql"
	"ChatServer/pkg/redis"

	"go.uber.org/zap"
)

// main 演示 zap、MySQL、Redis 的基础初始化与健康检查。
// 运行前请确保 docker-compose 中 mysql/redis 已启动：docker compose up -d
// 运行：go run main.go
func main() {
	// 日志初始化
	logCfg := config.DefaultLoggerConfig()
	lg, err := logger.Build(logCfg)
	if err != nil {
		log.Fatalf("build logger failed: %v", err)
	}
	defer lg.Sync()
	logger.ReplaceGlobal(lg)
/*
	// MySQL 初始化与健康检查（默认使用 compose 配置: root/root, db chat_server, host mysql）
	mysqlCfg := config.DefaultMySQLConfig()
	db, err := mysql.Build(mysqlCfg)
	if err != nil {
		logger.L().Fatal("mysql init failed", zap.Error(err))
	}
	mysql.ReplaceGlobal(db)
	if sqlDB, err := db.DB(); err != nil {
		logger.L().Fatal("mysql get sql db failed", zap.Error(err))
	} else if err := sqlDB.Ping(); err != nil {
		logger.L().Fatal("mysql ping failed", zap.Error(err))
	}
	logger.L().Info("mysql connected")*/

	// Redis 初始化与健康检查（默认 host redis）
	redisCfg := config.DefaultRedisConfig()
	rdb, err := redis.Build(redisCfg)
	if err != nil {
		logger.L().Fatal("redis init failed", zap.Error(err))
	}
	redis.ReplaceGlobal(rdb)
	ctx := context.Background()
	if err := rdb.Set(ctx, "healthcheck", "ok", time.Minute).Err(); err != nil {
		logger.L().Fatal("redis set failed", zap.Error(err))
	}
	val, err := rdb.Get(ctx, "healthcheck").Result()
	if err != nil {
		logger.L().Fatal("redis get failed", zap.Error(err))
	}
	logger.L().Info("redis connected", zap.String("healthcheck", val))

	// 示例日志
	logger.L().Info("service started", zap.String("MID", "1234567890"), zap.String("env", "dev"))
	logger.L().Warn("sample warning", zap.String("MID", "1234567890"), zap.String("module", "logger"))
	logger.L().Error("sample error", zap.String("MID", "1234567890"), zap.String("module", "logger"))
}
