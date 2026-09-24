package controllers

import (
	"net/http"

	"tiyu/models/shop"

	"github.com/gin-gonic/gin"
)

// GoodsProductQueryRequest 后台通用团购商品列表查询请求参数（controllers 层）。
// 与 models/shop.GoodsProduct 字段一一对应，便于接参后直接传入 models 查询函数。
type GoodsProductQueryRequest struct {
	Limit       int    `json:"limit" form:"limit"`
	Page        int    `json:"page" form:"page"`
	Order       string `json:"order" form:"order"`
	ID          int64  `json:"id" form:"id"`                         // 商品主键ID
	ShopID      int64  `json:"shop_id" form:"shop_id"`               // 所属店铺ID
	UserID      int64  `json:"user_id" form:"user_id"`               // 发布人（商户/店长用户ID）
	Title       string `json:"title" form:"title"`                   // 商品标题（模糊匹配）
	BizType     string `json:"biz_type" form:"biz_type"`             // 业态：EAT/HOTEL/TRAVEL/TOUR/SHOP/FUN
	ProductType string `json:"product_type" form:"product_type"`     // 产品形态：SET_MEAL/VOUCHER/ROOM_NIGHT/TICKET/TRANSFER/RENTAL
	Status      int    `json:"status" form:"status"`                 // 状态：0-草稿 1-待审核 2-已上架 3-已下架（0 视作"全部"）
	StockType   int    `json:"stock_type" form:"stock_type"`         // 库存机制：1-总库存 2-每日/场次动态库存
}

// GoodsProductResponse 列表返回的响应结构体（继承 models/shop.GoodsProduct，便于后续扩展展示字段）。
type GoodsProductResponse struct {
	shop.GoodsProduct
}

// toModelSearch 将 controllers 层请求结构体转换为 models/shop.GoodsProduct，作为查询 search 参数。
func (r *GoodsProductQueryRequest) toModelSearch() *shop.GoodsProduct {
	return &shop.GoodsProduct{
		Id:          r.ID,
		ShopId:      r.ShopID,
		UserId:      r.UserID,
		Title:       r.Title,
		BizType:     r.BizType,
		ProductType: r.ProductType,
		Status:      r.Status,
		StockType:   r.StockType,
	}
}

// GetAdminGoodsProductListHandler 后台通用团购商品分页列表查询。
// 支持按：商品ID、店铺ID、发布人、标题（模糊）、业态、产品形态、状态、库存机制 过滤，
// 同时返回分页信息（page/limit）与总数（total）。
// 该接口不做用户归属过滤，适用于后台管理员查看所有商品。
func GetAdminGoodsProductListHandler(c *gin.Context) {
	var req GoodsProductQueryRequest
	// 同时支持 JSON Body 与 QueryString 绑定
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "参数错误",
			"data": err.Error(),
		})
		return
	}

	// 分页默认值兜底
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	// 转换为 models 层查询条件
	search := req.toModelSearch()

	// 调用 models 层封装的列表/总数函数
	list := shop.GetGoodsProductList(req.Limit, req.Page, search, req.Order)
	total := shop.GetGoodsProductTotal(search)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": gin.H{
			"list":  list,
			"total": total,
			"page":  req.Page,
			"limit": req.Limit,
		},
	})
}