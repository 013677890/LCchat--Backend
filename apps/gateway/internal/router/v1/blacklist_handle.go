package v1

import (
	"ChatServer/apps/gateway/internal/dto"
	"ChatServer/apps/gateway/internal/middleware"
	"ChatServer/apps/gateway/internal/service"
	"ChatServer/apps/gateway/internal/utils"
	"ChatServer/consts"
	"ChatServer/pkg/logger"
	"ChatServer/pkg/result"

	"github.com/gin-gonic/gin"
)

// BlacklistHandler 黑名单处理器
type BlacklistHandler struct {
	blacklistService service.BlacklistService
}

// NewBlacklistHandler 创建黑名单处理器
func NewBlacklistHandler(blacklistService service.BlacklistService) *BlacklistHandler {
	return &BlacklistHandler{
		blacklistService: blacklistService,
	}
}

// AddBlacklist 拉黑用户接口
// @Summary 拉黑用户
// @Description 将用户加入黑名单
// @Tags 黑名单接口
// @Accept json
// @Produce json
// @Param request body dto.AddBlacklistRequest true "拉黑用户请求"
// @Success 200 {object} dto.AddBlacklistResponse
// @Router /api/v1/auth/blacklist [post]
func (h *BlacklistHandler) AddBlacklist(c *gin.Context) {
	ctx := middleware.NewContextWithGin(c)

	var req dto.AddBlacklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		result.Fail(c, nil, consts.CodeParamError)
		return
	}

	resp, err := h.blacklistService.AddBlacklist(ctx, &req)
	if err != nil {
		if consts.IsNonServerError(utils.ExtractErrorCode(err)) {
			result.Fail(c, nil, utils.ExtractErrorCode(err))
			return
		}

		logger.Error(ctx, "拉黑用户服务内部错误",
			logger.ErrorField("error", err),
		)
		result.Fail(c, nil, consts.CodeInternalError)
		return
	}

	result.Success(c, resp)
}
