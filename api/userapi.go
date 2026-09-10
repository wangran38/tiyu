package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MockUserInfo 模拟返回用户昵称、手机号、unionid
func MockUserInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "获取模拟用户信息成功",
		"data": gin.H{
			"nickname": "体育达人",
			"mobile":   "13800138000",
			"unionid":  "wx_unionid_mock_1234567890abcdef",
		},
	})
}
