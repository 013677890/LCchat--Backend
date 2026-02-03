package middleware

import (
	"ChatServer/pkg/deviceactive"
	pkgredis "ChatServer/pkg/redis"
	"context"
	"fmt"
	"time"

	"ChatServer/pkg/logger"
)

const deviceActiveTTL = 45 * 24 * time.Hour

func updateDeviceActive(userUUID, deviceID string) {
	if userUUID == "" || deviceID == "" {
		return
	}

	now := time.Now()
	cacheKey := userUUID + ":" + deviceID
	if !deviceactive.ShouldUpdate(cacheKey, now) {
		return
	}

	redisClient := pkgredis.Client()
	if redisClient == nil {
		return
	}

	ctx := context.Background()
	key := fmt.Sprintf("user:devices:active:%s", userUUID)
	ts := now.Unix()

	pipe := redisClient.Pipeline()
	pipe.HSet(ctx, key, deviceID, ts)
	pipe.Expire(ctx, key, deviceActiveTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		logger.Warn(ctx, "更新设备活跃时间失败",
			logger.String("user_uuid", userUUID),
			logger.String("device_id", deviceID),
			logger.ErrorField("error", err),
		)
	}
}
