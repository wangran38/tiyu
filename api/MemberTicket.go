package api

import (
	"errors"
	"net/http"
	"tiyu/models"

	"github.com/gin-gonic/gin"
)

type RedeemMemberTicketRequest struct {
	TicketID uint64 `json:"ticket_id" binding:"required"`
	CouponID uint64 `json:"coupon_id" binding:"required"`
}

type MemberTicketListRequest struct {
	Limit          int    `json:"limit"`
	Page           int    `json:"page"`
	Order          string `json:"order"`
	ExchangeStatus int8   `json:"exchange_status"`
	TicketCategory string `json:"ticket_category"`
}

// GetMyMemberTicketListHandler 从 MySQL 获取当前会员的“我的票根”列表。
func GetMyMemberTicketListHandler(c *gin.Context) {
	userIDValue, exists := c.Get("userId")
	userID, ok := userIDValue.(int64)
	if !exists || !ok || userID <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 401, "message": "请先登录"})
		return
	}

	var req MemberTicketListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误", "data": err.Error()})
		return
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	search := &models.MemberTicket{
		UserID:         uint64(userID),
		ExchangeStatus: req.ExchangeStatus,
		TicketCategory: req.TicketCategory,
	}
	list := models.GetMemberTicketList(req.Limit, req.Page, search, req.Order)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":  list,
			"total": models.GetMemberTicketTotal(search),
			"page":  req.Page,
			"limit": req.Limit,
		},
	})
}

// RedeemMemberTicketHandler 使用当前会员的票根兑换优惠券。
func RedeemMemberTicketHandler(c *gin.Context) {
	userIDValue, exists := c.Get("userId")
	userID, ok := userIDValue.(int64)
	if !exists || !ok || userID <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 401, "message": "请先登录"})
		return
	}

	var req RedeemMemberTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误", "data": err.Error()})
		return
	}

	ticket, coupon, err := models.RedeemMemberTicket(req.TicketID, uint64(userID), req.CouponID)
	if errors.Is(err, models.ErrMemberTicketNotFound) {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "票根不存在或已经兑换"})
		return
	}
	if errors.Is(err, models.ErrCouponNotAvailable) {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "优惠券不可兑换或库存不足"})
		return
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "兑换优惠券失败", "data": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "优惠券兑换成功",
		"data": gin.H{
			"ticket": ticket,
			"coupon": coupon,
		},
	})
}
