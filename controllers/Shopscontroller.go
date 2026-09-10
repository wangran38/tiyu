package controllers

import (
	"net/http"
	"strconv"

	"tiyu/models"

	"github.com/gin-gonic/gin"
)

// GetShopList 获取商家列表
type ShopQueryRequest struct {
	Limit      int     `json:"limit" form:"limit"`
	Page       int     `json:"page" form:"page"`
	Order      string  `json:"order" form:"order"`
	Status     int8    `json:"status" form:"status"`
	CategoryID int64   `json:"category_id" form:"category_id"`
	CityID     int64   `json:"city_id" form:"city_id"`
	UserID     int64   `json:"user_id" form:"user_id"`
	MerchantNo string  `json:"merchant_no" form:"merchant_no"`
	CityCode   string  `json:"city_code" form:"city_code"`
	Name       string  `json:"name" form:"name"`
	UserLng    float64 `json:"user_lng" form:"user_lng"`
	UserLat    float64 `json:"user_lat" form:"user_lat"`
	RadiusKm   float64 `json:"radius_km" form:"radius_km"`
	StartTime  string  `json:"start_time" form:"start_time"` // 新增：开始时间
	EndTime    string  `json:"end_time" form:"end_time"`     // 新增：结束时间
}

func GetShopList(c *gin.Context) {
	var req ShopQueryRequest
	_ = c.ShouldBind(&req)

	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if c.GetBool("public_shop_list") {
		req.Status = 1
	}

	search := &models.Shop{
		Status:     req.Status,
		CategoryID: req.CategoryID,
		CityID:     req.CityID,
		UserID:     req.UserID,
		MerchantNo: req.MerchantNo,
		CityCode:   req.CityCode,
		Name:       req.Name,
		UserLng:    req.UserLng,
		UserLat:    req.UserLat,
		RadiusKm:   req.RadiusKm,
	}

	// 注意：如果在 models.GetShopList 和 GetShopTotal 中需要处理时间段过滤（例如 created_at BETWEEN ? AND ?）
	// 你可以在调用前把 StartTime 和 EndTime 传进去，或者在底层 DAO 里面通过时间范围拼接 SQL。

	listdata, err := models.GetShopList(req.Limit, req.Page, search, req.StartTime, req.EndTime, req.Order)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "获取失败：" + err.Error(),
		})
		return
	}

	total := models.GetShopTotal(search, req.StartTime, req.EndTime)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": gin.H{
			"list":  listdata,
			"total": total,
		},
	})
}

// GetPublicShopList 获取前台商家列表，支持分类、城市、位置距离和排序筛选。
func GetPublicShopList(c *gin.Context) {
	c.Set("public_shop_list", true)
	GetShopList(c)
}

// GetShopDetail 获取商家详情
func GetShopDetail(c *gin.Context) {
	id, _ := strconv.ParseInt(c.PostForm("id"), 10, 64)
	if id <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "参数错误: id不能为空",
		})
		return
	}

	data, err := models.GetShopByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "获取失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": data,
	})
}

// AddShop 添加商家
func AddShop(c *gin.Context) {
	// 定义接收 JSON 数据的结构体
	var req struct {
		MerchantNo   string  `json:"merchant_no"`
		Name         string  `json:"name"`
		CategoryID   int64   `json:"category_id"`
		CityID       int64   `json:"city_id"`
		UserID       int64   `json:"user_id"`
		Logo         string  `json:"logo"`
		CoverImages  string  `json:"cover_images"`
		ContactName  string  `json:"contact_name"`
		ContactPhone string  `json:"contact_phone"`
		ServicePhone string  `json:"service_phone"`
		Description  string  `json:"description"`
		Address      string  `json:"address"`
		Longitude    float64 `json:"longitude"`
		Latitude     float64 `json:"latitude"`
		Status       int8    `json:"status"`
	}

	// 绑定前端发来的 JSON 数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "参数解析失败：" + err.Error(),
		})
		return
	}

	shop := &models.Shop{
		MerchantNo:   req.MerchantNo,
		Name:         req.Name,
		CategoryID:   req.CategoryID,
		CityID:       req.CityID,
		UserID:       req.UserID,
		Logo:         req.Logo,
		CoverImages:  req.CoverImages,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		ServicePhone: req.ServicePhone,
		Description:  req.Description,
		Address:      req.Address,
		Longitude:    req.Longitude,
		Latitude:     req.Latitude,
		Status:       req.Status,
	}

	if err := models.AddShop(shop); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "添加失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "添加成功",
	})
}

// EditShop 编辑商家
func EditShop(c *gin.Context) {
	var req struct {
		ID           int64   `json:"id"`
		MerchantNo   string  `json:"merchant_no"`
		Name         string  `json:"name"`
		CategoryID   int64   `json:"category_id"`
		CityID       int64   `json:"city_id"`
		UserID       int64   `json:"user_id"`
		Logo         string  `json:"logo"`
		CoverImages  string  `json:"cover_images"`
		ContactName  string  `json:"contact_name"`
		ContactPhone string  `json:"contact_phone"`
		ServicePhone string  `json:"service_phone"`
		Description  string  `json:"description"`
		Address      string  `json:"address"`
		Longitude    float64 `json:"longitude"`
		Latitude     float64 `json:"latitude"`
		Status       int8    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.ID <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "参数解析失败或ID不能为空",
		})
		return
	}

	shop := &models.Shop{
		ID:           req.ID,
		MerchantNo:   req.MerchantNo,
		Name:         req.Name,
		CategoryID:   req.CategoryID,
		CityID:       req.CityID,
		UserID:       req.UserID,
		Logo:         req.Logo,
		CoverImages:  req.CoverImages,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		ServicePhone: req.ServicePhone,
		Description:  req.Description,
		Address:      req.Address,
		Longitude:    req.Longitude,
		Latitude:     req.Latitude,
		Status:       req.Status,
	}

	if err := models.UpdateShop(req.ID, shop); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "更新失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "更新成功",
	})
}

// DelShop 删除商家
// DelShop 删除商家
func DelShop(c *gin.Context) {
	var req struct {
		ID int64 `json:"id"`
	}

	// 绑定前端发来的 JSON 数据
	if err := c.ShouldBindJSON(&req); err != nil || req.ID <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "参数错误: id不能为空",
		})
		return
	}

	if err := models.DeleteShop(req.ID); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "删除失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "删除成功",
	})
}
