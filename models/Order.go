package models

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
	"tiyu/global"
)

var (
	ErrOrderCouponInvalid = errors.New("coupon does not belong to shop or is unavailable")
	ErrOrderTicketInvalid = errors.New("ticket not found or already exchanged")
	ErrOrderAmountInvalid = errors.New("order amount does not match coupon rules")
)

// Order 会员在商户消费并使用票根优惠券生成的订单。
// Order 会员在商户消费并使用票根优惠券生成的订单。
type Order struct {
	ID      uint64 `xorm:"pk autoincr bigint 'id'" json:"id"`
	OrderNo string `xorm:"varchar(40) notnull unique 'order_no'" json:"order_no"`
	UserID  uint64 `xorm:"bigint notnull index 'user_id'" json:"user_id"`
	ShopID  int64  `xorm:"bigint notnull index 'shop_id'" json:"shop_id"`
	// TicketID       uint64  `xorm:"bigint notnull index 'ticket_id'" json:"ticket_id"`
	// CouponID       uint64  `xorm:"bigint notnull index 'coupon_id'" json:"coupon_id"`
	OriginalAmount float64 `xorm:"decimal(10,2) notnull 'original_amount'" json:"original_amount"`
	DiscountAmount float64 `xorm:"decimal(10,2) notnull default 0.00 'discount_amount'" json:"discount_amount"`
	PayableAmount  float64 `xorm:"decimal(10,2) notnull 'payable_amount'" json:"payable_amount"`

	// 支付方式：0:未选择 1:在线支付 2:到店/当面核销支付
	PayType int8 `xorm:"tinyint notnull default 0 'pay_type' comment('支付模式 0:未选择 1:在线支付 2:当面核销付')" json:"pay_type"`

	// 锁状态：1:已锁定(下单即锁定票根/券，防止并发或重复使用) 0:未锁定/已释放
	IsLock int8 `xorm:"tinyint notnull default 1 index 'is_lock' comment('锁定状态 1:已锁定 0:已释放')" json:"is_lock"`

	// 订单状态：1:待处理/待支付 2:待核销 3:已核销/已完成 4:已取消
	Status int8 `xorm:"tinyint notnull default 1 index 'status' comment('状态 1:待支付 2:待核销 3:已核销 4:已取消')" json:"status"`

	PaymentMethod    string     `xorm:"varchar(32) default '' 'payment_method' comment('支付通道: alipay/wechat等')" json:"payment_method"`
	PaidAt           *time.Time `xorm:"datetime null 'paid_at' comment('支付成功回调时间')" json:"paid_at"`
	PaymentNotifyURL string     `xorm:"varchar(512) default '' 'payment_notify_url' comment('支付回调通知URL')" json:"payment_notify_url"`
	// 关联优惠权益（非必填，根据业务组合使用）
	TicketID   uint64 `xorm:"bigint notnull default 0 index 'ticket_id' comment('票根ID')" json:"ticket_id"`
	CouponID   uint64 `xorm:"bigint notnull default 0 index 'coupon_id' comment('优惠券ID')" json:"coupon_id"`
	GroupBuyID uint64 `xorm:"bigint notnull default 0 index 'group_buy_id' comment('关联团购活动/商品ID')" json:"group_buy_id"`
	// 核销相关字段（建议补充）
	VerifiedAt *time.Time `xorm:"datetime null 'verified_at' comment('商家核销时间')" json:"verified_at"`
	VerifierID uint64     `xorm:"bigint default 0 'verifier_id' comment('执行核销的店员/商家ID')" json:"verifier_id"`

	CreatedAt time.Time `xorm:"created index 'created_at'" json:"created_at"`
	UpdatedAt time.Time `xorm:"updated 'updated_at'" json:"updated_at"`
}

func (Order) TableName() string {
	return "member_orders"
}

// GetMemberOrderList 获取指定会员的订单分页列表。
// OrderQueryRequest 订单查询请求参数结构体（支持所有列作为条件过滤）
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
	IsLock        *int8  `json:"is_lock" form:"is_lock"`               // 锁定状态 (用指针支持 0:已释放, 1:已锁定 的精确过滤)
	Status        int8   `json:"status" form:"status"`                 // 订单状态 (1:待支付 2:待核销 3:已核销 4:已取消)
	PaymentMethod string `json:"payment_method" form:"payment_method"` // 支付通道 (alipay/wechat)
	VerifierID    uint64 `json:"verifier_id" form:"verifier_id"`       // 核销店员/商家ID
	StartTime     string `json:"start_time" form:"start_time"`         // 创建时间区间-开始
	EndTime       string `json:"end_time" form:"end_time"`             // 创建时间区间-结束
}

// GetMemberOrderList 获取指定会员的订单分页列表（适配全字段查询）
func GetMemberOrderList(limit, page int, search *OrderQueryRequest, order string) []*Order {
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	byOrder := "created_at DESC"
	switch order {
	case "id":
		byOrder = "id ASC"
	case "-id":
		byOrder = "id DESC"
	case "created_at":
		byOrder = "created_at ASC"
	case "-created_at":
		byOrder = "created_at DESC"
	}

	query := global.Dorm.Table("member_orders")
	if search != nil {
		// 1. 主键 ID
		if search.ID > 0 {
			query = query.And("id = ?", search.ID)
		}
		// 2. 订单号（模糊查询）
		if search.OrderNo != "" {
			query = query.And("order_no like ?", "%"+search.OrderNo+"%")
		}
		// 3. 用户ID
		if search.UserID > 0 {
			query = query.And("user_id = ?", search.UserID)
		}
		// 4. 商户ID
		if search.ShopID > 0 {
			query = query.And("shop_id = ?", search.ShopID)
		}
		// 5. 票根ID
		if search.TicketID > 0 {
			query = query.And("ticket_id = ?", search.TicketID)
		}
		// 6. 优惠券ID
		if search.CouponID > 0 {
			query = query.And("coupon_id = ?", search.CouponID)
		}
		// 7. 支付方式
		if search.PayType > 0 {
			query = query.And("pay_type = ?", search.PayType)
		}
		// 8. 锁状态（通过指针判断是否传了该参数，支持 0 和 1）
		if search.IsLock != nil {
			query = query.And("is_lock = ?", *search.IsLock)
		}
		// 9. 订单状态
		if search.Status > 0 {
			query = query.And("status = ?", search.Status)
		}
		// 10. 支付通道
		if search.PaymentMethod != "" {
			query = query.And("payment_method = ?", search.PaymentMethod)
		}
		// 11. 核销人ID
		if search.VerifierID > 0 {
			query = query.And("verifier_id = ?", search.VerifierID)
		}
		// 12. 创建时间范围（使用 >= 和 <= 代替 BETWEEN，杜绝失效）
		if search.StartTime != "" {
			query = query.And("created_at >= ?", search.StartTime)
		}
		if search.EndTime != "" {
			query = query.And("created_at <= ?", search.EndTime)
		}
	}

	list := []*Order{}
	query.OrderBy(byOrder).Limit(limit, limit*(page-1)).Find(&list)
	return list
}

// GetMemberOrderTotal 获取指定会员的订单总数（适配全字段查询）
func GetMemberOrderTotal(search *OrderQueryRequest) int64 {
	query := global.Dorm.Table("member_orders")
	if search != nil {
		if search.ID > 0 {
			query = query.And("id = ?", search.ID)
		}
		if search.OrderNo != "" {
			query = query.And("order_no like ?", "%"+search.OrderNo+"%")
		}
		if search.UserID > 0 {
			query = query.And("user_id = ?", search.UserID)
		}
		if search.ShopID > 0 {
			query = query.And("shop_id = ?", search.ShopID)
		}
		if search.TicketID > 0 {
			query = query.And("ticket_id = ?", search.TicketID)
		}
		if search.CouponID > 0 {
			query = query.And("coupon_id = ?", search.CouponID)
		}
		if search.PayType > 0 {
			query = query.And("pay_type = ?", search.PayType)
		}
		if search.IsLock != nil {
			query = query.And("is_lock = ?", *search.IsLock)
		}
		if search.Status > 0 {
			query = query.And("status = ?", search.Status)
		}
		if search.PaymentMethod != "" {
			query = query.And("payment_method = ?", search.PaymentMethod)
		}
		if search.VerifierID > 0 {
			query = query.And("verifier_id = ?", search.VerifierID)
		}
		// 12. 创建时间范围（使用 >= 和 <= 代替 BETWEEN，杜绝失效）
		if search.StartTime != "" {
			query = query.And("created_at >= ?", search.StartTime)
		}
		if search.EndTime != "" {
			query = query.And("created_at <= ?", search.EndTime)
		}
	}

	total, err := query.Count(new(Order))
	if err != nil {
		return 0
	}
	return total
}

// CreateOrderWithTicketCoupon 生成订单，并在同一事务中兑换票根优惠券。
func CreateOrderWithTicketCoupon(order *Order) (*Order, *MemberTicket, *Coupon, error) {
	session := global.Dorm.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return nil, nil, nil, fmt.Errorf("开启事务失败: %w", err)
	}
	rollback := true
	defer func() {
		if rollback {
			_ = session.Rollback()
		}
	}()

	now := time.Now()
	var ticket *MemberTicket
	if order.TicketID > 0 {
		ticket = new(MemberTicket)
		has, err := session.ID(order.TicketID).
			Where("user_id = ? AND exchange_status = 0", order.UserID).
			Get(ticket)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("查询票根异常 (ticket_id=%d, user_id=%d): %w", order.TicketID, order.UserID, err)
		}
		if !has {
			return nil, nil, nil, fmt.Errorf("票根无效或不存在 (ticket_id=%d, user_id=%d, exchange_status=0)", order.TicketID, order.UserID)
		}
	}

	var coupon *Coupon
	if order.CouponID > 0 {
		coupon = new(Coupon)
		has, err := session.ID(order.CouponID).Get(coupon)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("查询优惠券数据库异常 (coupon_id=%d): %w", order.CouponID, err)
		}

		// 1. 基础状态校验
		if !has {
			return nil, nil, nil, fmt.Errorf("优惠券不存在 (coupon_id=%d)", order.CouponID)
		}
		if coupon.ShopId != order.ShopID {
			return nil, nil, nil, fmt.Errorf("优惠券不属于当前商户 (coupon_shop_id=%d, order_shop_id=%d)", coupon.ShopId, order.ShopID)
		}
		if coupon.Status != 1 {
			return nil, nil, nil, fmt.Errorf("优惠券状态非可发放状态 (coupon_id=%d, status=%d)", order.CouponID, coupon.Status)
		}

		// 2. 检查已使用数量是否已达到发行总量
		if coupon.TotalCount > 0 && coupon.UseCount >= coupon.TotalCount {
			return nil, nil, nil, fmt.Errorf("优惠券已领完/使用完 (use_count=%d, total_count=%d)", coupon.UseCount, coupon.TotalCount)
		}

		// 3. 时间校验（兼容 Go time 零值）
		if !coupon.StartTime.IsZero() && now.Before(coupon.StartTime) {
			return nil, nil, nil, fmt.Errorf("优惠券尚未生效 (now=%s, start_time=%s)", now.Format("2006-01-02 15:04:05"), coupon.StartTime.Format("2006-01-02 15:04:05"))
		}
		if !coupon.EndTime.IsZero() && now.After(coupon.EndTime) {
			return nil, nil, nil, fmt.Errorf("优惠券已过期 (now=%s, end_time=%s)", now.Format("2006-01-02 15:04:05"), coupon.EndTime.Format("2006-01-02 15:04:05"))
		}

		// 4. 金额计算与防篡改校验
		discountAmount, err := CalculateCouponDiscount(coupon, order.OriginalAmount)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("计算优惠券折抵金额失败: %w", err)
		}
		payableAmount := order.OriginalAmount - discountAmount
		if payableAmount < 0 {
			payableAmount = 0
		}
		if !OrderAmountsMatch(order.DiscountAmount, order.PayableAmount, discountAmount, payableAmount) {
			return nil, nil, nil, fmt.Errorf("优惠券计算金额与提交金额不匹配 (提交优惠=%.2f, 计算优惠=%.2f; 提交实付=%.2f, 计算实付=%.2f)",
				order.DiscountAmount, discountAmount, order.PayableAmount, payableAmount)
		}

		// 5. 扣减/更新数据库：递增 use_count 字段
		affected, err := session.ID(order.CouponID).
			Where("shop_id = ? AND status = 1", order.ShopID).
			Incr("use_count", 1).
			Update(new(Coupon))
		if err != nil {
			return nil, nil, nil, fmt.Errorf("更新优惠券使用次数SQL异常 (coupon_id=%d): %w", order.CouponID, err)
		}
		if affected == 0 {
			return nil, nil, nil, fmt.Errorf("更新优惠券使用次数失败 (未匹配到符合条件记录, coupon_id=%d, shop_id=%d)", order.CouponID, order.ShopID)
		}

		order.DiscountAmount = discountAmount
		order.PayableAmount = payableAmount
	} else {
		if !OrderAmountsMatch(order.DiscountAmount, order.PayableAmount, 0, order.OriginalAmount) {
			return nil, nil, nil, fmt.Errorf("未选择优惠券时金额不匹配 (提交优惠=%.2f, 提交实付=%.2f, 原价=%.2f)",
				order.DiscountAmount, order.PayableAmount, order.OriginalAmount)
		}
		order.DiscountAmount = 0
		if order.PayableAmount < 0 {
			order.PayableAmount = 0
		}
	}

	// 保留 Handler 传入的状态（兼容 PayType=2 当面付进入状态 2）
	if order.Status == 0 {
		order.Status = 1
	}
	order.CreatedAt = now
	if _, err := session.Insert(order); err != nil {
		return nil, nil, nil, fmt.Errorf("插入订单表记录失败: %w", err)
	}

	if ticket != nil {
		affected, err := session.ID(order.TicketID).
			Where("user_id = ? AND exchange_status = 0", order.UserID).
			Cols("exchange_status", "coupon_id", "exchanged_at").
			Update(&MemberTicket{
				ExchangeStatus: 1,
				CouponID:       order.CouponID,
				ExchangedAt:    &now,
			})
		if err != nil {
			return nil, nil, nil, fmt.Errorf("更新票根兑换状态SQL异常 (ticket_id=%d): %w", order.TicketID, err)
		}
		if affected == 0 {
			return nil, nil, nil, fmt.Errorf("更新票根兑换状态失败 (票根可能已被抢先兑换, ticket_id=%d)", order.TicketID)
		}
	}

	if err := session.Commit(); err != nil {
		return nil, nil, nil, fmt.Errorf("事务提交(Commit)失败: %w", err)
	}
	rollback = false

	if ticket != nil {
		ticket.ExchangeStatus = 1
		ticket.CouponID = order.CouponID
		ticket.ExchangedAt = &now
	}
	if coupon != nil {
		coupon.UseCount++ // 匹配结构体字段名
	}
	return order, ticket, coupon, nil
}

func NewOrderNo(userID uint64) string {
	return fmt.Sprintf("MO%d%d", time.Now().UnixNano(), userID)
}

// OrderAmountsMatch 按金额分精度比较前端金额与后端计算结果。
func OrderAmountsMatch(discount, payable, expectedDiscount, expectedPayable float64) bool {
	const tolerance = 0.01
	return absFloat(discount-expectedDiscount) < tolerance && absFloat(payable-expectedPayable) < tolerance
}

// ResolveOrderAmounts 兼容前端传原价或已优惠后的实付金额。
// 当 amount 与 payable_amount 相等时，amount 视为前端已扣券后的金额。
func ResolveOrderAmounts(amount, discount, payable float64) (original, final float64) {
	if discount > 0 && absFloat(amount-payable) < 0.01 {
		return amount + discount, payable
	}
	return amount, payable
}

func absFloat(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

// GetOrderByNo 根据订单号查询订单详情
func GetOrderByNo(orderNo string) (*Order, error) {
	if orderNo == "" {
		return nil, errors.New("订单号不能为空")
	}

	var order Order
	// 使用 Xorm 根据 order_no 字段进行精确查询
	has, err := global.Dorm.Table("member_orders").Where("order_no = ?", orderNo).Get(&order)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil // 查无此订单时返回 nil
	}

	return &order, nil
}

// VerifyOrderTransaction 商家核销订单事务处理
func VerifyOrderTransaction(order *Order, verifierID uint64) error {
	session := global.Dorm.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	// 1. 更新订单状态：标记为已核销，解锁 IsLock=0
	now := time.Now()
	order.Status = 3 // 3: 已核销/已完成
	order.IsLock = 0 // 解除锁定状态
	order.VerifiedAt = &now
	order.VerifierID = verifierID

	_, err := session.ID(order.ID).Cols("status", "is_lock", "verified_at", "verifier_id", "updated_at").Update(order)
	if err != nil {
		session.Rollback()
		return err
	}

	// 2. 将对应的票根 (Ticket) 状态更新为已核销 (假设状态 2 为已核销)
	if order.TicketID > 0 {
		_, err = session.Table("member_tickets").Where("id = ?", order.TicketID).Update(map[string]interface{}{
			"exchange_status": 2,
			"exchanged_at":    now,
		})
		if err != nil {
			session.Rollback()
			return err
		}
	}

	// 3. 将对应的优惠券 (Coupon) 状态更新为已使用 (假设状态 2 为已使用)
	if order.CouponID > 0 {
		_, err = session.Table("pgh5_coupon").Where("id = ?", order.CouponID).Update(map[string]interface{}{
			"status": 2,
			// "updated_at": now,
		})
		if err != nil {
			session.Rollback()
			return err
		}
	}

	return session.Commit()
}

// UpdateOrderPayType 根据订单号更新支付方式 (pay_type)
// payType: 0:未选择 1:在线支付 2:到店/当面核销支付
func UpdateOrderPayType(orderNo string, payType int8) error {
	if orderNo == "" {
		return errors.New("订单号不能为空")
	}

	// 限制只能修改“待支付/待处理”(Status = 1) 状态下的订单支付方式（可根据实际业务需要选择是否加此校验）
	// 执行更新操作
	affected, err := global.Dorm.Table(new(Order)).
		Where("order_no = ? AND status = 1", orderNo).
		Cols("pay_type", "updated_at").
		Update(&Order{
			PayType:   payType,
			UpdatedAt: time.Now(),
		})

	if err != nil {
		return fmt.Errorf("更新订单支付方式数据库异常 (order_no=%s): %w", orderNo, err)
	}

	if affected == 0 {
		return errors.New("更新失败：订单不存在、已支付、已取消或支付方式未发生改变")
	}

	return nil
}

// CreateOrderWithItemsTransaction 数据库事务：创建商品订单与对应的核销明细
func CreateOrderWithItemsTransaction(order *Order, items []*OrderItem) (*Order, []*OrderItem, error) {
	session := global.Dorm.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return nil, nil, err
	}

	// 1. 插入订单主表
	if _, err := session.Insert(order); err != nil {
		session.Rollback()
		return nil, nil, err
	}

	// 2. 绑定新的 OrderId 并插入商品明细表
	for _, item := range items {
		item.OrderId = int64(order.ID)
	}

	if len(items) > 0 {
		if _, err := session.Insert(&items); err != nil {
			session.Rollback()
			return nil, nil, err
		}
	}

	// 3. 提交事务
	if err := session.Commit(); err != nil {
		return nil, nil, err
	}

	return order, items, nil
}

// GenerateVerifyCode 生成不重复的随机数字/字母核销码
func GenerateVerifyCode() string {
	return fmt.Sprintf("%d%04d", time.Now().UnixNano()%100000000, rand.Intn(10000))
}
