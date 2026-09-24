package controllers

import (
	"net/http"

	"tiyu/models"

	"github.com/gin-gonic/gin"
)

// OrderQueryRequest 订单查询请求参数结构体（严格对齐 ShopQueryRequest 模式）

// OrderResponse 定义返回给前端的带有计算字段的订单响应结构体

// OrderQueryRequest 订单查询请求参数结构体（严格对齐 ShopQueryRequest 模式）
type OrderQueryRequest struct {
	Limit         int    `json:"limit" form:"limit"`
	Page          int    `json:"page" form:"page"`
	Order         string `json:"order" form:"order"`
	ID            uint64 `json:"id" form:"id"`                         // 订单主键ID
	OrderNo       string `json:"order_no" form:"order_no"`             // 订单号（支持模糊查询）
	UserID        uint64 `json:"user_id" form:"user_id"`               // 会员ID
	ShopID        int64  `json:"shop_id" form:"shop_id"`               // 商户ID
	TicketID      uint64 `json:"ticket_id" form:"ticket_id"`           // 票根ID
	CouponID      uint64 `json:"coupon_id" form:"coupon_id"`           // 优惠券ID
	PayType       int8   `json:"pay_type" form:"pay_type"`             // 支付模式 (0:未选择 1:在线支付 2:当面核销付)
	IsLock        *int8  `json:"is_lock" form:"is_lock"`               // 锁定状态 (用指针支持 0:已释放, 1:已锁定)
	Status        int8   `json:"status" form:"status"`                 // 订单状态 (1:待支付 2:待核销 3:已核销 4:已取消)
	PaymentMethod string `json:"payment_method" form:"payment_method"` // 支付通道 (alipay/wechat)
	VerifierID    uint64 `json:"verifier_id" form:"verifier_id"`       // 核销店员/商家ID
	StartTime     string `json:"start_time" form:"start_time"`         // 开始时间
	EndTime       string `json:"end_time" form:"end_time"`             // 结束时间
}

// OrderResponse 定义返回给前端的带有计算字段的订单响应结构体
type OrderResponse struct {
	models.Order // 嵌入原有的 Order 模型，直接继承所有底层字段

	// --- 商户配置费率与计算字段 ---
	Shopname               string  `json:"shop_name"`                // 平台抽成比例(%)
	CommissionRate         float64 `json:"commission_rate"`          // 平台抽成比例(%)
	ServiceFee             float64 `json:"service_fee"`              // 固定服务费(元)
	TransactionFeeRate     float64 `json:"transaction_fee_rate"`     // 通道手续费率(%)
	PlatformCommissionFee  float64 `json:"platform_commission_fee"`  // 计算得出的抽成金额
	TransactionFeeAmount   float64 `json:"transaction_fee_amount"`   // 计算得出的通道手续费金额
	TotalPlatformDeduction float64 `json:"total_platform_deduction"` // 平台总扣费
	MerchantIncome         float64 `json:"merchant_income"`          // 商户实际到账金额
}

// GetOrderList 获取订单列表（后台通用管理接口）
// GetOrderList 获取订单列表（后台通用管理接口）
func GetOrderList(c *gin.Context) {
	var req OrderQueryRequest
	_ = c.ShouldBind(&req)

	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	modelReq := &models.OrderQueryRequest{
		Limit:         req.Limit,
		Page:          req.Page,
		Order:         req.Order,
		ID:            req.ID,
		OrderNo:       req.OrderNo,
		UserID:        req.UserID,
		ShopID:        req.ShopID,
		TicketID:      req.TicketID,
		CouponID:      req.CouponID,
		PayType:       req.PayType,
		IsLock:        req.IsLock,
		Status:        req.Status,
		PaymentMethod: req.PaymentMethod,
		VerifierID:    req.VerifierID,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
	}

	listdata := models.GetMemberOrderList(req.Limit, req.Page, modelReq, req.Order)

	shopCache := make(map[int64]*models.Shop)

	var resultList []OrderResponse
	for _, item := range listdata {
		if item == nil {
			continue
		}

		var shopName string = "—"
		var commissionRate float64 = 0.00
		var serviceFee float64 = 0.00
		var transactionFeeRate float64 = 0.00

		if item.ShopID > 0 {
			shop, exists := shopCache[item.ShopID]
			if !exists {
				shop, _ = models.GetShopByID(item.ShopID)
				shopCache[item.ShopID] = shop
			}
			if shop != nil {
				// 注意：请根据你 models.Shop 中实际的商户名字段调整（如 shop.Name 或 shop.ShopName）
				shopName = shop.Name
				commissionRate = shop.CommissionRate
				serviceFee = shop.ServiceFee
				transactionFeeRate = shop.TransactionFeeRate
			}
		}

		payableAmount := item.PayableAmount
		commissionAmount := payableAmount * (commissionRate / 100.0)
		transactionFee := payableAmount * (transactionFeeRate / 100.0)
		totalPlatformDeduction := commissionAmount + serviceFee + transactionFee
		merchantIncome := payableAmount - totalPlatformDeduction

		resp := OrderResponse{
			Order:                  *item,
			Shopname:               shopName,
			CommissionRate:         commissionRate,
			ServiceFee:             serviceFee,
			TransactionFeeRate:     transactionFeeRate,
			PlatformCommissionFee:  commissionAmount,
			TransactionFeeAmount:   transactionFee,
			TotalPlatformDeduction: totalPlatformDeduction,
			MerchantIncome:         merchantIncome,
		}

		resultList = append(resultList, resp)
	}

	total := models.GetMemberOrderTotal(modelReq)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": gin.H{
			"list":  resultList,
			"total": total,
			"page":  req.Page,
			"limit": req.Limit,
		},
	})
}
