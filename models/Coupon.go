package models

import (
	"errors"
	"time"
	"tiyu/global"
)

var ErrCouponMinPoint = errors.New("订单金额未达到优惠券使用门槛")

// Coupon 优惠券/权益定义表（商家发布）
type Coupon struct {
	Id             int64     `json:"id"`
	ShopId         int64     `json:"shop_id" xorm:"index notnull comment('所属店铺ID')"`
	UserId         int64     `json:"user_id" xorm:"index notnull comment('发布该优惠券的商户/店长用户ID')"`
	Title          string    `json:"title" xorm:"varchar(100) notnull comment('优惠券标题')"`
	Categories     string    `json:"categories" xorm:"varchar(100) notnull comment('多业态分类ID组合，例如: 1,3,4 代表支持景区、餐饮、零售')"`
	CouponType     int       `json:"coupon_type" xorm:"notnull comment('形式: 1-满减券, 2-折扣券, 3-首道门票折扣, 4-专属票价')"`
	DiscountAmount float64   `json:"discount_amount" xorm:"decimal(10,2) notnull comment('优惠面额或折扣率')"`
	MinPoint       float64   `json:"min_point" xorm:"decimal(10,2) not null default 0.00 comment('使用门槛金额 0代表无门槛')"`
	TotalCount     int       `json:"total_count" xorm:"not not null default 0 comment('发放总数量')"`
	ReceiveCount   int       `json:"receive_count" xorm:"not not null default 0 comment('已领取数量')"`
	UseCount       int       `json:"use_count" xorm:"not not null default 0 comment('已核销使用数量')"`
	StartTime      time.Time `json:"starttime" xorm:"notnull comment('生效时间')"`
	EndTime        time.Time `json:"endtime" xorm:"notnull comment('过期时间')"`
	Status         int       `json:"status" xorm:"not not null default 1 comment('状态: 1-正常发放中, 2-已抢光, 3-已下架/过期')"`
	Created        time.Time `json:"createtime" xorm:"created int"`
	Updated        time.Time `json:"updatetime" xorm:"updated int"`
}

func (a *Coupon) TableName() string {
	return "pgh5_coupon"
}

func GetCouponByID(id uint64) (*Coupon, error) {
	coupon := new(Coupon)
	has, err := global.Dorm.ID(id).Get(coupon)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return coupon, nil
}

// CalculateCouponDiscount 根据优惠券规则计算优惠金额，订单金额由服务端确定。
func CalculateCouponDiscount(coupon *Coupon, amount float64) (float64, error) {
	if amount <= 0 || amount < coupon.MinPoint {
		return 0, ErrCouponMinPoint
	}

	discount := coupon.DiscountAmount
	if coupon.CouponType == 2 {
		rate := coupon.DiscountAmount
		if rate > 1 {
			rate /= 10
		}
		if rate < 0 || rate > 1 {
			return 0, errors.New("优惠券折扣率无效")
		}
		discount = amount * (1 - rate)
	}
	if discount < 0 {
		discount = 0
	}
	if discount > amount {
		discount = amount
	}
	return discount, nil
}

// 分页列表（使用结构体指针作为 search 条件）
// 分页列表（使用结构体指针作为 search 条件）
func GetCouponList(limit int, page int, search *Coupon, order string) []*Coupon {
	offset := page - 1
	if offset < 0 {
		offset = 0
	}
	byorder := "id DESC"
	if order != "" {
		byorder = order
	}

	listdata := []*Coupon{}
	query := global.Dorm.Table("pgh5_coupon")

	if search != nil {
		if search.Title != "" {
			query = query.Where("title like ?", "%"+search.Title+"%")
		}
		if search.ShopId > 0 {
			query = query.And("shop_id = ?", search.ShopId)
		}
		if search.UserId > 0 {
			query = query.And("user_id = ?", search.UserId)
		}
		if search.Categories != "" {
			query = query.And("find_in_set(?, categories)", search.Categories)
		}
		if search.Status > 0 {
			query = query.And("status = ?", search.Status)
		}
		// 判断 EndTime 是否传值（非零值时生效）
		if !search.EndTime.IsZero() {
			query = query.And("end_time >= ?", search.EndTime)
		}
	}

	query.OrderBy(byorder).
		Limit(limit, limit*offset).
		Find(&listdata)
	return listdata
}

// 获取总数
func GetCouponTotal(search *Coupon) int64 {
	a := new(Coupon)
	query := global.Dorm.Table("pgh5_coupon")

	if search != nil {
		if search.Title != "" {
			query = query.Where("title like ?", "%"+search.Title+"%")
		}
		if search.ShopId > 0 {
			query = query.And("shop_id = ?", search.ShopId)
		}
		if search.UserId > 0 {
			query = query.And("user_id = ?", search.UserId)
		}
		if search.Categories != "" {
			query = query.And("find_in_set(?, categories)", search.Categories)
		}
		if search.Status > 0 {
			query = query.And("status = ?", search.Status)
		}
		// 同样的 EndTime 过滤条件
		if !search.EndTime.IsZero() {
			query = query.And("end_time >= ?", search.EndTime)
		}
	}

	total, err := query.Count(a)
	if err != nil {
		return 0
	}
	return total
}

// 新增
func AddCoupon(a *Coupon) error {
	_, err := global.Dorm.Insert(a)
	return err
}

// 修改
func EditCoupon(a *Coupon) error {
	_, err := global.Dorm.ID(a.Id).Update(a)
	return err
}

// 删除
func DelCoupon(id int64) int {
	a := new(Coupon)
	outnum, _ := global.Dorm.ID(id).Delete(a)
	return int(outnum)
}

// 修改（带商户归属校验，防止越权）
func EditCouponByShop(a *Coupon) int64 {
	// 使用 ID 和 ShopId 联合定位，确保只能修改自己店铺的券
	affected, err := global.Dorm.ID(a.Id).Where("shop_id = ?", a.ShopId).Update(a)
	if err != nil {
		return 0
	}
	return affected
}

// MemberCouponRecord 会员已领取的优惠券记录（连表票根，展示兑换使用的票根信息）
type MemberCouponRecord struct {
	// 兑换所用票根信息 (member_tickets)
	TicketID       uint64     `xorm:"ticket_id" json:"ticket_id"`
	ChannelType    string     `xorm:"channel_type" json:"channel_type"`
	TicketCategory string     `xorm:"ticket_category" json:"ticket_category"`
	TemplateID     uint64     `xorm:"template_id" json:"template_id"`
	UserImageURL   string     `xorm:"user_image_url" json:"user_image_url"`
	TicketTitle    string     `xorm:"ticket_title" json:"ticket_title"`
	HolderName     string     `xorm:"holder_name" json:"holder_name"`
	EventDate      string     `xorm:"event_date" json:"event_date"`
	Seat           string     `xorm:"seat" json:"seat"`
	TicketAmount   float64    `xorm:"ticket_amount" json:"ticket_amount"`
	IsHandwritten  bool       `xorm:"is_handwritten" json:"is_handwritten"`
	ExchangedAt    *time.Time `xorm:"exchanged_at" json:"exchanged_at"`
	// 关联优惠券信息 (pgh5_coupon，通过 c.* 扩展)
	Coupon *Coupon `xorm:"extends" json:"coupon"`
}

// memberCouponColumns 连表查询需展示的列。因 pgh5_coupon 与 member_tickets 均含 title 等重名列，
// XORM 无法区分同名列，故仅对票根的重名列使用别名，优惠券字段通过 c.* 全部展开。
const memberCouponColumns = `
	c.*,
	mt.id AS ticket_id, mt.channel_type, mt.ticket_category, mt.template_id,
	mt.user_image_url, mt.title AS ticket_title, mt.holder_name, mt.event_date,
	mt.seat, mt.amount AS ticket_amount, mt.is_handwritten, mt.exchanged_at`

// GetMemberCouponList 获取指定会员已领取（兑换成功）的优惠券分页列表，以优惠券为主表 LEFT JOIN 票根，展示兑换所用票根。按兑换时间倒序。
func GetMemberCouponList(limit, page int, userID uint64) []*MemberCouponRecord {
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	list := []*MemberCouponRecord{}
	global.Dorm.Table("pgh5_coupon c").
		Select(memberCouponColumns).
		Join("LEFT JOIN", "member_tickets mt", "mt.coupon_id = c.id AND mt.exchange_status = 1").
		Where("c.id IN (SELECT coupon_id FROM member_tickets WHERE user_id = ? AND exchange_status = 1 AND coupon_id > 0)", userID).
		OrderBy("mt.exchanged_at DESC").
		Limit(limit, offset).
		Find(&list)
	return list
}

// GetMemberCouponTotal 获取指定会员已领取优惠券的总数。
func GetMemberCouponTotal(userID uint64) int64 {
	total, _ := global.Dorm.Table("member_tickets").
		Where("user_id = ?", userID).
		And("exchange_status = 1").
		And("coupon_id > 0").
		Count(new(MemberTicket))
	return total
}
