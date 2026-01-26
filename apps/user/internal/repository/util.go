package repository

import (
	"math/rand"
	"time"
)

// getRandomExpireTime 生成带随机抖动的过期时间
// baseExpire: 基础过期时间
// 返回: 基础过期时间 ± 10% 的随机时间
func getRandomExpireTime(baseExpire time.Duration) time.Duration {
	// 计算随机抖动范围（±10%）
	jitterRange := float64(baseExpire) * 0.1
	jitter := time.Duration(rand.Float64()*float64(jitterRange)*2 - float64(jitterRange))

	return baseExpire + jitter
}

// getRandomBool 生成随机布尔值
// probability: 概率
// 返回: 概率为probability的布尔值
func getRandomBool(probability float64) bool {
	return rand.Float64() < probability
}