package repository

import (
	"ChatServer/model"
	"ChatServer/pkg/async"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// friendRepositoryImpl 好友关系数据访问层实现
type friendRepositoryImpl struct {
	db          *gorm.DB
	redisClient *redis.Client
}

// NewFriendRepository 创建好友关系仓储实例
func NewFriendRepository(db *gorm.DB, redisClient *redis.Client) IFriendRepository {
	return &friendRepositoryImpl{db: db, redisClient: redisClient}
}

// GetFriendList 获取好友列表
func (r *friendRepositoryImpl) GetFriendList(ctx context.Context, userUUID, groupTag string, page, pageSize int) ([]*model.UserRelation, int64, int64, error) {
	// 兜底分页参数
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// 基础条件：仅好友关系 + 指定用户 + 未删除
	query := r.db.WithContext(ctx).
		Model(&model.UserRelation{}).
		Where("user_uuid = ? AND status = ? AND deleted_at IS NULL", userUUID, 0)
	if groupTag != "" {
		query = query.Where("group_tag = ?", groupTag)
	}

	var total int64
	var version int64

	// 只在第一页计算 total 和 version，减少数据库开销
	if page == 1 {
		// 先查总数
		if err := query.Count(&total).Error; err != nil {
			return nil, 0, 0, WrapDBError(err)
		}

		// 全量初始化版本号取当前服务器时间
		version = time.Now().UnixMilli()
	}

	// 再查列表，按创建时间倒序（加二级排序保证稳定性）
	var relations []*model.UserRelation
	if err := query.
		Order("created_at DESC, id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&relations).
		Error; err != nil {
		return nil, 0, 0, WrapDBError(err)
	}

	return relations, total, version, nil
}

// GetFriendRelation 获取好友关系
func (r *friendRepositoryImpl) GetFriendRelation(ctx context.Context, userUUID, friendUUID string) (*model.UserRelation, error) {
	return nil, nil // TODO: 实现获取好友关系
}

// CreateFriendRelation 创建好友关系（双向）
// 使用 Upsert (INSERT ON DUPLICATE KEY UPDATE) 策略：
//   - 原子性：不存在"查不到然后插入报错"的时间差
//   - 性能：2 条 SELECT + 2 条 INSERT 变成 1 条 INSERT
//   - 稳健：正确处理软删除记录恢复场景
func (r *friendRepositoryImpl) CreateFriendRelation(ctx context.Context, userUUID, friendUUID string) error {
	now := time.Now()

	// 1. 构建双向关系
	relations := []*model.UserRelation{
		{
			UserUuid:  userUUID,
			PeerUuid:  friendUUID,
			Status:    0, // 正常状态
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			UserUuid:  friendUUID,
			PeerUuid:  userUUID,
			Status:    0, // 正常状态
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// 2. 批量 Upsert (Insert On Duplicate Key Update)
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		// 指定冲突列（必须是数据库的唯一索引列）
		Columns: []clause.Column{{Name: "user_uuid"}, {Name: "peer_uuid"}},
		// 冲突时执行更新操作
		DoUpdates: clause.Assignments(map[string]interface{}{
			"status":     0,   // 恢复正常状态
			"deleted_at": nil, // 【关键】恢复软删除
			"updated_at": now, // 更新时间
		}),
	}).Create(&relations).Error

	if err != nil {
		return WrapDBError(err)
	}

	// 3. 异步更新 Redis 好友列表缓存（合并为一个调用减少协程开销）
	r.invalidateFriendCacheAsync(ctx, userUUID, friendUUID)

	return nil
}

// DeleteFriendRelation 删除好友关系（单向）
func (r *friendRepositoryImpl) DeleteFriendRelation(ctx context.Context, userUUID, friendUUID string) error {
	return nil // TODO: 实现删除好友关系
}

// SetFriendRemark 设置好友备注
func (r *friendRepositoryImpl) SetFriendRemark(ctx context.Context, userUUID, friendUUID, remark string) error {
	return nil // TODO: 设置好友备注
}

// SetFriendTag 设置好友标签
func (r *friendRepositoryImpl) SetFriendTag(ctx context.Context, userUUID, friendUUID, groupTag string) error {
	return nil // TODO: 设置好友标签
}

// GetTagList 获取标签列表
func (r *friendRepositoryImpl) GetTagList(ctx context.Context, userUUID string) ([]string, error) {
	return nil, nil // TODO: 获取标签列表
}

// IsFriend 检查是否是好友
// 采用 Cache-Aside Pattern：优先查 Redis Set，未命中则回源 MySQL 并缓存
func (r *friendRepositoryImpl) IsFriend(ctx context.Context, userUUID, friendUUID string) (bool, error) {
	cacheKey := fmt.Sprintf("user:relation:friend:%s", userUUID)

	// ==================== 1. 组合查询 Redis (Pipeline) ====================
	// 使用 Pipeline 一次性发送命令，减少网络 RTT
	pipe := r.redisClient.Pipeline()

	// 命令1: 检查 Key 是否存在 (区分缓存命中/未命中)
	existsCmd := pipe.Exists(ctx, cacheKey)
	// 命令2: 检查是否是好友 (只有 Key 存在时此结果才有效)
	isMemberCmd := pipe.SIsMember(ctx, cacheKey, friendUUID)

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
			// 此时 Redis 是权威的。SIsMember 说 false 就是 false (绝对非好友)。
			// 注意：哪怕 Set 里只有 "__EMPTY__"，SIsMember 也会正确返回 false。
			return isMemberCmd.Val(), nil
		}
		// Case B: 缓存未命中 (Miss) -> Exists 返回 0
		// 代码继续往下走，去查数据库
	}

	// ==================== 2. 缓存未命中，回源查询 MySQL ====================
	var relations []model.UserRelation
	err = r.db.WithContext(ctx).
		Where("user_uuid = ? AND status = ? AND deleted_at IS NULL", userUUID, 0).
		Find(&relations).Error

	if err != nil {
		return false, WrapDBError(err)
	}

	// ==================== 3. 重建缓存 (保持 Set 类型) ====================
	if len(relations) == 0 {
		// [修复类型冲突] 空列表也用 Set，写入特殊标记
		async.RunSafe(ctx, func(runCtx context.Context) {
			pipe := r.redisClient.Pipeline()
			pipe.Del(runCtx, cacheKey)
			pipe.SAdd(runCtx, cacheKey, "__EMPTY__")
			pipe.Expire(runCtx, cacheKey, 5*time.Minute)
			if _, err := pipe.Exec(runCtx); err != nil {
				LogRedisError(runCtx, err)
			}
		}, 0)
	} else {
		// 提取 UUID
		friendUUIDs := make([]interface{}, len(relations))
		// 优化：顺便在内存里判断一下结果，省得最后再遍历
		isFriendFound := false
		for i, relation := range relations {
			friendUUIDs[i] = relation.PeerUuid
			if relation.PeerUuid == friendUUID {
				isFriendFound = true
			}
		}

		async.RunSafe(ctx, func(runCtx context.Context) {
			pipe := r.redisClient.Pipeline()
			pipe.Del(runCtx, cacheKey)
			pipe.SAdd(runCtx, cacheKey, friendUUIDs...)
			pipe.Expire(runCtx, cacheKey, getRandomExpireTime(24*time.Hour))
			if _, err := pipe.Exec(runCtx); err != nil {
				LogRedisError(runCtx, err)
			}
		}, 0)

		return isFriendFound, nil
	}

	// 如果是空列表，那肯定不是好友
	return false, nil
}

// GetRelationStatus 获取关系状态
func (r *friendRepositoryImpl) GetRelationStatus(ctx context.Context, userUUID, peerUUID string) (*model.UserRelation, error) {
	return nil, nil // TODO: 获取关系状态
}

// SyncFriendList 增量同步好友列表
// 返回值: 变更列表, nextVersion(客户端下次用的时间戳), hasMore(是否还有更多), error
func (r *friendRepositoryImpl) SyncFriendList(ctx context.Context, userUUID string, version int64, limit int) ([]*model.UserRelation, int64, bool, error) {
    // 1. 准备查询
    // 客户端传来的 version 是毫秒，转成 time.Time
    lastTime := time.UnixMilli(version)
    
    var relations []*model.UserRelation

    // 2. 执行查询 (极致精简)
    // 核心假设：GORM 软删除时会自动更新 updated_at。如果你的 GORM 配置没关这个，这就没问题。
    err := r.db.WithContext(ctx).
        Unscoped(). // 必须查出已删除的
        Model(&model.UserRelation{}).
        Where("user_uuid = ?", userUUID).
        Where("updated_at > ?", lastTime). // 核心：只看 update 时间，利用索引
        Order("updated_at ASC").           // 核心：利用索引排序，千万别用函数
        Limit(limit + 1).                  // 多查一条，用于判断 hasMore
        Find(&relations).Error

    if err != nil {
        return nil, 0, false, WrapDBError(err)
    }

    // 3. 计算 hasMore 和 nextVersion
    hasMore := false
    var nextVersion int64

    if len(relations) > limit {
        hasMore = true
        relations = relations[:limit] // 去掉多查的那一条
        // 情况 A：还有更多数据，Cursor 必须是本批次最后一条的时间
        nextVersion = relations[len(relations)-1].UpdatedAt.UnixMilli()
    } else {
        hasMore = false
        // 情况 B：没有更多数据了（追平了）
        // 这里的 nextVersion 可以是最后一条的时间，也可以是 ServerTime
        // 推荐：取 ServerTime 并回退 5 秒（安全窗口），防止事务并发导致的数据丢失
        safeTime := time.Now().Add(-5 * time.Second).UnixMilli()
        
        // 如果列表为空，直接用 safeTime；如果不为空，取 max(lastItem, safeTime)
        if len(relations) > 0 {
            lastItemTime := relations[len(relations)-1].UpdatedAt.UnixMilli()
            if lastItemTime > safeTime {
                nextVersion = lastItemTime
            } else {
                nextVersion = safeTime
            }
        } else {
            // 如果本来就没数据，说明 version 已经很新了，保持原样或推进到 safeTime
            if safeTime > version {
                nextVersion = safeTime
            } else {
                nextVersion = version
            }
        }
    }

    return relations, nextVersion, hasMore, nil
}

// BatchCheckIsFriend 批量检查是否为好友（使用Redis Set优化）
// 返回：map[peerUUID]isFriend
func (r *friendRepositoryImpl) BatchCheckIsFriend(ctx context.Context, userUUID string, peerUUIDs []string) (map[string]bool, error) {
	if len(peerUUIDs) == 0 {
		return make(map[string]bool), nil
	}

	// 构建 Redis Set key
	cacheKey := fmt.Sprintf("user:relation:friend:%s", userUUID)

	// ==================== 1. 组合查询 Redis (Pipeline) ====================
	// 优化：使用多个 SIsMember 而不是 SMembers
	// 好处：用户有 2000 好友，只查 2 人时，网络传输从 2000 个 UUID → 2 个 bool
	pipe := r.redisClient.Pipeline()

	// 命令1: 检查 Key 是否存在 (区分缓存命中/未命中)
	existsCmd := pipe.Exists(ctx, cacheKey)

	// 命令2: 批量检查每个 peerUUID 是否是好友
	isMemberCmds := make([]*redis.BoolCmd, len(peerUUIDs))
	for i, peerUUID := range peerUUIDs {
		isMemberCmds[i] = pipe.SIsMember(ctx, cacheKey, peerUUID)
	}

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
			// 此时 Redis 是权威的，直接返回结果
			result := make(map[string]bool, len(peerUUIDs))
			for i, peerUUID := range peerUUIDs {
				// 如果 SIsMember 出错，保守返回 false（后续会降级查 DB）
				if isMemberCmds[i].Err() != nil {
					LogRedisError(ctx, isMemberCmds[i].Err())
					result[peerUUID] = false
				} else {
					result[peerUUID] = isMemberCmds[i].Val()
				}
			}
			return result, nil
		}
		// Case B: 缓存未命中 (Miss) -> Exists 返回 0
		// 代码继续往下走，去查数据库
	}

	// ==================== 2. 缓存未命中，回源查询 MySQL ====================
	var relations []model.UserRelation
	err = r.db.WithContext(ctx).
		Where("user_uuid = ? AND status = ? AND deleted_at IS NULL", userUUID, 0).
		Find(&relations).Error

	if err != nil {
		return nil, WrapDBError(err)
	}

	// ==================== 3. 统一重建缓存 (保持 Set 类型) ====================
	// 优化：合并空列表和非空列表的 Pipeline 逻辑，避免代码重复
	async.RunSafe(ctx, func(runCtx context.Context) {
		pipe := r.redisClient.Pipeline()
		pipe.Del(runCtx, cacheKey)
		if len(relations) == 0 {
			pipe.SAdd(runCtx, cacheKey, "__EMPTY__")
			pipe.Expire(runCtx, cacheKey, 5*time.Minute)
		} else {
			friendUUIDs := make([]interface{}, len(relations))
			for i, relation := range relations {
				friendUUIDs[i] = relation.PeerUuid
			}
			pipe.SAdd(runCtx, cacheKey, friendUUIDs...)
			pipe.Expire(runCtx, cacheKey, getRandomExpireTime(24*time.Hour))
		}
		if _, err := pipe.Exec(runCtx); err != nil {
			LogRedisError(runCtx, err)
		}
	}, 0)

	// ==================== 4. 构建返回结果 ====================
	// 将 DB 查询到的好友集合转为 map
	friendSet := make(map[string]bool, len(relations))
	for _, relation := range relations {
		friendSet[relation.PeerUuid] = true
	}

	// 构建返回结果
	result := make(map[string]bool, len(peerUUIDs))
	for _, peerUUID := range peerUUIDs {
		result[peerUUID] = friendSet[peerUUID]
	}

	return result, nil
}

// invalidateFriendCacheAsync 异步更新双方的好友缓存
// 在单个协程中同时处理 userUUID 和 friendUUID 的缓存更新
func (r *friendRepositoryImpl) invalidateFriendCacheAsync(ctx context.Context, userUUID, friendUUID string) {
	async.RunSafe(ctx, func(runCtx context.Context) {
		// 处理两个用户的缓存
		pairs := []struct{ userKey, newFriend string }{
			{fmt.Sprintf("user:relation:friend:%s", userUUID), friendUUID},
			{fmt.Sprintf("user:relation:friend:%s", friendUUID), userUUID},
		}

		for _, pair := range pairs {
			// 检查缓存是否存在
			exists, err := r.redisClient.Exists(runCtx, pair.userKey).Result()
			if err != nil {
				LogRedisError(runCtx, err)
				continue
			}

			if exists > 0 {
				// 缓存存在，直接添加新好友到 Set
				pipe := r.redisClient.Pipeline()
				pipe.SRem(runCtx, pair.userKey, "__EMPTY__")
				pipe.SAdd(runCtx, pair.userKey, pair.newFriend)
				pipe.Expire(runCtx, pair.userKey, getRandomExpireTime(24*time.Hour))
				if _, err := pipe.Exec(runCtx); err != nil {
					LogRedisError(runCtx, err)
				}
			}
		}
	}, 0)
}
