package shop

// GoodsProductDetail 产品完整详情组合结构体
type GoodsProductDetail struct {
	Product        *GoodsProduct        `json:"product"`
	Skus           []*GoodsSku          `json:"skus"`
	Items          []*GoodsItem         `json:"items"`
	Rule           *GoodsRule           `json:"rule"`
	TicketDiscount *GoodsTicketDiscount `json:"ticket_discount,omitempty"`
}

// GetGoodsProductDetailByID 聚合查询：一次性取出某个产品的全部关联结构
func GetGoodsProductDetailByID(productId int64) (*GoodsProductDetail, error) {
	product, err := GetGoodsProductByID(productId)
	if err != nil || product == nil {
		return nil, err
	}

	detail := &GoodsProductDetail{
		Product: product,
	}

	// 1. 获取 SKU 列表
	detail.Skus, _ = GetGoodsSkuListByProductID(productId)

	// 2. 获取套餐/包含明细
	detail.Items, _ = GetGoodsItemListByProductID(productId)

	// 3. 获取规则表
	detail.Rule, _ = GetGoodsRuleByProductID(productId)

	// 4. 获取票根专属优惠配置
	detail.TicketDiscount, _ = GetGoodsTicketDiscountByProductID(productId)

	return detail, nil
}
