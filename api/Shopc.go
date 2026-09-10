package api

import (
	"net/http"
	"tiyu/models"

	"github.com/gin-gonic/gin"
)

type ShopListRequest struct {
	Limit      int     `json:"limit" form:"limit"`
	Page       int     `json:"page" form:"page"`
	Order      string  `json:"order" form:"order"`
	CategoryID int64   `json:"category_id" form:"category_id"`
	CityID     int64   `json:"city_id" form:"city_id"`
	CityCode   string  `json:"city_code" form:"city_code"`
	Name       string  `json:"name" form:"name"`
	UserLng    float64 `json:"user_lng" form:"user_lng"`
	UserLat    float64 `json:"user_lat" form:"user_lat"`
	RadiusKm   float64 `json:"radius_km" form:"radius_km"`
	StartTime  string  `json:"start_time" form:"start_time"`
	EndTime    string  `json:"end_time" form:"end_time"`
}

type ShopListItem struct {
	ID             int64   `json:"id"`
	MerchantNo     string  `json:"merchant_no"`
	CategoryID     int64   `json:"category_id"`
	CityID         int64   `json:"city_id"`
	CityCode       string  `json:"city_code"`
	Name           string  `json:"name"`
	Logo           string  `json:"logo"`
	CoverImages    string  `json:"cover_images"`
	Description    string  `json:"description"`
	Discounts      string  `json:"discounts"`
	Address        string  `json:"address"`
	Longitude      float64 `json:"longitude"`
	Latitude       float64 `json:"latitude"`
	OpeningHours   string  `json:"opening_hours"`
	DeliveryType   int8    `json:"delivery_type"`
	DeliveryRadius float64 `json:"delivery_radius"`
	MinOrderAmount float64 `json:"min_order_amount"`
	AvgCost        float64 `json:"avg_cost"`
	DistanceKm     float64 `json:"distance_km"`
}

func toShopListItem(shop *models.Shop) ShopListItem {
	return ShopListItem{
		ID:             shop.ID,
		MerchantNo:     shop.MerchantNo,
		CategoryID:     shop.CategoryID,
		CityID:         shop.CityID,
		CityCode:       shop.CityCode,
		Name:           shop.Name,
		Logo:           shop.Logo,
		CoverImages:    shop.CoverImages,
		Description:    shop.Description,
		Discounts:      shop.Discounts,
		Address:        shop.Address,
		Longitude:      shop.Longitude,
		Latitude:       shop.Latitude,
		OpeningHours:   shop.OpeningHours,
		DeliveryType:   shop.DeliveryType,
		DeliveryRadius: shop.DeliveryRadius,
		MinOrderAmount: shop.MinOrderAmount,
		AvgCost:        shop.AvgCost,
		DistanceKm:     shop.DistanceKm,
	}
}

// GetShopList 获取前台商家列表，支持分类、城市、距离和排序筛选。
func GetShopList(c *gin.Context) {
	var req ShopListRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数绑定失败", "data": err.Error()})
		return
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	search := &models.Shop{
		Status:     1,
		CategoryID: req.CategoryID,
		CityID:     req.CityID,
		CityCode:   req.CityCode,
		Name:       req.Name,
		UserLng:    req.UserLng,
		UserLat:    req.UserLat,
		RadiusKm:   req.RadiusKm,
	}
	list, err := models.GetShopList(req.Limit, req.Page, search, req.StartTime, req.EndTime, req.Order)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "获取商家列表失败", "data": err.Error()})
		return
	}
	items := make([]ShopListItem, 0, len(list))
	for _, shop := range list {
		items = append(items, toShopListItem(shop))
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":  items,
			"total": models.GetShopTotal(search, req.StartTime, req.EndTime),
		},
	})
}

// EditShop 商家通过审核后编辑/补充店铺资料接口
// 请求路由: POST /api/user/shop/edit
func EditShop(c *gin.Context) {
	// 1. 获取当前登录用户的 ID
	userIDVal, exists := c.Get("userId")
	if !exists {
		c.JSON(200, gin.H{
			"code":    401,
			"message": "请先登录",
			"data":    nil,
		})
		return
	}
	userID := userIDVal.(int64)

	// 2. 采用结构体接收 POST 提交的 JSON 数据（复用 models.Shop 结构体，或者你专门接收修改字段的 Request 结构体）
	var req models.Shop
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "参数绑定失败",
			"data":    err.Error(),
		})
		return
	}

	// 3. 必须传入店铺 ID (req.ID) 来确认修改的是哪家店铺
	if req.ID <= 0 {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "店铺ID不能为空",
			"data":    nil,
		})
		return
	}

	// 4. 查询该店铺是否存在，并校验所有权（防止越权修改别人的店铺）
	existingShop, err := models.GetShopByID(req.ID)
	if err != nil {
		c.JSON(200, gin.H{
			"code":    500,
			"message": "服务器内部错误",
			"data":    nil,
		})
		return
	}
	if existingShop == nil {
		c.JSON(200, gin.H{
			"code":    404,
			"message": "未找到对应的店铺信息",
			"data":    nil,
		})
		return
	}

	// 校验归属权
	if existingShop.UserID != userID {
		c.JSON(200, gin.H{
			"code":    403,
			"message": "无权操作该店铺资料",
			"data":    nil,
		})
		return
	}

	// 5. 执行更新操作（你可以选择使用你 model 里的 UpdateShop，或者指定修改某些特定可补充的字段）
	// 这里以通过 ID 指定更新部分或全部可编辑字段为例：
	err = models.UpdateShop(req.ID, &req)
	if err != nil {
		c.JSON(200, gin.H{
			"code":    500,
			"message": "保存修改失败",
			"data":    err.Error(),
		})
		return
	}

	// 6. 返回成功响应
	c.JSON(200, gin.H{
		"code":    200,
		"message": "资料修改成功",
		"data":    nil,
	})
}

// GetShopInfo 根据店铺 ID 和用户 ID 获取商家详细信息
// GetShopInfoRequest 定义获取商家详情的请求参数结构体
type GetShopInfoRequest struct {
	ShopID int64 `json:"shop_id" binding:"required"`
}

// GetShopInfo 根据店铺 ID 和用户 ID 获取商家详细信息 (POST 方式)
// 请求路由: POST /api/user/shop/info
func GetShopInfo(c *gin.Context) {
	// 1. 获取当前登录用户的 ID
	userIDVal, exists := c.Get("userId")
	if !exists {
		c.JSON(200, gin.H{
			"code":    401,
			"message": "请先登录",
			"data":    nil,
		})
		return
	}
	userID := userIDVal.(int64)

	// 2. 采用结构体绑定 POST 提交的 JSON 数据
	var req GetShopInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "参数绑定失败",
			"data":    err.Error(),
		})
		return
	}

	if req.ShopID <= 0 {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "参数 shop_id 错误",
			"data":    nil,
		})
		return
	}

	// 3. 根据店铺 ID 调用已有 model 查询店铺信息
	shop, err := models.GetShopByID(req.ShopID)
	if err != nil {
		c.JSON(200, gin.H{
			"code":    500,
			"message": "服务器内部错误",
			"data":    nil,
		})
		return
	}
	if shop == nil {
		c.JSON(200, gin.H{
			"code":    404,
			"message": "未找到对应的店铺信息",
			"data":    nil,
		})
		return
	}

	// 4. 严格校验归属权：确保该店铺属于当前登录的用户
	if shop.UserID != userID {
		c.JSON(200, gin.H{
			"code":    403,
			"message": "无权查看该店铺信息",
			"data":    nil,
		})
		return
	}

	// 5. 返回该商家的所有详细信息
	c.JSON(200, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    shop,
	})
}
