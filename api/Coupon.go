package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"tiyu/models"

	"github.com/gin-gonic/gin"
)

// AddCouponRequest 商家添加优惠券请求参数
type AddCouponRequest struct {
	Title          string  `json:"title" binding:"required"`           // 优惠券标题
	Categories     []int   `json:"categories" binding:"required"`      // 多业态分类ID数组，例如: [1, 3, 4] (1-景区, 2-住宿, 3-餐饮, 4-文创零售, 5-交通服务, 6-演出展览)
	CouponType     int     `json:"coupon_type" binding:"required"`     // 形式: 1-满减券, 2-折扣券, 3-首道门票折扣, 4-专属票价
	DiscountAmount float64 `json:"discount_amount" binding:"required"` // 优惠面额或折扣率
	MinPoint       float64 `json:"min_point"`                          // 使用门槛金额
	TotalCount     int     `json:"total_count" binding:"required"`     // 发放总数量（库存）
	StartTime      string  `json:"start_time" binding:"required"`      // 生效时间 (格式: "2006-01-02 15:04:05")
	EndTime        string  `json:"end_time" binding:"required"`        // 过期时间
}

// EditCouponRequest 商家编辑优惠券请求参数
type EditCouponRequest struct {
	Id             int64   `json:"id" binding:"required"`              // 优惠券 ID
	Title          string  `json:"title" binding:"required"`           // 优惠券标题
	Categories     []int   `json:"categories" binding:"required"`      // 多业态分类 ID 数组
	CouponType     int     `json:"coupon_type" binding:"required"`     // 优惠券类型
	DiscountAmount float64 `json:"discount_amount" binding:"required"` // 优惠面额或折扣率
	MinPoint       float64 `json:"min_point"`                          // 使用门槛金额
	TotalCount     int     `json:"total_count" binding:"required"`     // 发放总数量
	StartTime      string  `json:"start_time" binding:"required"`      // 生效时间
	EndTime        string  `json:"end_time" binding:"required"`        // 过期时间
}

// CouponListRequest 商家优惠券列表请求参数
type CouponListRequest struct {
	Limit      int    `json:"limit" form:"limit"`
	Page       int    `json:"page" form:"page"`
	Order      string `json:"order" form:"order"`
	Title      string `json:"title" form:"title"`
	Categories string `json:"categories" form:"categories"`
	Status     int    `json:"status" form:"status"`
}

// MemberCouponListRequest 会员我的优惠券列表请求参数
type MemberCouponListRequest struct {
	Limit int    `json:"limit" form:"limit"`
	Page  int    `json:"page" form:"page"`
	Order string `json:"order" form:"order"`
}

// GetMyCouponListHandler 会员“我的优惠券”接口：展示当前会员已领取（兑换成功）的优惠券，连表票根展示兑换记录
func GetMyCouponListHandler(c *gin.Context) {
	userIDValue, exists := c.Get("userId")
	userID, ok := userIDValue.(int64)
	if !exists || !ok || userID <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 401, "message": "请先登录"})
		return
	}

	var req MemberCouponListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误", "data": err.Error()})
		return
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	list := models.GetMemberCouponList(req.Limit, req.Page, uint64(userID))

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":  list,
			"total": models.GetMemberCouponTotal(uint64(userID)),
			"page":  req.Page,
			"limit": req.Limit,
		},
	})
}

// CouponShopListRequest 按店铺查询优惠券列表请求参数
type CouponShopListRequest struct {
	ShopID int64  `json:"shop_id" form:"shop_id" binding:"required"`
	Limit  int    `json:"limit" form:"limit"`
	Page   int    `json:"page" form:"page"`
	Order  string `json:"order" form:"order"`
	Title  string `json:"title" form:"title"`
}

// GetCouponListByShopIDHandler 获取指定店铺的有效优惠券分页列表
func GetCouponListByShopIDHandler(c *gin.Context) {
	var req CouponShopListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数绑定失败", "data": err.Error()})
		return
	}
	if req.ShopID <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "shop_id 参数错误"})
		return
	}
	shop, err := models.GetShopByID(req.ShopID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询商户信息失败", "data": err.Error()})
		return
	}
	if shop == nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "商户不存在"})
		return
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	search := &models.Coupon{
		ShopId: req.ShopID,
		Title:  req.Title,
		Status: 1,
	}
	list := models.GetCouponList(req.Limit, req.Page, search, req.Order)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"shop": gin.H{
				"id":            shop.ID,
				"name":          shop.Name,
				"logo":          shop.Logo,
				"cover_images":  shop.CoverImages,
				"contact_name":  shop.ContactName,
				"contact_phone": shop.ContactPhone,
				"service_phone": shop.ServicePhone,
				"description":   shop.Description,
				"discounts":     shop.Discounts,
				"address":       shop.Address,
				"longitude":     shop.Longitude,
				"latitude":      shop.Latitude,
				"opening_hours": shop.OpeningHours,
			},
			"list":  list,
			"total": models.GetCouponTotal(search),
			"page":  req.Page,
			"limit": req.Limit,
		},
	})
}

// GetCouponListHandler 获取当前商户的优惠券分页列表
func GetCouponListHandler(c *gin.Context) {
	userIdVal, userExists := c.Get("userId")
	userId, ok := userIdVal.(int64)
	if !userExists || !ok || userId <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 401, "message": "未登录或用户信息无效"})
		return
	}

	shop, err := models.GetShopByUserID(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询商户信息失败", "data": err.Error()})
		return
	}
	if shop == nil {
		c.JSON(http.StatusOK, gin.H{"code": 403, "message": "当前用户未绑定店铺"})
		return
	}

	var req CouponListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数绑定失败", "data": err.Error()})
		return
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	search := &models.Coupon{
		ShopId:     shop.ID,
		Title:      req.Title,
		Categories: req.Categories,
		Status:     req.Status,
	}
	list := models.GetCouponList(req.Limit, req.Page, search, req.Order)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":  list,
			"total": models.GetCouponTotal(search),
			"page":  req.Page,
			"limit": req.Limit,
		},
	})
}

// AddCouponHandler 商家添加优惠券接口 (POST /api/shop/coupon/add)
func AddCouponHandler(c *gin.Context) {
	// 1. 从上下文中获取当前登录用户，并查询其所属店铺
	userIdVal, userExists := c.Get("userId")
	if !userExists {
		c.JSON(http.StatusOK, gin.H{"code": 401, "msg": "未登录或用户信息不存在"})
		return
	}
	userId, ok := userIdVal.(int64)
	if !ok || userId <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 401, "msg": "用户信息无效"})
		return
	}

	shop, err := models.GetShopByUserID(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "查询商户信息失败: " + err.Error()})
		return
	}
	if shop == nil {
		c.JSON(http.StatusOK, gin.H{"code": 403, "msg": "当前用户未绑定店铺"})
		return
	}
	shopId := shop.ID

	// 2. 绑定并校验前端传参
	var req AddCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	// 3. 处理多分类数组转字符串（例如 [1, 3] 转为 "1,3"）
	if len(req.Categories) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "请至少选择一个业态分类"})
		return
	}
	var catStrs []string
	for _, id := range req.Categories {
		catStrs = append(catStrs, fmt.Sprintf("%d", id))
	}
	categoriesStr := strings.Join(catStrs, ",")

	// 4. 解析时间字符串
	loc, _ := time.LoadLocation("Local")
	startTime, err1 := time.ParseInLocation("2006-01-02 15:04:05", req.StartTime, loc)
	endTime, err2 := time.ParseInLocation("2006-01-02 15:04:05", req.EndTime, loc)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "时间格式错误，请使用 YYYY-MM-DD HH:mm:ss"})
		return
	}

	// 5. 构造 Coupon 模型对象
	coupon := &models.Coupon{
		ShopId:         shopId,
		UserId:         userId,
		Title:          req.Title,
		Categories:     categoriesStr,
		CouponType:     req.CouponType,
		DiscountAmount: req.DiscountAmount,
		MinPoint:       req.MinPoint,
		TotalCount:     req.TotalCount,
		ReceiveCount:   0,
		UseCount:       0,
		StartTime:      startTime,
		EndTime:        endTime,
		Status:         1, // 默认正常发放中
	}

	// 6. 写入数据库
	err = models.AddCoupon(coupon)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "创建优惠券失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "优惠券创建成功",
		"data": coupon,
	})
}

// EditCouponHandler 商家编辑优惠券接口 (POST /api/user/shop/coupon/edit)
func EditCouponHandler(c *gin.Context) {
	userIdVal, userExists := c.Get("userId")
	userId, ok := userIdVal.(int64)
	if !userExists || !ok || userId <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 401, "msg": "未登录或用户信息无效"})
		return
	}

	shop, err := models.GetShopByUserID(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "查询商户信息失败: " + err.Error()})
		return
	}
	if shop == nil {
		c.JSON(http.StatusOK, gin.H{"code": 403, "msg": "当前用户未绑定店铺"})
		return
	}

	var req EditCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}
	if len(req.Categories) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "请至少选择一个业态分类"})
		return
	}

	var catStrs []string
	for _, id := range req.Categories {
		catStrs = append(catStrs, fmt.Sprintf("%d", id))
	}

	loc, _ := time.LoadLocation("Local")
	startTime, err1 := time.ParseInLocation("2006-01-02 15:04:05", req.StartTime, loc)
	endTime, err2 := time.ParseInLocation("2006-01-02 15:04:05", req.EndTime, loc)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "时间格式错误，请使用 YYYY-MM-DD HH:mm:ss"})
		return
	}

	coupon := &models.Coupon{
		Id:             req.Id,
		ShopId:         shop.ID,
		UserId:         userId,
		Title:          req.Title,
		Categories:     strings.Join(catStrs, ","),
		CouponType:     req.CouponType,
		DiscountAmount: req.DiscountAmount,
		MinPoint:       req.MinPoint,
		TotalCount:     req.TotalCount,
		StartTime:      startTime,
		EndTime:        endTime,
	}

	if affected := models.EditCouponByShop(coupon); affected == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 404, "msg": "优惠券不存在或无权编辑"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "优惠券编辑成功",
		"data": coupon,
	})
}
