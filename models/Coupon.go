package models

import (
	"errors"
	"time"
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
	has, err := Dorm.ID(id).Get(coupon)
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
	query := Dorm.Table("pgh5_coupon")

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
		// 如果按某个特定分类过滤，可以使用 FIND_IN_SET 来匹配逗号分隔的多分类字段
		if search.Categories != "" {
			query = query.And("find_in_set(?, categories)", search.Categories)
		}
		if search.Status > 0 {
			query = query.And("status = ?", search.Status)
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
	query := Dorm.Table("pgh5_coupon")

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
	}

	total, err := query.Count(a)
	if err != nil {
		return 0
	}
	return total
}

// 新增
func AddCoupon(a *Coupon) error {
	_, err := Dorm.Insert(a)
	return err
}

// 修改
func EditCoupon(a *Coupon) error {
	_, err := Dorm.ID(a.Id).Update(a)
	return err
}

// 删除
func DelCoupon(id int64) int {
	a := new(Coupon)
	outnum, _ := Dorm.ID(id).Delete(a)
	return int(outnum)
}

// 修改（带商户归属校验，防止越权）
func EditCouponByShop(a *Coupon) int64 {
	// 使用 ID 和 ShopId 联合定位，确保只能修改自己店铺的券
	affected, err := Dorm.ID(a.Id).Where("shop_id = ?", a.ShopId).Update(a)
	if err != nil {
		return 0
	}
	return affected
}
