package api

import (
	"net/http"
	"tiyu/global"
	"tiyu/models"

	"github.com/gin-gonic/gin"
)

// GetUserProfile 获取当前登录用户的个人资料
// @Router /api/user/profile [get]
func GetUserProfile(c *gin.Context) {
	// 从 context 中获取 utils.UserJWTAuth 注入的 userId
	userIDVal, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, gin.H{"code": 401, "msg": "未获取到用户登录状态"})
		return
	}
	userID := userIDVal.(int64)

	// 查询数据库
	var user models.User
	has, err := global.Dorm.ID(userID).Get(&user)
	if err != nil || !has {
		c.JSON(http.StatusOK, gin.H{"code": 404, "msg": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "获取成功",
		"data": gin.H{
			"user_id":  user.Id,
			"mobile":   user.Mobile,
			"nickname": user.Nickname,
			"avatar":   user.Avatar,
			"status":   user.Status,
			"point":    user.Point,
			"level":    user.Level,
		},
	})
}
