package consts

// 通用错误码
const (
	CodeSuccess = 0 // 成功
)

// 客户端错误 (1xxxx)
const (
	CodeParamError       = 10001 // 参数验证失败
	CodeBodyError        = 10002 // 请求体格式错误
	CodeResourceNotFound = 10003 // 资源不存在
	CodeMethodNotAllowed = 10004 // 请求方法不允许
	CodeTooManyRequests  = 10005 // 请求过于频繁
	CodeBodyTooLarge     = 10006 // 请求体过大
)

// 认证错误 (2xxxx)
const (
	CodeUnauthorized   = 20001 // 未认证
	CodeInvalidToken   = 20002 // Token 无效
	CodeTokenExpired   = 20003 // Token 已过期
	CodePermissionDeny = 20004 // 权限不足
)

// 用户模块错误 (11xxx)
const (
	CodeUserNotFound     = 11001 // 用户不存在
	CodeUserAlreadyExist = 11002 // 用户已存在
	CodePasswordError    = 11003 // 密码错误
	CodeUserDisabled     = 11004 // 用户已被禁用
	CodePhoneError       = 11005 // 手机号格式错误
	CodeVerifyCodeError  = 11006 // 验证码错误
	CodeVerifyCodeExpire = 11007 // 验证码已过期
)

// 好友模块错误 (12xxx)
const (
	CodeAlreadyFriend     = 12001 // 已经是好友
	CodeFriendRequestSent = 12002 // 好友申请已发送
	CodeNotFriend         = 12003 // 不存在该好友关系
	CodeIsBlacklist       = 12004 // 已经是黑名单
)

// 消息模块错误 (13xxx)
const (
	CodeMessageNotFound       = 13001 // 消息不存在
	CodeMessageSendFail       = 13002 // 消息发送失败
	CodeMessageTypeNotSupport = 13003 // 消息类型不支持
	CodeConversationNotFound  = 13004 // 会话不存在
)

// 群组模块错误 (14xxx)
const (
	CodeGroupNotFound  = 14001 // 群组不存在
	CodeNotGroupMember = 14002 // 不是群成员
	CodeNoPermission   = 14003 // 没有权限
	CodeGroupFull      = 14004 // 群成员已满
)

// 服务端错误 (3xxxx)
const (
	CodeInternalError      = 30001 // 服务器内部错误
	CodeServiceUnavailable = 30002 // 服务暂不可用
	CodeTimeoutError       = 30003 // 超时错误
)

// 错误消息映射
var CodeMessage = map[int32]string{
	CodeSuccess: "success",

	// 客户端错误
	CodeParamError:       "参数验证失败",
	CodeBodyError:        "请求体格式错误",
	CodeResourceNotFound: "资源不存在",
	CodeMethodNotAllowed: "请求方法不允许",
	CodeTooManyRequests:  "请求过于频繁",
	CodeBodyTooLarge:     "请求体过大",

	// 认证错误
	CodeUnauthorized:   "未认证",
	CodeInvalidToken:   "Token 无效",
	CodeTokenExpired:   "Token 已过期",
	CodePermissionDeny: "权限不足",

	// 用户模块
	CodeUserNotFound:     "用户不存在",
	CodeUserAlreadyExist: "用户已存在",
	CodePasswordError:    "密码错误",
	CodeUserDisabled:     "用户已被禁用",
	CodePhoneError:       "手机号格式错误",
	CodeVerifyCodeError:  "验证码错误",
	CodeVerifyCodeExpire: "验证码已过期",

	// 好友模块
	CodeAlreadyFriend:     "已经是好友",
	CodeFriendRequestSent: "好友申请已发送",
	CodeNotFriend:         "不存在该好友关系",
	CodeIsBlacklist:       "已经是黑名单",

	// 消息模块
	CodeMessageNotFound:       "消息不存在",
	CodeMessageSendFail:       "消息发送失败",
	CodeMessageTypeNotSupport: "消息类型不支持",
	CodeConversationNotFound:  "会话不存在",

	// 群组模块
	CodeGroupNotFound:  "群组不存在",
	CodeNotGroupMember: "不是群成员",
	CodeNoPermission:   "没有权限",
	CodeGroupFull:      "群成员已满",

	// 服务端错误
	CodeInternalError:      "服务器内部错误",
	CodeServiceUnavailable: "服务暂不可用",
	CodeTimeoutError:       "超时错误",
}

// GetMessage 根据错误码获取错误消息
func GetMessage(code int32) string {
	if msg, ok := CodeMessage[code]; ok {
		return msg
	}
	return "未知错误"
}
