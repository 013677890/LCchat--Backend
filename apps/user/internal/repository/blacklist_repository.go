package repository

import (
	"ChatServer/pkg/async"
	"ChatServer/model"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// blacklistRepositoryImpl 黑名单数据访问层实现
type blacklistRepositoryImpl struct {
	db          *gorm.DB
	redisClient *redis.Client
}

// NewBlacklistRepository 创建黑名单仓储实例
func NewBlacklistRepository(db *gorm.DB, redisClient *redis.Client) IBlacklistRepository {
	return &blacklistRepositoryImpl{db: db, redisClient: redisClient}
}

// AddBlacklist 拉黑用户
func (r *blacklistRepositoryImpl) AddBlacklist(ctx context.Context, userUUID, targetUUID string) error {
	now := time.Now()
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// A -> B: 标记为拉黑（status=1/3）
		// - status=1: 原先为好友（含已删除好友）
		// - status=3: 原先非好友
		status := int8(3)
		var existing model.UserRelation
		if err := tx.Unscoped().
			Select("status").
			Where("user_uuid = ? AND peer_uuid = ?", userUUID, targetUUID).
			First(&existing).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		} else {
			switch existing.Status {
			case 0, 1, 2:
				status = 1
			case 3:
				status = 3
			default:
				status = 3
			}
		}

		relationAB := &model.UserRelation{
			UserUuid:  userUUID,
			PeerUuid:  targetUUID,
			Status:    status,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_uuid"}, {Name: "peer_uuid"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"status":     status,
				"deleted_at": nil,
				"updated_at": now,
			}),
		}).Create(relationAB).Error; err != nil {
			return err
		}

		// B -> A: 不变（保留好友关系，由消息链路查询黑名单拦截）
		return nil
	})
	if err != nil {
		return WrapDBError(err)
	}

	// 异步更新黑名单缓存（仅更新当前用户侧）
	r.updateBlacklistCacheAsync(ctx, userUUID, targetUUID)

	return nil
}

// RemoveBlacklist 取消拉黑
func (r *blacklistRepositoryImpl) RemoveBlacklist(ctx context.Context, userUUID, targetUUID string) error {
	return nil // TODO: 取消拉黑
}

// GetBlacklistList 获取黑名单列表
func (r *blacklistRepositoryImpl) GetBlacklistList(ctx context.Context, userUUID string, page, pageSize int) ([]*model.UserRelation, int64, error) {
	return nil, 0, nil // TODO: 获取黑名单列表
}

// IsBlocked 检查是否被拉黑
// 检查 userUUID 是否拉黑了 targetUUID
// 采用 Cache-Aside Pattern：优先查 Redis Set，未命中则回源 MySQL 并缓存
func (r *blacklistRepositoryImpl) IsBlocked(ctx context.Context, userUUID, targetUUID string) (bool, error) {
	cacheKey := fmt.Sprintf("user:relation:blacklist:%s", userUUID)

	// ==================== 1. 组合查询 Redis (Pipeline) ====================
	// 使用 Pipeline 一次性发送命令，减少网络 RTT
	pipe := r.redisClient.Pipeline()

	// 命令1: 检查 Key 是否存在 (区分缓存命中/未命中)
	existsCmd := pipe.Exists(ctx, cacheKey)
	// 命令2: 检查是否已拉黑 (只有 Key 存在时此结果才有效)
	isMemberCmd := pipe.SIsMember(ctx, cacheKey, targetUUID)

	// 概率续期优化：1% 的概率在读取时顺便续期
	// 无论 Key 是否存在，Expire 都是安全的 (不存在则返回0)
	if getRandomBool(0.01) {
		pipe.Expire(ctx, cacheKey, getRandomExpireTime(24*time.Hour))
	}

	_, err := pipe.Exec(ctx)

	if err != nil && err != redis.Nil {
		// Redis 挂了，记录日志，降级去查 DB
		LogRedisError(ctx, err)
	} else if err == nil {
		// Redis 正常返回
		// 核心逻辑：先看 Key 在不在
		if existsCmd.Val() > 0 {
			// Case A: 缓存命中 (Hit)
			// 此时 Redis 是权威的。SIsMember 说 false 就是 false (绝对未拉黑)。
			// 注意：哪怕 Set 里只有 "__EMPTY__"，SIsMember 也会正确返回 false。
			return isMemberCmd.Val(), nil
		}
		// Case B: 缓存未命中 (Miss) -> Exists 返回 0
		// 代码继续往下走，去查数据库
	}

	// ==================== 2. 缓存未命中，回源查询 MySQL ====================
	var relation model.UserRelation
	err = r.db.WithContext(ctx).
		Where("user_uuid = ? AND peer_uuid = ? AND status IN ? AND deleted_at IS NULL",
			userUUID, targetUUID, []int{1, 3}).
		First(&relation).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 用户没有拉黑目标，需要缓存空值
			// ==================== 3. 重建缓存（空值） ====================
			async.RunSafe(ctx, func(runCtx context.Context) {
				pipe := r.redisClient.Pipeline()
				pipe.Del(runCtx, cacheKey)
				pipe.SAdd(runCtx, cacheKey, "__EMPTY__")
				pipe.Expire(runCtx, cacheKey, 5*time.Minute)
				if _, err := pipe.Exec(runCtx); err != nil {
					LogRedisError(runCtx, err)
				}
			}, 0)

			return false, nil
		}
		// 其他数据库错误
		return false, WrapDBError(err)
	}

	// ==================== 4. 找到了拉黑记录，重建缓存 ====================
	// 异步执行写入，不需要等待结果，让接口响应更快
	async.RunSafe(ctx, func(runCtx context.Context) {
		pipe := r.redisClient.Pipeline()
		pipe.Del(runCtx, cacheKey)
		pipe.SAdd(runCtx, cacheKey, targetUUID)
		pipe.Expire(runCtx, cacheKey, getRandomExpireTime(24*time.Hour))
		if _, err := pipe.Exec(runCtx); err != nil {
			LogRedisError(runCtx, err)
		}
	}, 0)

	return true, nil
}

// GetBlacklistRelation 获取拉黑关系
func (r *blacklistRepositoryImpl) GetBlacklistRelation(ctx context.Context, userUUID, targetUUID string) (*model.UserRelation, error) {
	return nil, nil // TODO: 获取拉黑关系
}

// updateBlacklistCacheAsync 异步更新黑名单缓存（单向）
// 仅在缓存存在时做增量更新，避免过期后写入不完整 Set
func (r *blacklistRepositoryImpl) updateBlacklistCacheAsync(ctx context.Context, userUUID, targetUUID string) {
	if userUUID == "" || targetUUID == "" {
		return
	}

	cacheKey := fmt.Sprintf("user:relation:blacklist:%s", userUUID)
	async.RunSafe(ctx, func(runCtx context.Context) {
		luaScript := redis.NewScript(luaAddBlacklistIfExists)
		expireSeconds := int(getRandomExpireTime(24 * time.Hour).Seconds())
		_, err := luaScript.Run(runCtx, r.redisClient,
			[]string{cacheKey},
			targetUUID,
			expireSeconds,
		).Result()

		if err != nil && err != redis.Nil {
			if isRedisWrongType(err) {
				_ = r.redisClient.Del(runCtx, cacheKey).Err()
				return
			}
			LogRedisError(runCtx, err)
		}
	}, 0)
}

// removeFriendCacheAsync 异步删除好友缓存（单向）
// 仅在缓存存在时做增量更新，避免过期后写入不完整 Hash
func (r *blacklistRepositoryImpl) removeFriendCacheAsync(ctx context.Context, userUUID, friendUUID string) {
	if userUUID == "" || friendUUID == "" {
		return
	}

	cacheKey := fmt.Sprintf("user:relation:friend:%s", userUUID)
	async.RunSafe(ctx, func(runCtx context.Context) {
		luaScript := redis.NewScript(luaRemoveFriendMetaIfExists)
		placeholderJSON := buildFriendMetaJSON("", "", "", 0)
		expireSeconds := int(getRandomExpireTime(24 * time.Hour).Seconds())
		_, err := luaScript.Run(runCtx, r.redisClient,
			[]string{cacheKey},
			friendUUID,
			placeholderJSON,
			expireSeconds,
		).Result()

		if err != nil && err != redis.Nil {
			if isRedisWrongType(err) {
				_ = r.redisClient.Del(runCtx, cacheKey).Err()
				return
			}
			LogRedisError(runCtx, err)
		}
	}, 0)
}
