package controllers

import (
	"tiyu/models"
	"tiyu/utils"

	"github.com/gin-gonic/gin"
)

// 从 token 拿当前登录的管理员账号
func currentOperator(c *gin.Context) string {
	token := c.Request.Header.Get("Authorization")
	if token == "" {
		return ""
	}
	return utils.GetLoginUser(token).Username
}

type UserFlowSearch struct {
	UserId   int64  `json:"user_id"`
	Username string `json:"username"`
	Type     string `json:"type"`
	Limit    int    `json:"limit"`
	Page     int    `json:"page"`
	Order    string `json:"sort"`
}

// 会员流水列表
func GetUserFlowlist(c *gin.Context) {
	var s UserFlowSearch
	c.BindJSON(&s)

	search := &models.UserFlow{
		UserId:   s.UserId,
		Username: s.Username,
		Type:     s.Type,
	}

	listdata := models.GetUserFlowList(s.Limit, s.Page, search, s.Order)
	listnum := models.GetUserFlowTotal(search)

	result := make(map[string]interface{})
	result["page"] = s.Page
	result["totalnum"] = listnum
	result["limit"] = s.Limit
	result["listdata"] = listdata

	c.JSON(200, gin.H{
		"code":    200,
		"message": "数据获取成功",
		"data":    result,
	})
}

// 手动调整会员积分（带流水记录）
func AdjustUserPoint(c *gin.Context) {
	var req struct {
		UserId int64  `json:"user_id"`
		Change int    `json:"change"`
		Reason string `json:"reason"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(200, gin.H{"code": 201, "message": "参数错误", "data": err.Error()})
		return
	}
	if req.UserId == 0 {
		c.JSON(200, gin.H{"code": 201, "message": "缺少 user_id", "data": ""})
		return
	}

	u, _ := models.SelectMemberById(req.UserId)
	if u == nil {
		c.JSON(200, gin.H{"code": 201, "message": "会员不存在", "data": ""})
		return
	}

	before := u.Point
	after := before + req.Change
	if after < 0 {
		c.JSON(200, gin.H{"code": 201, "message": "积分不足", "data": ""})
		return
	}

	u.Point = after
	err := models.EditMember(u)
	if err != nil {
		c.JSON(200, gin.H{"code": 201, "message": "积分更新失败", "data": err.Error()})
		return
	}

	// 写流水
	operator := currentOperator(c)
	_ = models.AddUserFlow(&models.UserFlow{
		UserId:   req.UserId,
		Username: u.Username,
		Type:     "point",
		Change:   req.Change,
		Before:   before,
		After:    after,
		Reason:   req.Reason,
		Operator: operator,
	})

	c.JSON(200, gin.H{
		"code":    200,
		"message": "积分调整成功",
		"data": gin.H{
			"before":  before,
			"after":   after,
			"change":  req.Change,
		},
	})
}

// 手动调整会员等级（带流水记录）
func AdjustUserLevel(c *gin.Context) {
	var req struct {
		UserId int64  `json:"user_id"`
		Level  int    `json:"level"`
		Reason string `json:"reason"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(200, gin.H{"code": 201, "message": "参数错误", "data": err.Error()})
		return
	}
	if req.UserId == 0 {
		c.JSON(200, gin.H{"code": 201, "message": "缺少 user_id", "data": ""})
		return
	}
	if req.Level < 1 || req.Level > 99 {
		c.JSON(200, gin.H{"code": 201, "message": "等级范围 1-99", "data": ""})
		return
	}

	u, _ := models.SelectMemberById(req.UserId)
	if u == nil {
		c.JSON(200, gin.H{"code": 201, "message": "会员不存在", "data": ""})
		return
	}

	before := u.Level
	after := req.Level
	if before == after {
		c.JSON(200, gin.H{"code": 201, "message": "等级未变化", "data": ""})
		return
	}

	u.Level = after
	err := models.EditMember(u)
	if err != nil {
		c.JSON(200, gin.H{"code": 201, "message": "等级更新失败", "data": err.Error()})
		return
	}

	operator := currentOperator(c)
	_ = models.AddUserFlow(&models.UserFlow{
		UserId:   req.UserId,
		Username: u.Username,
		Type:     "level",
		Change:   after - before,
		Before:   before,
		After:    after,
		Reason:   req.Reason,
		Operator: operator,
	})

	c.JSON(200, gin.H{
		"code":    200,
		"message": "等级调整成功",
		"data": gin.H{
			"before": before,
			"after":  after,
			"change": after - before,
		},
	})
}
