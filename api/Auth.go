package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"tiyu/lib"    // 替换为你项目的 lib 包路径 (包含 Password 函数)
	"tiyu/models" // 替换为你项目的 models 包路径
	"tiyu/utils"  // 替换为你项目的 utils 包路径

	"github.com/gin-gonic/gin"
)

// PhoneSmsLoginReq 手机号验证码快捷登录请求体
type PhoneSmsLoginReq struct {
	Mobile string `json:"mobile" binding:"required"` // 手机号
	Code   string `json:"code" binding:"required"`   // 短信验证码
}

// QuickLoginByPhone 手机号 + 验证码快捷登录/自动注册 API
// @Router /api/v1/auth/quick-login [post]
func QuickLoginByPhone(c *gin.Context) {
	var req PhoneSmsLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "请求参数不合法，请输入手机号和验证码"})
		return
	}

	mobile := strings.TrimSpace(req.Mobile)
	code := strings.TrimSpace(req.Code)

	// 1. 基础手机号格式校验
	if len(mobile) != 11 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "手机号格式不正确"})
		return
	}

	// 2. 校验验证码 (支持测试万能后门验证码 888888)
	if code == "888888" {
		// 如果是后门/测试验证码 888888，直接通过，不做 Redis 验证码比对
	} else {
		// 正常逻辑：校验 Redis 中的验证码
		ok, msg := utils.VerifyAndHandleSmsCode(mobile, "login", code)
		if !ok {
			c.JSON(http.StatusOK, gin.H{"code": 400, "msg": msg})
			return
		}
	}

	// 3. 查询数据库判断用户是否存在
	var user models.User
	has, err := models.Dorm.Where("mobile = ?", mobile).Get(&user)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "数据库查询异常"})
		return
	}

	// 4. 如果用户不存在，进行“自动注册”
	isNewUser := false
	if !has {
		isNewUser = true

		// 随机生成一个默认密码并加盐哈希保存 (用户后续可去修改密码)
		randomPassword := utils.GenerateRandomCode(8)
		hashedPassword, salt := lib.Password(6, randomPassword)

		maskedMobile := mobile[:3] + "****" + mobile[7:]
		defaultUsername := "u_" + mobile[7:]

		user = models.User{
			Username:    defaultUsername,
			Nickname:    "用户" + maskedMobile,
			Mobile:      mobile,
			Password:    hashedPassword,
			Salt:        salt,
			Avatar:      "",
			Status:      "normal",
			Point:       0,
			Level:       1,
			LastLoginIp: c.ClientIP(),
			LastLoginAt: time.Now().Unix(),
		}

		// 插入新用户
		_, err = models.Dorm.Insert(&user)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "自动创建账号失败，请稍后重试"})
			return
		}
	} else {
		// 如果用户存在，判断账号状态是否被禁用
		if user.Status == "disabled" || user.Status == "banned" {
			c.JSON(http.StatusOK, gin.H{"code": 403, "msg": "该账号已被禁用，请联系管理员"})
			return
		}

		// 更新最后登录 IP 和时间
		user.LastLoginIp = c.ClientIP()
		user.LastLoginAt = time.Now().Unix()
		models.Dorm.Id(user.Id).Cols("last_login_ip", "last_login_at").Update(&user)
	}

	// 5. 签发 JWT Token
	token, err := utils.GenerateToken(user.Id, user.Mobile)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "Token 签发失败"})
		return
	}

	// 6. 成功返回
	loginMsg := "登录成功"
	if isNewUser {
		loginMsg = "注册并登录成功"
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  loginMsg,
		"data": gin.H{
			"token":       token,
			"user_id":     user.Id,
			"mobile":      user.Mobile,
			"nickname":    user.Nickname,
			"username":    user.Username,
			"avatar":      user.Avatar,
			"is_new_user": isNewUser,
		},
	})
}

// SendSmsReq 发送短信请求体
type SendSmsReq struct {
	Mobile string `json:"mobile" binding:"required"` // 手机号
}

// SendLoginSms 发送快捷登录验证码 API
// @Router /api/v1/auth/send-sms [post]
func SendLoginSms(c *gin.Context) {
	var req SendSmsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "请输入手机号"})
		return
	}

	mobile := strings.TrimSpace(req.Mobile)
	if len(mobile) != 11 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "手机号格式不正确"})
		return
	}

	// 1. 防刷校验：检查 Redis 中该手机号 60 秒内是否已经发过
	limitKey := fmt.Sprintf("sms_limit_%s_%s", "login", mobile)
	if utils.CheckRedisExits(limitKey) == 1 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "发送太频繁，请 60 秒后再试"})
		return
	}

	// 2. 生成 6 位随机验证码
	code := utils.GenerateRandomCode(6)

	// 3. 将验证码存入 Redis (设置 5 分钟有效 = 300秒)
	smsKey := fmt.Sprintf("sms_%s_%s", "login", mobile)
	utils.InsertRedisKeyExpire(smsKey, code, 300)

	// 4. 设置 60 秒防刷频 key
	utils.InsertRedisKeyExpire(limitKey, "1", 60)

	// 5. 动态读取数据库 SmsConfig 并通过腾讯云发送短信
	err := utils.SendDynamicSmsCode(mobile, "login", code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "发送失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "验证码发送成功",
	})
}
