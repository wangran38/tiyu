package shop

import (
	"time"

	"tiyu/global"
)

// GoodsDailyCalendar 日历价格库存表（酒店/景区等存在"平周末价差"的业态使用）。
//
// 用法：
//   - 当 GoodsProduct.StockType=2（每日/场次动态库存）时启用本表。
//   - 按 (product_id, sku_id, date) 三元组唯一：一个 SKU 某一天的 价格/库存。
//   - sku_id=0 表示商品维度的日历（无规格商品）；>0 表示具体 SKU 日历。
//   - 未命中某天的记录时，回落到 GoodsProduct.SellingPrice / GoodsSku.Stock。
type GoodsDailyCalendar struct {
	Id        int64     `json:"id" xorm:"pk autoincr comment('主键ID')"`
	ProductId int64     `json:"product_id" xorm:"unique(name) notnull index comment('关联团购产品ID')"`
	SkuId     int64     `json:"sku_id" xorm:"unique(name) notnull default 0 index comment('关联SKU ID，0表示商品维度')"`
	Date      time.Time `json:"date" xorm:"unique(name) date notnull index comment('日期，仅日期部分有效')"`
	Price     float64   `json:"price" xorm:"decimal(10,2) notnull default 0 comment('当日售价')"`
	Stock     int       `json:"stock" xorm:"notnull default 0 comment('当日库存')"`
	Status    int       `json:"status" xorm:"notnull default 1 comment('状态: 1-可售, 0-停售/闭店/闭园')"`
	Created   time.Time `json:"createtime" xorm:"created int comment('创建时间')"`
	Updated   time.Time `json:"updatetime" xorm:"updated int comment('更新时间')"`
}

func (c *GoodsDailyCalendar) TableName() string {
	return "goods_daily_calendar"
}

// GetGoodsDailyCalendarList 查询某商品（可指定 SKU）的日历记录。
// startDate/endDate 为空时不过滤日期范围。
func GetGoodsDailyCalendarList(productId, skuId int64, startDate, endDate string) ([]*GoodsDailyCalendar, error) {
	list := []*GoodsDailyCalendar{}
	session := global.Dorm.Where("product_id = ?", productId)
	if skuId > 0 {
		session = session.And("sku_id = ?", skuId)
	}
	if startDate != "" {
		session = session.And("date >= ?", startDate)
	}
	if endDate != "" {
		session = session.And("date <= ?", endDate)
	}
	err := session.Asc("date").Find(&list)
	return list, err
}

// AddGoodsDailyCalendar 批量新增日历记录（编辑时用：先删后插）
func AddGoodsDailyCalendar(items []*GoodsDailyCalendar) error {
	if len(items) == 0 {
		return nil
	}
	_, err := global.Dorm.Insert(items)
	return err
}

// DeleteGoodsDailyCalendarByProductID 按产品ID清空日历（编辑时用：先删后插）
func DeleteGoodsDailyCalendarByProductID(productId int64) error {
	_, err := global.Dorm.Where("product_id = ?", productId).Delete(new(GoodsDailyCalendar))
	return err
}
