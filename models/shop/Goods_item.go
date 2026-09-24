package shop

import "tiyu/global"

// GoodsItem 产品明细表（如：套餐包含菜品列表、门票包含项目、酒店赠送权益）
type GoodsItem struct {
	Id            int64   `json:"id" xorm:"pk autoincr comment('明细ID')"`
	ProductId     int64   `json:"product_id" xorm:"index notnull comment('关联团购产品ID')"`
	GroupName     string  `json:"group_name" xorm:"varchar(64) notnull comment('分组类别，如：热菜/主食/赠送项目/费用包含')"`
	ItemName      string  `json:"item_name" xorm:"varchar(128) notnull comment('细项名称，如：清蒸石斑鱼/景区门票/延迟退房')"`
	Quantity      int     `json:"quantity" xorm:"notnull default 1 comment('数量')"`
	Unit          string  `json:"unit" xorm:"varchar(16) default '份' comment('单位: 份/张/位/次')"`
	Price         float64 `json:"price" xorm:"decimal(10,2) default 0.00 comment('单品参考价值/价值额')"`
	IsChoice      int     `json:"is_choice" xorm:"notnull default 0 comment('是否多选一: 0-必选项, 1-多选一选项')"`
	ChoiceGroupId int     `json:"choice_group_id" xorm:"default 0 comment('多选一组ID，用于前端将多个可选项分组')"`
}

func (i *GoodsItem) TableName() string {
	return "goods_item"
}

// GetGoodsItemListByProductID 获取指定产品对应的包含明细项
func GetGoodsItemListByProductID(productId int64) ([]*GoodsItem, error) {
	items := []*GoodsItem{}
	err := global.Dorm.Where("product_id = ?", productId).Find(&items)
	return items, err
}

// AddGoodsItem 批量新增套餐/费用明细
func AddGoodsItem(items []*GoodsItem) error {
	if len(items) == 0 {
		return nil
	}
	_, err := global.Dorm.Insert(items)
	return err
}

// DeleteGoodsItemByProductID 按产品ID清空该产品的所有明细项（编辑时用：先删后插）
func DeleteGoodsItemByProductID(productId int64) error {
	_, err := global.Dorm.Where("product_id = ?", productId).Delete(new(GoodsItem))
	return err
}

// EditGoodsItem 修改单条明细
func EditGoodsItem(item *GoodsItem) error {
	_, err := global.Dorm.ID(item.Id).Update(item)
	return err
}
