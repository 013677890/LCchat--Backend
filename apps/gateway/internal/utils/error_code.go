package utils

import (
	"strconv"

	"github.com/013677890/LCchat-Backend/consts"
	"google.golang.org/grpc/status"
)

// ExtractErrorCode提取业务错误码
func ExtractErrorCode(err error) int {
	if err == nil {
		return 0
	}

	// 优先从 gRPC status message 提取业务错误码（user 服务约定：message=业务码字符串）
	if st, ok := status.FromError(err); ok {
		if bizCode, parseErr := strconv.Atoi(st.Message()); parseErr == nil {
			return bizCode
		}
		return consts.CodeInternalError
	}

	return consts.CodeInternalError
}
