package models

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrOrderCouponInvalid = errors.New("coupon does not belong to shop or is unavailable")
	ErrOrderTicketInvalid = errors.New("ticket not found or already exchanged")
	ErrOrderAmountInvalid = errors.New("order amount does not match coupon rules")
)

// Order 会员在商户消费并使用票根优惠券生成的订单。
type Order struct {
	ID               uint64     `xorm:"pk autoincr bigint 'id'" json:"id"`
	OrderNo          string     `xorm:"varchar(40) notnull unique 'order_no'" json:"order_no"`
	UserID           uint64     `xorm:"bigint notnull index 'user_id'" json:"user_id"`
	ShopID           int64      `xorm:"bigint notnull index 'shop_id'" json:"shop_id"`
	TicketID         uint64     `xorm:"bigint notnull index 'ticket_id'" json:"ticket_id"`
	CouponID         uint64     `xorm:"bigint notnull index 'coupon_id'" json:"coupon_id"`
	OriginalAmount   float64    `xorm:"decimal(10,2) notnull 'original_amount'" json:"original_amount"`
	DiscountAmount   float64    `xorm:"decimal(10,2) notnull default 0.00 'discount_amount'" json:"discount_amount"`
	PayableAmount    float64    `xorm:"decimal(10,2) notnull 'payable_amount'" json:"payable_amount"`
	Status           int8       `xorm:"tinyint notnull default 1 index 'status' comment('1:待支付 2:已支付 3:已取消')" json:"status"`
	PaymentMethod    string     `xorm:"varchar(32) default '' 'payment_method' comment('支付方式: alipay/wechat/unionpay/digital_rmb等')" json:"payment_method"`
	PaidAt           *time.Time `xorm:"datetime null 'paid_at' comment('支付成功回调时间')" json:"paid_at"`
	PaymentNotifyURL string     `xorm:"varchar(512) default '' 'payment_notify_url' comment('支付回调通知URL')" json:"payment_notify_url"`
	CreatedAt        time.Time  `xorm:"created index 'created_at'" json:"created_at"`
	UpdatedAt        time.Time  `xorm:"updated 'updated_at'" json:"updated_at"`
}

func (Order) TableName() string {
	return "member_orders"
}

// GetMemberOrderList 获取指定会员的订单分页列表。
func GetMemberOrderList(limit, page int, search *Order, order string) []*Order {
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
	}

	query := Dorm.Table("member_orders")
	if search != nil {
		if search.UserID > 0 {
			query = query.And("user_id = ?", search.UserID)
		}
		if search.ShopID > 0 {
			query = query.And("shop_id = ?", search.ShopID)
		}
		if search.Status > 0 {
			query = query.And("status = ?", search.Status)
		}
		if search.OrderNo != "" {
			query = query.And("order_no like ?", "%"+search.OrderNo+"%")
		}
	}

	list := []*Order{}
	query.OrderBy(byOrder).Limit(limit, limit*(page-1)).Find(&list)
	return list
}

// GetMemberOrderTotal 获取指定会员的订单总数。
func GetMemberOrderTotal(search *Order) int64 {
	query := Dorm.Table("member_orders")
	if search != nil {
		if search.UserID > 0 {
			query = query.And("user_id = ?", search.UserID)
		}
		if search.ShopID > 0 {
			query = query.And("shop_id = ?", search.ShopID)
		}
		if search.Status > 0 {
			query = query.And("status = ?", search.Status)
		}
		if search.OrderNo != "" {
			query = query.And("order_no like ?", "%"+search.OrderNo+"%")
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
	session := Dorm.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return nil, nil, nil, err
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
			return nil, nil, nil, err
		}
		if !has {
			return nil, nil, nil, ErrOrderTicketInvalid
		}
	}

	var coupon *Coupon
	if order.CouponID > 0 {
		coupon = new(Coupon)
		has, err := session.ID(order.CouponID).Get(coupon)
		if err != nil {
			return nil, nil, nil, err
		}
		if !has || coupon.ShopId != order.ShopID || coupon.Status != 1 || coupon.ReceiveCount >= coupon.TotalCount || now.Before(coupon.StartTime) || now.After(coupon.EndTime) {
			return nil, nil, nil, ErrOrderCouponInvalid
		}
		discountAmount, err := CalculateCouponDiscount(coupon, order.OriginalAmount)
		if err != nil {
			return nil, nil, nil, ErrOrderAmountInvalid
		}
		payableAmount := order.OriginalAmount - discountAmount
		if payableAmount < 0 {
			payableAmount = 0
		}
		if !OrderAmountsMatch(order.DiscountAmount, order.PayableAmount, discountAmount, payableAmount) {
			return nil, nil, nil, ErrOrderAmountInvalid
		}

		affected, err := session.ID(order.CouponID).
			Where("shop_id = ? AND status = 1 AND receive_count < total_count", order.ShopID).
			Incr("receive_count", 1).
			Update(new(Coupon))
		if err != nil {
			return nil, nil, nil, err
		}
		if affected == 0 {
			return nil, nil, nil, ErrOrderCouponInvalid
		}

		order.DiscountAmount = discountAmount
		order.PayableAmount = payableAmount
	} else {
		if !OrderAmountsMatch(order.DiscountAmount, order.PayableAmount, 0, order.OriginalAmount) {
			return nil, nil, nil, ErrOrderAmountInvalid
		}
		order.DiscountAmount = 0
		if order.PayableAmount < 0 {
			order.PayableAmount = 0
		}
	}

	order.Status = 1
	order.CreatedAt = now
	if _, err := session.Insert(order); err != nil {
		return nil, nil, nil, err
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
			return nil, nil, nil, err
		}
		if affected == 0 {
			return nil, nil, nil, ErrOrderTicketInvalid
		}
	}

	if err := session.Commit(); err != nil {
		return nil, nil, nil, err
	}
	rollback = false
	if ticket != nil {
		ticket.ExchangeStatus = 1
		ticket.CouponID = order.CouponID
		ticket.ExchangedAt = &now
	}
	if coupon != nil {
		coupon.ReceiveCount++
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
