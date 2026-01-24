package repository

const (
	// luaIncrementWithExpire 递增计数器，仅在首次创建时设置过期时间
	// KEYS[1]: 计数器 key
	// ARGV[1]: 过期时间（秒）
	// 返回: 递增后的值
	luaIncrementWithExpire = `
local key = KEYS[1]
local expire = tonumber(ARGV[1])
local current = redis.call('INCR', key)

-- 如果是第一次创建值为1,则设置过期时间
if current == 1 then
	redis.call('EXPIRE', key, expire)
end

return current
`
)