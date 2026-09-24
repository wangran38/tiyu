package models

import (
	"errors"
	"time"
	"tiyu/global"
)

var (
	ErrOrderItemNotFound = errors.New("订单明细不存在")
	ErrOrderItemVerified = errors.New("该券码已被核销或已失效")
)

// OrderItem 订单商品/团购核销明细表
type OrderItem struct {
	Id          int64      `json:"id" xorm:"pk autoincr bigint 'id'"`
	OrderId     int64      `json:"order_id" xorm:"bigint notnull index 'order_id' comment('关联订单ID')"`
	OrderNo     string     `json:"order_no" xorm:"varchar(40) notnull index 'order_no' comment('订单号')"`
	ShopId      int64      `json:"shop_id" xorm:"bigint notnull index 'shop_id' comment('所属商家/店铺ID')"`
	UserId      int64      `json:"user_id" xorm:"bigint notnull index 'user_id' comment('买家用户ID')"`
	ProductId   int64      `json:"product_id" xorm:"bigint notnull comment('团购商品/套餐ID')"`
	ProductName string     `json:"product_name" xorm:"varchar(128) notnull comment('商品/套餐名称')"`
	Price       float64    `json:"price" xorm:"decimal(10,2) notnull default 0.00 comment('商品单价')"`
	Quantity    int        `json:"quantity" xorm:"int notnull default 1 comment('购买数量')"`
	VerifyCode  string     `json:"verify_code" xorm:"varchar(32) default '' unique 'verify_code' comment('团购/兑换独立核销码')"`
	Status      int        `json:"status" xorm:"tinyint notnull default 1 index 'status' comment('状态: 1-待核销/待使用, 2-已核销, 3-已退款, 4-已作废/过期')"`
	VerifiedAt  *time.Time `json:"verified_at" xorm:"datetime null comment('商家核销时间')"`
	VerifierId  int64      `json:"verifier_id" xorm:"bigint default 0 comment('核销操作员/商家ID')"`
	Created     time.Time  `json:"createtime" xorm:"created int"`
	Updated     time.Time  `json:"updatetime" xorm:"updated int"`
}

func (m *OrderItem) TableName() string {
	return "order_item"
}

// GetOrderItemByID 根据ID获取记录
func GetOrderItemByID(id uint64) (*OrderItem, error) {
	item := new(OrderItem)
	has, err := global.Dorm.ID(id).Get(item)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return item, nil
}

// GetOrderItemByVerifyCode 根据核销码查找（常用于商家扫码核销场景）
func GetOrderItemByVerifyCode(code string) (*OrderItem, error) {
	if code == "" {
		return nil, errors.New("核销码不能为空")
	}
	item := new(OrderItem)
	has, err := global.Dorm.Where("verify_code = ?", code).Get(item)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return item, nil
}

// GetOrderItemsByOrderID 获取指定订单下的所有商品明细列表
func GetOrderItemsByOrderID(orderID int64) ([]*OrderItem, error) {
	listdata := []*OrderItem{}
	err := global.Dorm.Table("order_item").
		Where("order_id = ?", orderID).
		Find(&listdata)
	return listdata, err
}

// GetOrderItemList 分页列表（使用结构体指针作为 search 条件）
func GetOrderItemList(limit int, page int, search *OrderItem, order string) []*OrderItem {
	offset := page - 1
	if offset < 0 {
		offset = 0
	}
	byorder := "id DESC"
	if order != "" {
		byorder = order
	}

	listdata := []*OrderItem{}
	query := global.Dorm.Table("order_item")

	if search != nil {
		if search.OrderNo != "" {
			query = query.Where("order_no = ?", search.OrderNo)
		}
		if search.OrderId > 0 {
			query = query.And("order_id = ?", search.OrderId)
		}
		if search.ShopId > 0 {
			query = query.And("shop_id = ?", search.ShopId)
		}
		if search.UserId > 0 {
			query = query.And("user_id = ?", search.UserId)
		}
		if search.ProductId > 0 {
			query = query.And("product_id = ?", search.ProductId)
		}
		if search.VerifyCode != "" {
			query = query.And("verify_code = ?", search.VerifyCode)
		}
		if search.Status > 0 {
			query = query.And("status = ?", search.Status)
		}
		if search.ProductName != "" {
			query = query.And("product_name like ?", "%"+search.ProductName+"%")
		}
	}

	query.OrderBy(byorder).
		Limit(limit, limit*offset).
		Find(&listdata)
	return listdata
}

// GetOrderItemTotal 获取符合条件的总条数
func GetOrderItemTotal(search *OrderItem) int64 {
	item := new(OrderItem)
	query := global.Dorm.Table("order_item")

	if search != nil {
		if search.OrderNo != "" {
			query = query.Where("order_no = ?", search.OrderNo)
		}
		if search.OrderId > 0 {
			query = query.And("order_id = ?", search.OrderId)
		}
		if search.ShopId > 0 {
			query = query.And("shop_id = ?", search.ShopId)
		}
		if search.UserId > 0 {
			query = query.And("user_id = ?", search.UserId)
		}
		if search.ProductId > 0 {
			query = query.And("product_id = ?", search.ProductId)
		}
		if search.VerifyCode != "" {
			query = query.And("verify_code = ?", search.VerifyCode)
		}
		if search.Status > 0 {
			query = query.And("status = ?", search.Status)
		}
		if search.ProductName != "" {
			query = query.And("product_name like ?", "%"+search.ProductName+"%")
		}
	}

	total, err := query.Count(item)
	if err != nil {
		return 0
	}
	return total
}

// AddOrderItem 单条新增
func AddOrderItem(item *OrderItem) error {
	_, err := global.Dorm.Insert(item)
	return err
}

// AddOrderItems 批量插入明细（常用于购买多件/多张团购券）
func AddOrderItems(items []*OrderItem) error {
	if len(items) == 0 {
		return nil
	}
	_, err := global.Dorm.Insert(&items)
	return err
}

// EditOrderItem 修改
func EditOrderItem(item *OrderItem) error {
	_, err := global.Dorm.ID(item.Id).Update(item)
	return err
}

// EditOrderItemByShop 修改（带商家归属校验，防止越权操作）
func EditOrderItemByShop(item *OrderItem) int64 {
	affected, err := global.Dorm.ID(item.Id).Where("shop_id = ?", item.ShopId).Update(item)
	if err != nil {
		return 0
	}
	return affected
}

// DelOrderItem 删除
func DelOrderItem(id int64) int {
	item := new(OrderItem)
	outnum, _ := global.Dorm.ID(id).Delete(item)
	return int(outnum)
}

// VerifyOrderItem 执行团购核销操作（带状态控制与商家安全防护）
func VerifyOrderItem(verifyCode string, shopID int64, verifierID int64) error {
	now := time.Now()
	affected, err := global.Dorm.Table("order_item").
		Where("verify_code = ? AND shop_id = ? AND status = 1", verifyCode, shopID).
		Update(map[string]interface{}{
			"status":      2, // 标记为已核销
			"verified_at": now,
			"verifier_id": verifierID,
			"updatetime":  now.Unix(),
		})

	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrOrderItemVerified
	}
	return nil
}
