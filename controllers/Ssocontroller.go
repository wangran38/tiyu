package controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"tiyu/models" // 假设你的数据库模型在 models 包
	"tiyu/utils"  // 假设你的 JWT 签发工具在 utils 包

	"github.com/gin-gonic/gin"
)

// SSOLoginRequest 接收前端请求参数
type SSOLoginRequest struct {
	SSOCode string `json:"sso_code" binding:"required"`
}

// TicketSystemVerifyRequest 向票务系统(PHP)校验的参数结构
type TicketSystemVerifyRequest struct {
	SSOCode string `json:"sso_code"`
	Secret  string `json:"secret"` // 跨服务器通信秘钥，防止接口被伪造请求
}

// TicketSystemVerifyResponse 票务系统(PHP)返回的结构
type TicketSystemVerifyResponse struct {
	Code int `json:"code"`
	Data struct {
		Mobile   string `json:"mobile"`
		Nickname string `json:"nickname"`
	} `json:"data"`
	Msg string `json:"msg"`
}

// SSOLogin 处理跨系统免登录
func SSOLogin(c *gin.Context) {
	var req SSOLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数格式错误或缺失 sso_code",
		})
		return
	}

	// 1. 配置票务系统的校验接口地址和秘钥 (根据实际情况修改)
	phpVerifyURL := "https://ticket.yourdomain.com/api/verifySsoCode"
	ssoSecret := "your_sso_communication_secret_key" // 两端约定的加解密/通信密钥

	// 2. 组装请求向票务系统 (PHP) 发送 POST 校验请求
	postData := TicketSystemVerifyRequest{
		SSOCode: req.SSOCode,
		Secret:  ssoSecret,
	}
	jsonData, err := json.Marshal(postData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统内部错误"})
		return
	}

	// 设置 HTTP 请求超时时间（防止 PHP 端无响应导致 Go 挂起）
	httpClient := &http.Client{Timeout: 5 * time.Second}
	resp, err := httpClient.Post(phpVerifyURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"code": 502,
			"msg":  "连接票务验证服务失败",
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "读取验证响应失败"})
		return
	}

	// 3. 解析票务系统返回的结果
	var phpResp TicketSystemVerifyResponse
	if err := json.Unmarshal(body, &phpResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "解析验证响应失败"})
		return
	}

	// 校验失败（如 code 过期或已失效）
	if phpResp.Code != 200 || phpResp.Data.Mobile == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  phpResp.Msg,
		})
		return
	}

	// 4. 根据手机号在票根系统数据库查询或创建用户
	mobile := phpResp.Data.Mobile
	user, err := models.FindOrCreateUserByMobile(mobile, phpResp.Data.Nickname)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "本地账号创建或更新失败",
		})
		return
	}

	// 5. 签发票根系统自己的 JWT Token
	token, err := utils.GenerateToken(user.Id, user.Mobile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "生成登录 Token 失败",
		})
		return
	}

	// 6. 返回登录成功信息与 Token
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "免登录成功",
		"data": gin.H{
			"token":   token,
			"user_id": user.Id,
			"mobile":  user.Mobile,
		},
	})
}
