package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"tiyu/global"
	"tiyu/models"
	"tiyu/utils"

	"github.com/gin-gonic/gin"
)

// ApplyShopReq C端申请商家入驻请求体

// ApplyShopReq C端申请商家入驻请求体
type ApplyShopReq struct {
	Name         string  `json:"name" binding:"required"`          // 门店名称
	CategoryID   int64   `json:"category_id" binding:"required"`   // 分类ID
	CityID       int64   `json:"city_id" binding:"required"`       // 城市ID
	ContactName  string  `json:"contact_name" binding:"required"`  // 负责人
	ContactPhone string  `json:"contact_phone" binding:"required"` // 负责人电话
	ServicePhone string  `json:"service_phone"`                    // 客服电话
	Logo         string  `json:"logo"`                             // Logo图片
	CoverImages  string  `json:"cover_images"`                     // 门头照/环境图JSON数组
	Address      string  `json:"address" binding:"required"`       // 详细地址
	Longitude    float64 `json:"longitude"`                        // 经度
	Latitude     float64 `json:"latitude"`                         // 纬度
	Description  string  `json:"description"`                      // 商家简介 (长文本)
	Discounts    string  `json:"discounts"`                        // 优惠内容 (新增)
}

// SubmitShopApplication C端用户提交/重新提交商家申请
// @Router /api/user/shop-apply [post]
// SubmitShopApplication C端用户提交/重新提交商家申请
func SubmitShopApplication(c *gin.Context) {
	// 1. 获取当前登录用户的 UserID
	userIDVal, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusOK, gin.H{"code": 401, "msg": "未获取到登录信息"})
		return
	}
	userID := userIDVal.(int64)

	var req ApplyShopReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "请填写完整的申请资料"})
		return
	}

	// 2. 检查用户当前是否有生效中或审核中的店铺
	var existShop models.Shop
	has, err := global.Dorm.Where("user_id = ?", userID).Get(&existShop)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "数据库查询异常"})
		return
	}

	if has {
		if existShop.Status == 0 {
			c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "您已提交过申请，正在等待后台审核，请勿重复提交"})
			return
		}
		if existShop.Status == 1 || existShop.Status == 2 {
			c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "您已是开店商家，无需重复申请"})
			return
		}
		if existShop.Status == 3 {
			c.JSON(http.StatusOK, gin.H{"code": 403, "msg": "您的店铺已被冻结，请联系客服处理"})
			return
		}
	}

	// 3. 生成包含 UserID 的商家业务编号
	// 格式：M + YYYYMMDD + UserID + 4位随机码 (例如：M2026090312345678)
	merchantNo := fmt.Sprintf("M%s%d%s", time.Now().Format("20060102"), userID, utils.GenerateRandomCode(4))

	shopData := models.Shop{
		MerchantNo:   merchantNo,
		UserID:       userID,
		CategoryID:   req.CategoryID,
		CityID:       req.CityID,
		Status:       0, // 0: 待审核
		Name:         strings.TrimSpace(req.Name),
		Logo:         req.Logo,
		CoverImages:  req.CoverImages,
		ContactName:  strings.TrimSpace(req.ContactName),
		ContactPhone: strings.TrimSpace(req.ContactPhone),
		ServicePhone: req.ServicePhone,
		Address:      strings.TrimSpace(req.Address),
		Longitude:    req.Longitude,
		Latitude:     req.Latitude,
		Description:  req.Description,
		Discounts:    req.Discounts,
	}

	// 4. 如果是被驳回 (Status == 4) 则更新记录重新提交，否则新增
	if has && existShop.Status == 4 {
		shopData.ID = existShop.ID
		_, err = global.Dorm.ID(existShop.ID).AllCols().Update(&shopData)
	} else {
		err = models.AddShop(&shopData)
	}

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "提交申请失败，请稍后重试"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "入驻申请提交成功，请等待管理员审核",
		"data": gin.H{
			"shop_id":     shopData.ID,
			"merchant_no": shopData.MerchantNo,
		},
	})
}

// GetShopApplicationStatus 查询当前登录用户的商家申请/店铺状态
// @Router /api/user/shop-apply/status [get]
func GetShopApplicationStatus(c *gin.Context) {
	userIDVal, _ := c.Get("userId")
	userID := userIDVal.(int64)

	var shop models.Shop
	has, err := global.Dorm.Where("user_id = ?", userID).Get(&shop)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "查询状态失败"})
		return
	}

	if !has {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "获取成功",
			"data": gin.H{"has_apply": false, "status": -1}, // 未申请过
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "获取成功",
		"data": gin.H{
			"has_apply":   true,
			"shop_id":     shop.ID,
			"merchant_no": shop.MerchantNo,
			"name":        shop.Name,
			"status":      shop.Status, // 0:待审核 1:营业中 2:休息中 3:已冻结 4:审核驳回
			"created_at":  shop.CreatedAt,
		},
	})
}
