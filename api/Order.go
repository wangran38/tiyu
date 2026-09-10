package api

import (
	"errors"
	"net/http"
	"tiyu/models"

	"github.com/gin-gonic/gin"
)

type CreateOrderRequest struct {
	ShopID         int64    `json:"shop_id" binding:"required"`
	TicketID       uint64   `json:"ticket_id"`
	CouponID       uint64   `json:"coupon_id"`
	Amount         float64  `json:"amount" binding:"required"` // 前端金额：已选券时通常为优惠后金额
	DiscountAmount float64  `json:"discount_amount"`           // 前端计算的优惠金额
	PayableAmount  *float64 `json:"payable_amount"`            // 可选，前端计算的实付金额
}

type MemberOrderListRequest struct {
	Limit   int    `json:"limit"`
	Page    int    `json:"page"`
	Order   string `json:"order"`
	Status  int8   `json:"status"`
	ShopID  int64  `json:"shop_id"`
	OrderNo string `json:"order_no"`
}

// GetMyOrderListHandler 获取当前会员的订单分页列表。
func GetMyOrderListHandler(c *gin.Context) {
	userIDValue, exists := c.Get("userId")
	userID, ok := userIDValue.(int64)
	if !exists || !ok || userID <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 401, "message": "请先登录"})
		return
	}

	var req MemberOrderListRequest
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

	search := &models.Order{
		UserID:  uint64(userID),
		Status:  req.Status,
		ShopID:  req.ShopID,
		OrderNo: req.OrderNo,
	}
	list := models.GetMemberOrderList(req.Limit, req.Page, search, req.Order)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":  list,
			"total": models.GetMemberOrderTotal(search),
			"page":  req.Page,
			"limit": req.Limit,
		},
	})
}

// CreateOrderHandler 会员使用票根优惠券创建商户订单。
func CreateOrderHandler(c *gin.Context) {
	userIDValue, exists := c.Get("userId")
	userID, ok := userIDValue.(int64)
	if !exists || !ok || userID <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 401, "message": "请先登录"})
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误", "data": err.Error()})
		return
	}
	if req.ShopID <= 0 || req.Amount < 0 || req.DiscountAmount < 0 || (req.PayableAmount != nil && *req.PayableAmount < 0) {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "订单金额参数必须有效"})
		return
	}

	shop, err := models.GetShopByID(req.ShopID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询商户失败", "data": err.Error()})
		return
	}
	if shop == nil || shop.Status != 1 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "商户不存在或未营业"})
		return
	}

	var coupon *models.Coupon
	if req.CouponID > 0 {
		coupon, err = models.GetCouponByID(req.CouponID)
		if err != nil || coupon == nil || coupon.ShopId != req.ShopID {
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "优惠券不属于当前商户"})
			return
		}
	}
	var originalAmount, payableAmount float64
	if req.PayableAmount == nil {
		payableAmount = req.Amount
		originalAmount = payableAmount + req.DiscountAmount
	} else {
		originalAmount, payableAmount = models.ResolveOrderAmounts(req.Amount, req.DiscountAmount, *req.PayableAmount)
	}
	if originalAmount <= 0 || payableAmount < 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "订单金额参数必须有效"})
		return
	}
	if shop.MinOrderAmount > 0 && originalAmount < shop.MinOrderAmount {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "未达到商户起送金额"})
		return
	}

	var discountAmount float64
	if coupon != nil {
		discountAmount, err = models.CalculateCouponDiscount(coupon, originalAmount)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": err.Error()})
			return
		}
	}
	expectedPayable := originalAmount - discountAmount
	if expectedPayable < 0 {
		expectedPayable = 0
	}
	if !models.OrderAmountsMatch(req.DiscountAmount, payableAmount, discountAmount, expectedPayable) {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "订单优惠金额或实付金额校验不一致"})
		return
	}
	order := &models.Order{
		OrderNo:        models.NewOrderNo(uint64(userID)),
		UserID:         uint64(userID),
		ShopID:         req.ShopID,
		TicketID:       req.TicketID,
		CouponID:       req.CouponID,
		OriginalAmount: originalAmount,
		DiscountAmount: req.DiscountAmount,
		PayableAmount:  payableAmount,
	}
	createdOrder, ticket, usedCoupon, err := models.CreateOrderWithTicketCoupon(order)
	if errors.Is(err, models.ErrOrderTicketInvalid) {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "票根不存在或已经兑换"})
		return
	}
	if errors.Is(err, models.ErrOrderCouponInvalid) {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "优惠券不可用或库存不足"})
		return
	}
	if errors.Is(err, models.ErrOrderAmountInvalid) {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "订单优惠金额或实付金额校验不一致"})
		return
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "创建订单失败", "data": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "订单创建成功",
		"data": gin.H{
			"order":  createdOrder,
			"ticket": ticket,
			"coupon": usedCoupon,
		},
	})
}
