package shop

import (
	"errors"
	"time"
	"tiyu/global"
)

var (
	ErrProductNotFound = errors.New("团购产品不存在")
	ErrProductStock    = errors.New("团购产品库存不足")
)

// GoodsProduct 团购产品主表（支持 吃、住、行、游、购、娱 六大业态）
type GoodsProduct struct {
	Id            int64     `json:"id" xorm:"pk autoincr comment('主键ID')"`
	ShopId        int64     `json:"shop_id" xorm:"index notnull comment('所属店铺/商家ID')"`
	UserId        int64     `json:"user_id" xorm:"index notnull comment('发布该产品的商户/店长用户ID')"`
	Title         string    `json:"title" xorm:"varchar(128) notnull comment('产品标题，如：觉海听涛双人餐/双人观景房晚')"`
	SubTitle      string    `json:"sub_title" xorm:"varchar(255) default '' comment('副标题/卖点宣导')"`
	CoverImage    string    `json:"cover_image" xorm:"varchar(512) notnull comment('产品主图URL')"`
	Images        string    `json:"images" xorm:"json comment('轮播图URL数组 JSON，例如: [\"url1\",\"url2\"]')"`
	BizType       string    `json:"biz_type" xorm:"varchar(32) notnull comment('业务类型: EAT-吃, HOTEL-住, TRAVEL-行, TOUR-游, SHOP-购, FUN-娱')"`
	ProductType   string    `json:"product_type" xorm:"varchar(32) notnull comment('产品形态: SET_MEAL-套餐, VOUCHER-代金券, ROOM_NIGHT-房晚, TICKET-门票, TRANSFER-接送机, RENTAL-租车')"`
	OriginalPrice float64   `json:"original_price" xorm:"decimal(10,2) notnull comment('划线原价')"`
	SellingPrice  float64   `json:"selling_price" xorm:"decimal(10,2) notnull comment('售卖起始价/团购价')"`
	Status        int       `json:"status" xorm:"notnull default 0 comment('状态: 0-草稿, 1-待审核, 2-已上架, 3-已下架')"`
	StockType     int       `json:"stock_type" xorm:"notnull default 1 comment('库存机制: 1-总库存机制, 2-每日/场次动态库存')"`
	TotalStock    int       `json:"total_stock" xorm:"notnull default 0 comment('总库存量')"`
	SalesCount    int       `json:"sales_count" xorm:"notnull default 0 comment('已售数量')"`
	ValidType     int       `json:"valid_type" xorm:"notnull default 1 comment('有效期类型: 1-指定时间段, 2-购买后X天内有效')"`
	ValidStart    time.Time `json:"valid_start" xorm:"comment('有效期开始时间')"`
	ValidEnd      time.Time `json:"valid_end" xorm:"comment('有效期结束时间')"`
	ValidDays     int       `json:"valid_days" xorm:"default 0 comment('购买后有效天数(ValidType=2时生效)')"`
	Created       time.Time `json:"createtime" xorm:"created int comment('创建时间')"`
	Updated       time.Time `json:"updatetime" xorm:"updated int comment('更新时间')"`
}

func (g *GoodsProduct) TableName() string {
	return "goods_product"
}

// GetGoodsProductByID 获取产品基础信息
func GetGoodsProductByID(id int64) (*GoodsProduct, error) {
	product := new(GoodsProduct)
	has, err := global.Dorm.ID(id).Get(product)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return product, nil
}

// CheckGoodsProductOwner 校验商品归属：商品ID+店铺ID+商户用户ID 三者匹配才返回 true。
// 用于编辑/删除接口防止越权操作他人商品。
func CheckGoodsProductOwner(productID, shopID, userID int64) (bool, error) {
	if productID <= 0 || shopID <= 0 || userID <= 0 {
		return false, nil
	}
	return global.Dorm.Where("id = ? AND shop_id = ? AND user_id = ?", productID, shopID, userID).Exist(new(GoodsProduct))
}

// GetGoodsProductList 分页查询团购产品列表
func GetGoodsProductList(limit, page int, search *GoodsProduct, order string) []*GoodsProduct {
	offset := page - 1
	if offset < 0 {
		offset = 0
	}
	byorder := "id DESC"
	if order != "" {
		byorder = order
	}

	listdata := []*GoodsProduct{}
	query := global.Dorm.Table("goods_product")

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
		if search.BizType != "" {
			query = query.And("biz_type = ?", search.BizType)
		}
		if search.ProductType != "" {
			query = query.And("product_type = ?", search.ProductType)
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

// GetGoodsProductTotal 获取符合条件的团购产品总条数（筛选条件与 GetGoodsProductList 保持一致）
func GetGoodsProductTotal(search *GoodsProduct) int64 {
	if search == nil {
		return 0
	}
	session := global.Dorm.Table("goods_product")
	if search.Title != "" {
		session = session.And("title like ?", "%"+search.Title+"%")
	}
	if search.ShopId > 0 {
		session = session.And("shop_id = ?", search.ShopId)
	}
	if search.UserId > 0 {
		session = session.And("user_id = ?", search.UserId)
	}
	if search.BizType != "" {
		session = session.And("biz_type = ?", search.BizType)
	}
	if search.ProductType != "" {
		session = session.And("product_type = ?", search.ProductType)
	}
	if search.Status > 0 {
		session = session.And("status = ?", search.Status)
	}
	total, err := session.Count(new(GoodsProduct))
	if err != nil {
		return 0
	}
	return total
}

// AddGoodsProduct 新增团购产品
func AddGoodsProduct(a *GoodsProduct) error {
	_, err := global.Dorm.Insert(a)
	return err
}

// EditGoodsProductByShop 修改团购产品（带店铺归属校验，防止越权）
func EditGoodsProductByShop(a *GoodsProduct) int64 {
	affected, err := global.Dorm.ID(a.Id).Where("shop_id = ?", a.ShopId).Update(a)
	if err != nil {
		return 0
	}
	return affected
}

// EditGoodsProductWithTx 事务化编辑团购商品：更新主表 + 全量重建 SKU/明细/规则/票根优惠/日历价库存。
// 调用方必须保证 product.ShopId 与当前登录商户 user_id 已校验一致，防止越权。
// calendar 传 nil 表示不动日历；传 [] 清空；传值则全量重建。
func EditGoodsProductWithTx(product *GoodsProduct, skus []*GoodsSku, items []*GoodsItem, rule *GoodsRule, ticketDiscount *GoodsTicketDiscount, calendar []*GoodsDailyCalendar) (int64, error) {
	session := global.Dorm.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return 0, err
	}

	// 0. 归属校验：该商品必须属于 product.ShopId，否则越权
	exist, err := session.Where("id = ? AND shop_id = ?", product.Id, product.ShopId).Exist(new(GoodsProduct))
	if err != nil {
		session.Rollback()
		return 0, err
	}
	if !exist {
		session.Rollback()
		return 0, ErrProductNotFound
	}

	// 1. 更新商品主表（AllCols 全字段更新，更新时间自动刷新）
	product.Updated = time.Now()
	affected, err := session.ID(product.Id).Where("shop_id = ?", product.ShopId).AllCols().Update(product)
	if err != nil {
		session.Rollback()
		return 0, err
	}
	if affected == 0 {
		session.Rollback()
		return 0, ErrProductNotFound
	}

	// 2. 全量重建 SKU（先删后插）
	if skus != nil {
		if _, err := session.Where("product_id = ?", product.Id).Delete(new(GoodsSku)); err != nil {
			session.Rollback()
			return 0, err
		}
		if len(skus) > 0 {
			for _, sku := range skus {
				sku.Id = 0 // 强制自增，避免旧 ID 冲突
				sku.ProductId = product.Id
				sku.Status = 1
				sku.Created = time.Now()
			}
			if _, err := session.Insert(skus); err != nil {
				session.Rollback()
				return 0, err
			}
		}
	}

	// 3. 全量重建明细项
	if items != nil {
		if _, err := session.Where("product_id = ?", product.Id).Delete(new(GoodsItem)); err != nil {
			session.Rollback()
			return 0, err
		}
		if len(items) > 0 {
			for _, item := range items {
				item.Id = 0
				item.ProductId = product.Id
			}
			if _, err := session.Insert(items); err != nil {
				session.Rollback()
				return 0, err
			}
		}
	}

	// 4. 全量重建核销规则（单条，先删后插）
	if rule != nil {
		if _, err := session.Where("product_id = ?", product.Id).Delete(new(GoodsRule)); err != nil {
			session.Rollback()
			return 0, err
		}
		rule.Id = 0
		rule.ProductId = product.Id
		if _, err := session.Insert(rule); err != nil {
			session.Rollback()
			return 0, err
		}
	}

	// 5. 全量重建票根优惠配置
	if ticketDiscount != nil {
		if _, err := session.Where("product_id = ?", product.Id).Delete(new(GoodsTicketDiscount)); err != nil {
			session.Rollback()
			return 0, err
		}
		if ticketDiscount.IsEnabled == 1 {
			ticketDiscount.Id = 0
			ticketDiscount.ProductId = product.Id
			if _, err := session.Insert(ticketDiscount); err != nil {
				session.Rollback()
				return 0, err
			}
		}
	}

	// 6. 全量重建日历价库存（酒店/景区等 StockType=2 业态使用）
	if calendar != nil {
		if _, err := session.Where("product_id = ?", product.Id).Delete(new(GoodsDailyCalendar)); err != nil {
			session.Rollback()
			return 0, err
		}
		if len(calendar) > 0 {
			for _, cal := range calendar {
				cal.Id = 0
				cal.ProductId = product.Id
			}
			if _, err := session.Insert(calendar); err != nil {
				session.Rollback()
				return 0, err
			}
		}
	}

	if err := session.Commit(); err != nil {
		return 0, err
	}
	return affected, nil
}
