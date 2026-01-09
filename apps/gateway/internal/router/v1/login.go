package v1

import (
	"ChatServer/apps/gateway/internal/dto"
	"ChatServer/apps/gateway/internal/middleware"
	"ChatServer/apps/gateway/internal/pb"
	"ChatServer/apps/gateway/internal/utils"
	userpb "ChatServer/apps/user/pb" // 引入user服务的protobuf消息类型，使用别名避免冲突
	"ChatServer/consts"
	"ChatServer/pkg/logger"
	"ChatServer/pkg/util"
	"time"

	"github.com/gin-gonic/gin"
)

// Login 用户登录接口
// @Summary 用户登录
// @Description 用户通过手机号和密码登录
// @Tags 认证接口
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "登录请求"
// @Success 200 {object} dto.LoginResponse
// @Router /api/v1/public/login [post]
func Login(c *gin.Context) {
	ctx := middleware.NewContextWithGin(c)
	traceId := c.GetString("trace_id")
	ip := c.ClientIP()

	// 1. 绑定请求数据
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 参数错误由客户端输入导致,属于正常业务流程,不记录日志
		c.JSON(400, gin.H{
			"code":    consts.CodeParamError,
			"message": consts.GetMessage(consts.CodeParamError),
			"errors":  []gin.H{{"message": err.Error()}},
		})
		return
	}

	// 2. 业务参数合法性校验
	if len(req.Telephone) != 11 {
		// 用户输入错误,属于正常业务流程,不记录日志
		c.JSON(400, gin.H{
			"code":    consts.CodePhoneError,
			"message": consts.GetMessage(consts.CodePhoneError),
		})
		return
	}

	if len(req.Password) == 0 {
		// 用户输入错误,属于正常业务流程,不记录日志
		c.JSON(400, gin.H{
			"code":    consts.CodeParamError,
			"message": consts.GetMessage(consts.CodeParamError),
		})
		return
	}

	// 3. 调用用户服务进行身份认证(gRPC)
	startTime := time.Now()

	grpcReq := &userpb.LoginRequest{
		Telephone: req.Telephone,
		Password:  req.Password,
	}

	// 调用 gRPC 接口，并使用重试机制提高服务稳定性
	grpcResp, err := pb.LoginWithRetry(ctx, grpcReq, 3)
	duration := time.Since(startTime)

	if err != nil {
		// 所有重试失败,记录错误日志
		logger.Error(ctx, "调用用户服务 gRPC 失败",
			logger.String("trace_id", traceId),
			logger.String("ip", ip),
			logger.String("telephone", utils.MaskTelephone(req.Telephone)),
			logger.ErrorField("error", err),
			logger.Duration("duration", duration),
		)
		c.JSON(500, gin.H{
			"code":    30001,
			"message": "服务暂时不可用",
		})
		return
	}

	// 4. 处理用户服务返回的业务响应
	if grpcResp.Code != 0 {
		// 用户认证失败(如密码错误、账号锁定等),属于正常业务流程,不记录日志
		c.JSON(400, gin.H{
			"code":    grpcResp.Code,
			"message": grpcResp.Message,
		})
		return
	}

	if grpcResp.UserInfo == nil {
		// 成功返回但 UserInfo 为空，属于非预期的异常情况
		logger.Error(ctx, "成功响应中用户信息为空",
			logger.String("trace_id", traceId),
		)
		c.JSON(500, gin.H{
			"code":    30001,
			"message": "服务器内部错误",
		})
		return
	}

	// 5. 令牌生成逻辑
	// 优先从 Header 获取设备唯一标识，若无则生成一个新的 UUID 标识当前设备
	deviceId := c.GetHeader("X-Device-ID")
	if deviceId == "" {
		deviceId = util.NewUUID()
		logger.Debug(ctx, "请求头中无设备ID，生成新设备ID",
			logger.String("trace_id", traceId),
			logger.String("device_id", deviceId),
		)
	}

	// 生成 Access Token，用于后续接口请求的身份校验
	accessToken, err := utils.GenerateToken(grpcResp.UserInfo.Uuid, deviceId)
	if err != nil {
		// Token 生成失败通常是内部算法或 JWT 配置问题
		logger.Error(ctx, "生成 Access Token 失败",
			logger.String("trace_id", traceId),
			logger.String("user_uuid", utils.MaskUUID(grpcResp.UserInfo.Uuid)),
			logger.ErrorField("error", err),
		)
		c.JSON(500, gin.H{
			"code":    30001,
			"message": "服务器内部错误",
		})
		return
	}

	// 生成 Refresh Token，用于 Access Token 过期后的无感刷新
	refreshToken, err := utils.GenerateRefreshToken(grpcResp.UserInfo.Uuid, deviceId)
	if err != nil {
		// Refresh Token 生成失败也视为系统异常
		logger.Error(ctx, "生成 Refresh Token 失败",
			logger.String("trace_id", traceId),
			logger.String("user_uuid", utils.MaskUUID(grpcResp.UserInfo.Uuid)),
			logger.ErrorField("error", err),
		)
		c.JSON(500, gin.H{
			"code":    30001,
			"message": "服务器内部错误",
		})
		return
	}

	// 6. 构造响应
	response := dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(utils.AccessExpire / time.Second),
		UserInfo: dto.UserInfo{
			UUID:      grpcResp.UserInfo.Uuid,
			Nickname:  grpcResp.UserInfo.Nickname,
			Telephone: grpcResp.UserInfo.Telephone,
			Email:     grpcResp.UserInfo.Email,
			Avatar:    grpcResp.UserInfo.Avatar,
			Gender:    int8(grpcResp.UserInfo.Gender),
			Signature: grpcResp.UserInfo.Signature,
			Birthday:  grpcResp.UserInfo.Birthday,
		},
	}

	// 8. 记录登录成功日志
	totalDuration := time.Since(startTime)
	logger.Info(ctx, "登录成功",
		logger.String("trace_id", traceId),
		logger.String("ip", ip),
		logger.String("user_uuid", utils.MaskUUID(grpcResp.UserInfo.Uuid)),
		logger.String("telephone", utils.MaskTelephone(grpcResp.UserInfo.Telephone)),
		logger.String("nickname", grpcResp.UserInfo.Nickname),
		logger.String("platform", req.DeviceInfo.Platform),
		logger.Duration("total_duration", totalDuration),
	)

	c.JSON(200, gin.H{
		"code":    0,
		"message": "登录成功",
		"data":    response,
	})
}
