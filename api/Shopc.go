package api

import (
	"math"
	"net/http"
	"tiyu/models"
	"tiyu/models/shop"

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
	Status         int8    `json:"status"`
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
		Status:         shop.Status,
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
	// 1. 定义一个接收前端 JSON 参数的 Map
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数解析失败"})
		return
	}

	// 2. 校验必须包含 ID
	idVal, exists := req["id"]
	if !exists {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "ID不能为空"})
		return
	}

	// 处理 ID 类型转换
	var shopID int64
	switch v := idVal.(type) {
	case float64:
		shopID = int64(v)
	case int64:
		shopID = v
	default:
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "ID格式不正确"})
		return
	}

	if shopID <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "ID不能小于等于0"})
		return
	}

	// 3. 从更新 map 中移除 id 字段，避免把主键当做更新列
	delete(req, "id")
	// 移除无需或禁止更新的系统字段
	// delete(req, "id")
	delete(req, "merchant_no") // 👈 增加这行：剔除自动生成的商户编号，防止误更新引发 1062 冲突
	delete(req, "created_at")  // 建议顺便剔除创建时间
	delete(req, "updated_at")  // 建议顺便剔除更新时间（交由数据库或后端处理）
	delete(req, "deleted_at")  // 建议顺便剔除更新时间（交由数据库或后端处理）
	// 4. 调用更新方法
	if err := models.UpdateShopMap(shopID, req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "更新失败：" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功"})
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

// PublicShopDetailRequest 公开商家详情请求体
type PublicShopDetailRequest struct {
	ShopID  int64   `json:"shop_id" form:"shop_id" binding:"required"` // 商家ID，必填
	UserLng float64 `json:"user_lng" form:"user_lng"`                  // 可选，用户经度（用于计算距离）
	UserLat float64 `json:"user_lat" form:"user_lat"`                  // 可选，用户纬度
}

// PublicShopDetail C 端商家主页公开详情（已过滤敏感经营数据）
type PublicShopDetail struct {
	ID             int64   `json:"id"`
	MerchantNo     string  `json:"merchant_no"`
	CategoryID     int64   `json:"category_id"`
	CityID         int64   `json:"city_id"`
	CityCode       string  `json:"city_code"`
	Status         int8    `json:"status"`
	Name           string  `json:"name"`
	Logo           string  `json:"logo"`
	CoverImages    string  `json:"cover_images"`
	ServicePhone   string  `json:"service_phone"`
	Description    string  `json:"description"`
	Discounts      string  `json:"discounts"`
	ProvinceCode   string  `json:"province_code"`
	DistrictCode   string  `json:"district_code"`
	Address        string  `json:"address"`
	Longitude      float64 `json:"longitude"`
	Latitude       float64 `json:"latitude"`
	OpeningHours   string  `json:"opening_hours"`
	DeliveryType   int8    `json:"delivery_type"`
	MinOrderAmount float64 `json:"min_order_amount"`
	AvgCost        float64 `json:"avg_cost"`
	DistanceKm     float64 `json:"distance_km"`
	GoodsCount     int64   `json:"goods_count"` // 在售团购商品数量
}

// haversineKm 球面距离计算（公里），与 models 包同公式
func haversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const r = 6371.0
	rad := math.Pi / 180
	phi1, phi2 := lat1*rad, lat2*rad
	dPhi := (lat2 - lat1) * rad
	dLambda := (lng2 - lng1) * rad
	a := math.Sin(dPhi/2)*math.Sin(dPhi/2) + math.Cos(phi1)*math.Cos(phi2)*math.Sin(dLambda/2)*math.Sin(dLambda/2)
	return r * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// GetPublicShopDetailHandler 公开的商家详情接口：C 端用户查看商家主页。
// 无需登录；不返回佣金/手续费/负责人联系方式等敏感经营数据；
// 可选传用户经纬度计算距离，附带在售团购商品数量。
// 请求路由: GET/POST /api/shop/detail
func GetPublicShopDetailHandler(c *gin.Context) {
	var req PublicShopDetailRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误：shop_id 必填", "data": err.Error()})
		return
	}

	shopModel, err := models.GetShopByID(req.ShopID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询商家失败", "data": err.Error()})
		return
	}
	if shopModel == nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "商家不存在"})
		return
	}
	// 待审核/已冻结/审核驳回的商家对 C 端不可见（营业中/休息中可见）
	if shopModel.Status == 0 || shopModel.Status == 3 || shopModel.Status == 4 {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "商家不存在或暂未开放"})
		return
	}

	distance := 0.0
	if req.UserLng != 0 || req.UserLat != 0 {
		distance = haversineKm(req.UserLat, req.UserLng, shopModel.Latitude, shopModel.Longitude)
	}

	// 在售团购商品数量（已上架 status=2）
	goodsCount := shop.GetGoodsProductTotal(&shop.GoodsProduct{
		ShopId: req.ShopID,
		Status: 2,
	})

	detail := PublicShopDetail{
		ID:             shopModel.ID,
		MerchantNo:     shopModel.MerchantNo,
		CategoryID:     shopModel.CategoryID,
		CityID:         shopModel.CityID,
		CityCode:       shopModel.CityCode,
		Status:         shopModel.Status,
		Name:           shopModel.Name,
		Logo:           shopModel.Logo,
		CoverImages:    shopModel.CoverImages,
		ServicePhone:   shopModel.ServicePhone,
		Description:    shopModel.Description,
		Discounts:      shopModel.Discounts,
		ProvinceCode:   shopModel.ProvinceCode,
		DistrictCode:   shopModel.DistrictCode,
		Address:        shopModel.Address,
		Longitude:      shopModel.Longitude,
		Latitude:       shopModel.Latitude,
		OpeningHours:   shopModel.OpeningHours,
		DeliveryType:   shopModel.DeliveryType,
		MinOrderAmount: shopModel.MinOrderAmount,
		AvgCost:        shopModel.AvgCost,
		DistanceKm:     math.Round(distance*100) / 100, // 保留两位小数
		GoodsCount:     goodsCount,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    detail,
	})
}
