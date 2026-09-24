package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"
	"tiyu/models"                // 👈 加上别名 shopModel
	shopModel "tiyu/models/shop" // 👈 加上别名 shopModel

	"github.com/gin-gonic/gin"
)

// UnifiedCreateOrderRequest 统一下单请求结构体（兼容到店买单与团购商品购买）
type UnifiedCreateOrderRequest struct {
	ShopID         int64    `json:"shop_id" binding:"required"`
	ProductID      int64    `json:"product_id"`      // 0: 无商品到店买单；>0: 团购商品/套餐ID
	ProductName    string   `json:"product_name"`    // 商品/套餐名称
	Price          float64  `json:"price"`           // 商品单价 (买单模式下无需传，取 amount)
	Quantity       int      `json:"quantity"`        // 购买数量 (买单模式下默认为 1)
	TicketID       uint64   `json:"ticket_id"`       // 挂载的票根 ID
	CouponID       uint64   `json:"coupon_id"`       // 挂载的优惠券 ID
	Amount         float64  `json:"amount"`          // 原价总额
	DiscountAmount float64  `json:"discount_amount"` // 优惠折抵金额
	PayableAmount  *float64 `json:"payable_amount"`  // 实付金额
	PayType        int8     `json:"pay_type"`        // 1:在线支付 2:当面核销付
}

// CreateUnifiedOrderHandler 统一下单 Handler（前端仅需调用这一个接口）
// CreateUnifiedOrderHandler 统一下单 Handler
func CreateUnifiedOrderHandler(c *gin.Context) {
	// 1. 获取登录用户 ID
	userIDValue, exists := c.Get("userId")
	userID, ok := userIDValue.(int64)
	if !exists || !ok || userID <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 401, "message": "请先登录"})
		return
	}

	var req UnifiedCreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误", "data": err.Error()})
		return
	}

	// 2. 校验 PayType (1:在线支付 2:当面核销付)
	// if req.PayType != 1 && req.PayType != 2 {
	// 	c.JSON(http.StatusOK, gin.H{"code": 400, "message": "请选择有效的支付模式"})
	// 	return
	// }

	if req.ShopID <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "请选择有效的商户"})
		return
	}

	// 3. 校验商户有效性
	shop, err := models.GetShopByID(req.ShopID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询商户失败", "data": err.Error()})
		return
	}
	if shop == nil || shop.Status != 1 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "商户不存在或未营业"})
		return
	}

	// 4. 区分场景计算【原价总额】（从数据库校验商品真实价格，防止前端篡改）
	var calculatedOriginal float64
	var actualProductName string
	var actualProductPrice float64
	var ticketDiscountAmount float64 = 0.00 // 凭票根产生的立减折抵金额
	var memberTicket *models.MemberTicket   // 用于后续关联或透传

	if req.ProductID > 0 {
		// ================= 场景 A：团购商品购买 =================
		if req.Quantity <= 0 {
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "购买数量必须大于0"})
			return
		}

		// 调用 shopModel 包查询数据库真实商品信息
		product, err := shopModel.GetGoodsProductByID(req.ProductID)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询商品信息失败", "data": err.Error()})
			return
		}
		if product == nil || product.ShopId != req.ShopID || product.Status != 2 {
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "团购商品不存在、已下架或不属于该商户"})
			return
		}

		// 采用数据库真实价格与名称
		actualProductName = product.Title
		actualProductPrice = product.SellingPrice
		calculatedOriginal = actualProductPrice * float64(req.Quantity)

		// 🌟 核心逻辑：若上传/勾选了票根（req.TicketID > 0），进行票根联动规则校验与算价
		if req.TicketID > 0 {
			// 1. 查询商品是否有配置票根专属优惠规则
			ticketRule, err := shopModel.GetGoodsTicketDiscountByProductID(req.ProductID)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询票根优惠规则失败", "data": err.Error()})
				return
			}
			if ticketRule == nil || ticketRule.IsEnabled != 1 {
				c.JSON(http.StatusOK, gin.H{"code": 400, "message": "该商品未开启票根专属优惠"})
				return
			}

			// 2. 查询用户票根记录
			memberTicket, err = models.GetMemberTicketByIDAndUserID(req.TicketID, uint64(userID))
			if err != nil || memberTicket == nil {
				c.JSON(http.StatusOK, gin.H{"code": 400, "message": "票根不存在或不属于当前用户"})
				return
			}

			// 3. 校验票根规则并计算折抵金额
			ticketDiscountAmount, err = ticketRule.ValidateAndCalculate(
				memberTicket.ExchangeStatus,
				memberTicket.TicketCategory,
				memberTicket.TicketMainCategory,
				memberTicket.Title,
				memberTicket.OCRRawJSON,
				memberTicket.EventDate,
				calculatedOriginal,
			)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"code": 400, "message": "票根校验未通过: " + err.Error()})
				return
			}
		}
	} else {
		// ================= 场景 B：无商品到店优惠买单 =================
		calculatedOriginal = req.Amount
		req.Quantity = 1
	}

	// 5. 校验优惠券（仅当选择优惠券 req.CouponID > 0 时才校验）
	var coupon *models.Coupon
	if req.CouponID > 0 {
		coupon, err = models.GetCouponByID(req.CouponID)
		if err != nil || coupon == nil || coupon.ShopId != req.ShopID {
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "优惠券不属于当前商户或不存在"})
			return
		}
	}

	// 6. 解析原价与应付金额
	var originalAmount, payableAmount float64
	if req.PayableAmount == nil {
		payableAmount = calculatedOriginal
		originalAmount = payableAmount + req.DiscountAmount
	} else {
		originalAmount, payableAmount = models.ResolveOrderAmounts(calculatedOriginal, req.DiscountAmount, *req.PayableAmount)
	}

	if originalAmount <= 0 || payableAmount < 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "订单金额参数必须有效"})
		return
	}

	if shop.MinOrderAmount > 0 && originalAmount < shop.MinOrderAmount {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "未达到商户起送/最低消费金额"})
		return
	}

	// 7. 计算与校验优惠金额（总优惠 = 优惠券优惠 + 票根专属优惠）
	var couponDiscount float64 = 0.00
	if coupon != nil {
		// 选择优惠券：计算优惠券折抵金额
		couponDiscount, err = models.CalculateCouponDiscount(coupon, originalAmount)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": err.Error()})
			return
		}
	}

	// 汇总所有优惠金额（优惠券 + 票根立减/打折）
	totalDiscountAmount := couponDiscount + ticketDiscountAmount

	// 若未选择优惠券且没有票根优惠，但前端传了折抵金额，则拦截
	if coupon == nil && req.TicketID <= 0 && req.DiscountAmount > 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "未选择优惠券或票根，无法享受优惠折抵"})
		return
	}

	// 计算预期实付金额
	expectedPayable := originalAmount - totalDiscountAmount
	if expectedPayable < 0 {
		expectedPayable = 0
	}

	// 校验前端传来的折抵金额与实付金额是否和后端计算一致
	if !models.OrderAmountsMatch(req.DiscountAmount, payableAmount, totalDiscountAmount, expectedPayable) {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "订单优惠金额或实付金额校验不一致"})
		return
	}

	// 8. 决定初始状态：PayType 1(在线) => Status 1(待支付)；PayType 2(当面核销付) => Status 2(待核销)
	orderStatus := int8(1)
	if req.PayType == 2 {
		orderStatus = 2
	}

	// 9. 构造 Order 对象
	orderNo := models.NewOrderNo(uint64(userID))
	order := &models.Order{
		OrderNo:        orderNo,
		UserID:         uint64(userID),
		ShopID:         req.ShopID,
		TicketID:       req.TicketID,
		CouponID:       req.CouponID,
		OriginalAmount: originalAmount,
		DiscountAmount: totalDiscountAmount,
		PayableAmount:  payableAmount,
		PayType:        req.PayType,
		IsLock:         1,
		Status:         orderStatus,
	}

	// 10. 根据是否含有团购商品分流落库
	if req.ProductID > 0 {
		// ================= 场景 A：团购商品订单（创建订单及明细） =================
		var orderItems []*models.OrderItem
		for i := 0; i < req.Quantity; i++ {
			orderItem := &models.OrderItem{
				OrderNo:     orderNo,
				ShopId:      req.ShopID,
				UserId:      userID,
				ProductId:   req.ProductID,
				ProductName: actualProductName,
				Price:       actualProductPrice,
				Quantity:    1,
				VerifyCode:  models.GenerateVerifyCode(), // 每张券独立核销码
				Status:      1,                           // 1: 待核销
			}
			orderItems = append(orderItems, orderItem)
		}

		createdOrder, items, err := models.CreateOrderWithItemsTransaction(order, orderItems)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "message": "创建团购订单失败", "data": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "创建团购订单成功",
			"data": gin.H{
				"order": createdOrder,
				"items": items,
			},
		})
		return
	}

	// ================= 场景 B：无商品到店买单/票根优惠券订单 =================
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

// ---------------- 以下为你原有的查询、核销等辅助 Handler，保持不变 ----------------

// GetMyOrderListHandler 获取当前会员的订单分页列表
func GetMyOrderListHandler(c *gin.Context) {
	userIDValue, exists := c.Get("userId")
	userID, ok := userIDValue.(int64)
	if !exists || !ok || userID <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 401, "message": "请先登录"})
		return
	}

	var req models.OrderQueryRequest
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

	req.UserID = uint64(userID)

	list := models.GetMemberOrderList(req.Limit, req.Page, &req, req.Order)
	total := models.GetMemberOrderTotal(&req)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":  list,
			"total": total,
			"page":  req.Page,
			"limit": req.Limit,
		},
	})
}

// GetShopOrderListHandler 商户后台获取商户所有订单
func GetShopOrderListHandler(c *gin.Context) {
	userIDValue, exists := c.Get("userId")
	userID, ok := userIDValue.(int64)
	if !exists || !ok || userID <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 401, "message": "请先登录"})
		return
	}

	shop, err := models.GetShopByUserID(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询关联商户失败", "data": err.Error()})
		return
	}
	if shop == nil || shop.ID <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 403, "message": "当前账号未绑定有效商户"})
		return
	}

	var req models.OrderQueryRequest
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

	req.ShopID = shop.ID

	list := models.GetMemberOrderList(req.Limit, req.Page, &req, req.Order)
	total := models.GetMemberOrderTotal(&req)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"shop_id": shop.ID,
			"list":    list,
			"total":   total,
			"page":    req.Page,
			"limit":   req.Limit,
		},
	})
}

// VerifyShopOrderRequest 商家核销请求结构体
type VerifyShopOrderRequest struct {
	OrderNo   string `json:"order_no" binding:"required"`
	Timestamp int64  `json:"timestamp"`
	Sign      string `json:"sign"`
}

const QRCodeSecret = "YourAppSecretKeyKey2026"

// VerifyShopOrderHandler 商家核销订单
func VerifyShopOrderHandler(c *gin.Context) {
	userIDValue, exists := c.Get("userId")
	userID, ok := userIDValue.(int64)
	if !exists || !ok || userID <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 401, "message": "请先登录"})
		return
	}

	var req VerifyShopOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误", "data": err.Error()})
		return
	}

	if req.Sign != "" && req.Timestamp > 0 {
		now := time.Now().Unix()
		if now-req.Timestamp > 60 || req.Timestamp-now > 60 {
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "二维码已过期，请提示会员刷新"})
			return
		}

		expectedSign := generateQRSign(req.OrderNo, req.Timestamp, QRCodeSecret)
		if !hmac.Equal([]byte(req.Sign), []byte(expectedSign)) {
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的核销二维码"})
			return
		}
	}

	order, err := models.GetOrderByNo(req.OrderNo)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询订单失败", "data": err.Error()})
		return
	}
	if order == nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "订单不存在"})
		return
	}

	shop, err := models.GetShopByUserID(userID)
	if err != nil || shop == nil {
		c.JSON(http.StatusOK, gin.H{"code": 403, "message": "无权限操作当前商户订单"})
		return
	}
	if order.ShopID != shop.ID {
		c.JSON(http.StatusOK, gin.H{"code": 403, "message": "该订单不属于您的商户，无法核销"})
		return
	}

	if order.Status == 3 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "订单已核销过，请勿重复核销"})
		return
	}
	if order.Status == 4 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "订单已取消，无法核销"})
		return
	}
	if order.PayType == 1 && order.Status == 1 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "订单未支付，无法核销"})
		return
	}

	err = models.VerifyOrderTransaction(order, uint64(userID))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "核销失败", "data": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "核销成功",
		"data": gin.H{
			"order_no":    order.OrderNo,
			"verified_at": order.VerifiedAt,
		},
	})
}

func generateQRSign(orderNo string, timestamp int64, secret string) string {
	message := fmt.Sprintf("%s:%d", orderNo, timestamp)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil)) // 改为 hex.EncodeToString
}

// UpdatePayTypeRequest 修改支付方式请求结构体
type UpdatePayTypeRequest struct {
	OrderNo string `json:"order_no" binding:"required"`
	PayType int8   `json:"pay_type" binding:"required"`
}

// UpdateOrderPayTypeHandler 修改订单支付方式
func UpdateOrderPayTypeHandler(c *gin.Context) {
	var req UpdatePayTypeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	if req.PayType != 1 && req.PayType != 2 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "无效的支付方式类型"})
		return
	}

	if err := models.UpdateOrderPayType(req.OrderNo, req.PayType); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "修改支付方式成功"})
}
