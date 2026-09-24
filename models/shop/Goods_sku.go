package shop

import (
	"time"
	"tiyu/global"
)

// GoodsSku 团购产品规格表（如：大床房/双床房、工作日餐/周末餐、看台区/内场区）
type GoodsSku struct {
	Id            int64     `json:"id" xorm:"pk autoincr comment('SKU ID')"`
	ProductId     int64     `json:"product_id" xorm:"index notnull comment('关联团购产品ID')"`
	SkuName       string    `json:"sku_name" xorm:"varchar(64) notnull comment('SKU规格名称，如：双人豪华餐/大床房(含双早)')"`
	OriginalPrice float64   `json:"original_price" xorm:"decimal(10,2) notnull comment('该规格划线原价')"`
	Price         float64   `json:"price" xorm:"decimal(10,2) notnull comment('该规格实际售买价')"`
	Stock         int       `json:"stock" xorm:"notnull default 0 comment('当前SKU剩余库存')"`
	SkuCode       string    `json:"sku_code" xorm:"varchar(64) default '' comment('商家内部SKU编码/对接第三方系统编码')"`
	SpecData      string    `json:"spec_data" xorm:"json comment('扩展属性 JSON，如: {\"bed_type\":\"大床\",\"breakfast\":\"双早\"}')"`
	Status        int       `json:"status" xorm:"notnull default 1 comment('状态: 1-启用, 0-禁用')"`
	Created       time.Time `json:"createtime" xorm:"created int comment('创建时间')"`
}

func (s *GoodsSku) TableName() string {
	return "goods_sku"
}

// GetGoodsSkuListByProductID 获取指定产品下的所有有效 SKU
func GetGoodsSkuListByProductID(productId int64) ([]*GoodsSku, error) {
	skus := []*GoodsSku{}
	err := global.Dorm.Where("product_id = ? AND status = 1", productId).Find(&skus)
	return skus, err
}

// GetGoodsSkuRawListByProductID 获取指定产品下的全部 SKU 原始记录（不过滤 status）。
// 用于编辑回显：禁用中的 SKU 也要原样返回，避免回显丢数据。
func GetGoodsSkuRawListByProductID(productId int64) ([]*GoodsSku, error) {
	skus := []*GoodsSku{}
	err := global.Dorm.Where("product_id = ?", productId).Find(&skus)
	return skus, err
}

// AddGoodsSku 新增 SKU 规格
func AddGoodsSku(sku *GoodsSku) error {
	_, err := global.Dorm.Insert(sku)
	return err
}

// EditGoodsSku 修改 SKU 规格
func EditGoodsSku(sku *GoodsSku) error {
	_, err := global.Dorm.ID(sku.Id).Update(sku)
	return err
}

// DeleteGoodsSkuByProductID 按产品ID清空该产品的所有 SKU 规格（编辑时用：先删后插）
func DeleteGoodsSkuByProductID(productId int64) error {
	_, err := global.Dorm.Where("product_id = ?", productId).Delete(new(GoodsSku))
	return err
}
