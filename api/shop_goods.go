package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"tiyu/global"
	"tiyu/models"
	"tiyu/models/shop"
)

// AddGoodsProductRequest 商家添加团购商品请求体
type AddGoodsProductRequest struct {
	ShopID         int64                     `json:"shop_id" binding:"required"`
	Title          string                    `json:"title" binding:"required"`
	SubTitle       string                    `json:"sub_title"`
	CoverImage     string                    `json:"cover_image" binding:"required"`
	Images         string                    `json:"images"`
	BizType        string                    `json:"biz_type" binding:"required"`
	ProductType    string                    `json:"product_type" binding:"required"`
	OriginalPrice  float64                   `json:"original_price" binding:"required"`
	SellingPrice   float64                   `json:"selling_price" binding:"required"`
	StockType      int                       `json:"stock_type"`
	TotalStock     int                       `json:"total_stock" binding:"required"`
	ValidType      int                       `json:"valid_type"`
	ValidDays      int                       `json:"valid_days"`
	ValidStart     string                    `json:"valid_start"`
	ValidEnd       string                    `json:"valid_end"`
	Skus           []*shop.GoodsSku          `json:"skus"`
	Items          []*shop.GoodsItem         `json:"items"`
	Rule           *shop.GoodsRule           `json:"rule"`
	TicketDiscount *shop.GoodsTicketDiscount `json:"ticket_discount"`
}

// CreateGoodsProductWithTx 开启事务添加团购商品及附属明细数据。
// 日历价库存(calendar)不在添加时处理，商家后续通过编辑接口按需配置。
func CreateGoodsProductWithTx(product *shop.GoodsProduct, skus []*shop.GoodsSku, items []*shop.GoodsItem, rule *shop.GoodsRule, ticketDiscount *shop.GoodsTicketDiscount) (int64, error) {
	session := global.Dorm.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return 0, err
	}

	// 1. 保存商品主表
	product.Status = 1 // 1-待审核 或 2-已上架，可根据业务规则调整
	product.Created = time.Now()
	product.Updated = time.Now()
	if _, err := session.Insert(product); err != nil {
		session.Rollback()
		return 0, err
	}

	productID := product.Id

	// 2. 保存 SKU 规格
	if len(skus) > 0 {
		for _, sku := range skus {
			sku.ProductId = productID
			sku.Status = 1
			sku.Created = time.Now()
		}
		if _, err := session.Insert(skus); err != nil {
			session.Rollback()
			return 0, err
		}
	}

	// 3. 保存明细清单
	if len(items) > 0 {
		for _, item := range items {
			item.ProductId = productID
		}
		if _, err := session.Insert(items); err != nil {
			session.Rollback()
			return 0, err
		}
	}

	// 4. 保存核销规则
	if rule != nil {
		rule.ProductId = productID
		if _, err := session.Insert(rule); err != nil {
			session.Rollback()
			return 0, err
		}
	}

	// 5. 保存票根优惠设置
	if ticketDiscount != nil && ticketDiscount.IsEnabled == 1 {
		ticketDiscount.ProductId = productID
		if _, err := session.Insert(ticketDiscount); err != nil {
			session.Rollback()
			return 0, err
		}
	}

	if err := session.Commit(); err != nil {
		return 0, err
	}

	return productID, nil
}

// AddGoodsProductHandler 商家添加团购商品 API。
func AddGoodsProductHandler(c *gin.Context) {
	userIDValue, exists := c.Get("userId")
	userID, ok := userIDValue.(int64)
	if !exists || !ok || userID <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 401, "message": "请先登录"})
		return
	}

	var req AddGoodsProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误", "data": err.Error()})
		return
	}

	var startTime, endTime time.Time
	if req.ValidStart != "" {
		startTime, _ = time.ParseInLocation("2006-01-02 15:04:05", req.ValidStart, time.Local)
	}
	if req.ValidEnd != "" {
		endTime, _ = time.ParseInLocation("2006-01-02 15:04:05", req.ValidEnd, time.Local)
	}

	stockType := req.StockType
	if stockType <= 0 {
		stockType = 1
	}

	validType := req.ValidType
	if validType <= 0 {
		validType = 1
	}

	product := &shop.GoodsProduct{
		ShopId:        req.ShopID,
		UserId:        userID,
		Title:         req.Title,
		SubTitle:      req.SubTitle,
		CoverImage:    req.CoverImage,
		Images:        req.Images,
		BizType:       req.BizType,
		ProductType:   req.ProductType,
		OriginalPrice: req.OriginalPrice,
		SellingPrice:  req.SellingPrice,
		StockType:     stockType,
		TotalStock:    req.TotalStock,
		SalesCount:    0,
		ValidType:     validType,
		ValidStart:    startTime,
		ValidEnd:      endTime,
		ValidDays:     req.ValidDays,
	}

	productID, err := CreateGoodsProductWithTx(product, req.Skus, req.Items, req.Rule, req.TicketDiscount)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "添加团购商品失败", "data": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "添加团购商品成功",
		"data": gin.H{
			"product_id": productID,
		},
	})
}

// GoodsProductListRequest 商家查询自己的团购商品列表请求体
type GoodsProductListRequest struct {
	ShopID int64  `json:"shop_id" form:"shop_id"` // 可选，指定店铺；不传则查该商户名下全部商品
	Status int    `json:"status" form:"status"`   // 可选，0-全部 1-待审核 2-已上架 3-已下架
	Title  string `json:"title" form:"title"`     // 可选，标题关键词
	Limit  int    `json:"limit" form:"limit"`
	Page   int    `json:"page" form:"page"`
	Order  string `json:"order" form:"order"`
}

// GetGoodsProductListHandler 商家查看自己发布的全部团购商品列表（含草稿、待审核、已下架）。
// 请求路由: POST /api/user/shop/goods/list
func GetGoodsProductListHandler(c *gin.Context) {
	userIDValue, exists := c.Get("userId")
	userID, ok := userIDValue.(int64)
	if !exists || !ok || userID <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 401, "message": "请先登录"})
		return
	}

	var req GoodsProductListRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误", "data": err.Error()})
		return
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	// 强制按当前登录商户过滤，防止越权查看他人商品
	search := &shop.GoodsProduct{
		UserId: userID,
		ShopId: req.ShopID,
		Title:  req.Title,
		Status: req.Status,
	}

	list := shop.GetGoodsProductList(req.Limit, req.Page, search, req.Order)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":  list,
			"total": shop.GetGoodsProductTotal(search),
		},
	})
}

// PublicShopGoodsListRequest 公开查询指定商家的团购商品列表请求体
type PublicShopGoodsListRequest struct {
	ShopID      int64  `json:"shop_id" form:"shop_id" binding:"required"` // 商家ID，必填
	BizType     string `json:"biz_type" form:"biz_type"`                  // 可选，业态: EAT/HOTEL/TRAVEL/TOUR/SHOP/FUN
	ProductType string `json:"product_type" form:"product_type"`          // 可选，产品形态: SET_MEAL/VOUCHER/ROOM_NIGHT/TICKET/TRANSFER/RENTAL
	Keyword     string `json:"keyword" form:"keyword"`                    // 可选，标题关键词
	Limit       int    `json:"limit" form:"limit"`
	Page        int    `json:"page" form:"page"`
	Order       string `json:"order" form:"order"`
}

// GetPublicShopGoodsListHandler 公开接口：根据商家ID查看该商家已上架的团购商品列表。
// 无需登录，只返回已上架(status=2)的商品。
// 请求路由: GET/POST /api/shop/goods
func GetPublicShopGoodsListHandler(c *gin.Context) {
	var req PublicShopGoodsListRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误：shop_id 必填", "data": err.Error()})
		return
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	// 公开列表只展示已上架商品
	search := &shop.GoodsProduct{
		ShopId:      req.ShopID,
		Status:      2,
		Title:       req.Keyword,
		BizType:     req.BizType,
		ProductType: req.ProductType,
	}

	list := shop.GetGoodsProductList(req.Limit, req.Page, search, req.Order)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":  list,
			"total": shop.GetGoodsProductTotal(search),
		},
	})
}

// EditGoodsProductRequest 商家编辑团购商品请求体
type EditGoodsProductRequest struct {
	ID             int64                      `json:"id" binding:"required"` // 商品ID，必填
	ShopID         int64                      `json:"shop_id" binding:"required"`
	Title          string                     `json:"title" binding:"required"`
	SubTitle       string                     `json:"sub_title"`
	CoverImage     string                     `json:"cover_image" binding:"required"`
	Images         string                     `json:"images"`
	BizType        string                     `json:"biz_type" binding:"required"`
	ProductType    string                     `json:"product_type" binding:"required"`
	OriginalPrice  float64                    `json:"original_price" binding:"required"`
	SellingPrice   float64                    `json:"selling_price" binding:"required"`
	StockType      int                        `json:"stock_type"`
	TotalStock     int                        `json:"total_stock" binding:"required"`
	Status         int                        `json:"status"` // 0-草稿 1-待审核 2-已上架 3-已下架
	ValidType      int                        `json:"valid_type"`
	ValidDays      int                        `json:"valid_days"`
	ValidStart     string                     `json:"valid_start"`
	ValidEnd       string                     `json:"valid_end"`
	Skus           []*shop.GoodsSku           `json:"skus"`            // 传 nil 表示不动 SKU；传 [] 清空；传值则全量重建
	Items          []*shop.GoodsItem          `json:"items"`           // 同上
	Rule           *shop.GoodsRule            `json:"rule"`            // 传 nil 表示不动规则
	TicketDiscount *shop.GoodsTicketDiscount  `json:"ticket_discount"` // 传 nil 表示不动票根优惠
	Calendar       []*shop.GoodsDailyCalendar `json:"calendar"`        // 日历价库存；传 nil 不动；传 [] 清空；传值全量重建（酒店/景区 StockType=2 用）
}

// EditGoodsProductHandler 商家编辑团购商品。
// 事务化更新主表 + 全量重建 SKU/明细/规则/票根优惠（先删后插）。
// 请求路由: POST /api/user/shop/goods/edit
func EditGoodsProductHandler(c *gin.Context) {
	userIDValue, exists := c.Get("userId")
	userID, ok := userIDValue.(int64)
	if !exists || !ok || userID <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 401, "message": "请先登录"})
		return
	}

	var req EditGoodsProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误", "data": err.Error()})
		return
	}

	// 归属校验：该商品必须属于当前登录商户（防越权编辑他人商品）
	belongs, err := shop.CheckGoodsProductOwner(req.ID, req.ShopID, userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "校验商品归属失败", "data": err.Error()})
		return
	}
	if !belongs {
		c.JSON(http.StatusOK, gin.H{"code": 403, "message": "无权编辑该商品：商品不存在或不属于当前商户"})
		return
	}

	var startTime, endTime time.Time
	if req.ValidStart != "" {
		startTime, _ = time.ParseInLocation("2006-01-02 15:04:05", req.ValidStart, time.Local)
	}
	if req.ValidEnd != "" {
		endTime, _ = time.ParseInLocation("2006-01-02 15:04:05", req.ValidEnd, time.Local)
	}

	stockType := req.StockType
	if stockType <= 0 {
		stockType = 1
	}
	validType := req.ValidType
	if validType <= 0 {
		validType = 1
	}

	product := &shop.GoodsProduct{
		Id:            req.ID,
		ShopId:        req.ShopID,
		UserId:        userID,
		Title:         req.Title,
		SubTitle:      req.SubTitle,
		CoverImage:    req.CoverImage,
		Images:        req.Images,
		BizType:       req.BizType,
		ProductType:   req.ProductType,
		OriginalPrice: req.OriginalPrice,
		SellingPrice:  req.SellingPrice,
		StockType:     stockType,
		TotalStock:    req.TotalStock,
		Status:        req.Status,
		ValidType:     validType,
		ValidStart:    startTime,
		ValidEnd:      endTime,
		ValidDays:     req.ValidDays,
	}

	affected, err := shop.EditGoodsProductWithTx(product, req.Skus, req.Items, req.Rule, req.TicketDiscount, req.Calendar)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "编辑团购商品失败", "data": err.Error()})
		return
	}
	if affected == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "商品不存在或未变更"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "编辑团购商品成功",
		"data": gin.H{
			"id":      req.ID,
			"updated": affected,
		},
	})
}

// GoodsProductDetailRequest 商家获取商品详情（编辑回显）请求体
type GoodsProductDetailRequest struct {
	ID     int64 `json:"id" form:"id" binding:"required"` // 商品ID，必填
	ShopID int64 `json:"shop_id" form:"shop_id"`          // 可选，传了则一并校验归属
}

// GetGoodsProductDetailHandler 商家获取自己商品的完整详情（主表 + SKU + 明细 + 规则 + 票根优惠）。
// 专为编辑页回显设计：附表数据全部原样返回（含已关闭的票根优惠配置）。
// 请求路由: POST /api/user/shop/goods/detail
func GetGoodsProductDetailHandler(c *gin.Context) {
	userIDValue, exists := c.Get("userId")
	userID, ok := userIDValue.(int64)
	if !exists || !ok || userID <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 401, "message": "请先登录"})
		return
	}

	var req GoodsProductDetailRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误：id 必填", "data": err.Error()})
		return
	}

	// 归属校验：商品必须属于当前登录商户
	product, err := shop.GetGoodsProductByID(req.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询商品失败", "data": err.Error()})
		return
	}
	if product == nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "商品不存在"})
		return
	}
	if product.UserId != userID {
		c.JSON(http.StatusOK, gin.H{"code": 403, "message": "无权查看该商品"})
		return
	}
	if req.ShopID > 0 && product.ShopId != req.ShopID {
		c.JSON(http.StatusOK, gin.H{"code": 403, "message": "无权查看该商品：店铺不匹配"})
		return
	}

	// 组装完整详情（附表全部原样返回，用于编辑回显）
	// SKU 用原始记录查询（含禁用状态），避免回显丢数据
	skus, _ := shop.GetGoodsSkuRawListByProductID(req.ID)
	items, _ := shop.GetGoodsItemListByProductID(req.ID)
	rule, _ := shop.GetGoodsRuleByProductID(req.ID)
	// 票根优惠用原始记录查询（不过滤 is_enabled），关闭状态也要回显
	ticketDiscount, _ := shop.GetGoodsTicketDiscountRawByProductID(req.ID)
	// 日历价库存原样返回（酒店/景区 StockType=2 用）
	calendar, _ := shop.GetGoodsDailyCalendarList(req.ID, 0, "", "")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"product":         product,
			"skus":            skus,
			"items":           items,
			"rule":            rule,
			"ticket_discount": ticketDiscount,
			"calendar":        calendar,
		},
	})
}

// PublicGoodsDetailRequest 公开商品详情请求体
type PublicGoodsDetailRequest struct {
	ID int64 `json:"id" form:"id" binding:"required"` // 商品ID，必填
}

// PublicGoodsDetailShopInfo 商家摘要（公开商品详情附带）
type PublicGoodsDetailShopInfo struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Logo         string  `json:"logo"`
	Status       int8    `json:"status"`
	Address      string  `json:"address"`
	Longitude    float64 `json:"longitude"`
	Latitude     float64 `json:"latitude"`
	OpeningHours string  `json:"opening_hours"`
	ServicePhone string  `json:"service_phone"`
	AvgCost      float64 `json:"avg_cost"`
}

// GetPublicGoodsDetailHandler 公开的团购商品详情接口：C 端商品详情页。
// 无需登录；只返回已上架(status=2)商品；SKU 只含启用规格；
// 票根优惠只返回已开启的；附商家摘要（不含敏感经营数据）。
// 请求路由: GET/POST /api/shop/goods/detail
func GetPublicGoodsDetailHandler(c *gin.Context) {
	var req PublicGoodsDetailRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误：id 必填", "data": err.Error()})
		return
	}

	product, err := shop.GetGoodsProductByID(req.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询商品失败", "data": err.Error()})
		return
	}
	// 未上架商品对 C 端不可见
	if product == nil || product.Status != 2 {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "商品不存在或已下架"})
		return
	}

	// 公开视角：SKU 只含启用规格；票根优惠只含已开启的
	skus, _ := shop.GetGoodsSkuListByProductID(req.ID)
	items, _ := shop.GetGoodsItemListByProductID(req.ID)
	rule, _ := shop.GetGoodsRuleByProductID(req.ID)
	ticketDiscount, _ := shop.GetGoodsTicketDiscountByProductID(req.ID)
	// 日历价库存（酒店/景区 StockType=2 时前端渲染日历选择）
	calendar, _ := shop.GetGoodsDailyCalendarList(req.ID, 0, "", "")

	// 商家摘要
	var shopInfo *PublicGoodsDetailShopInfo
	shopModel, err := models.GetShopByID(product.ShopId)
	if err == nil && shopModel != nil && shopModel.Status != 0 && shopModel.Status != 3 && shopModel.Status != 4 {
		shopInfo = &PublicGoodsDetailShopInfo{
			ID:           shopModel.ID,
			Name:         shopModel.Name,
			Logo:         shopModel.Logo,
			Status:       shopModel.Status,
			Address:      shopModel.Address,
			Longitude:    shopModel.Longitude,
			Latitude:     shopModel.Latitude,
			OpeningHours: shopModel.OpeningHours,
			ServicePhone: shopModel.ServicePhone,
			AvgCost:      shopModel.AvgCost,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"product":         product,
			"skus":            skus,
			"items":           items,
			"rule":            rule,
			"ticket_discount": ticketDiscount,
			"calendar":        calendar,
			"shop":            shopInfo,
		},
	})
}
