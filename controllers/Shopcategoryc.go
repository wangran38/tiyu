package controllers

import (
	"tiyu/models"

	"github.com/gin-gonic/gin"
)

// ShopCategorySearch 分类查询请求体
type ShopCategorySearch struct {
	Name     string `json:"name"`
	Code     string `json:"code"`
	ParentID int64  `json:"parent_id"`
	Status   int8   `json:"status"`
	Limit    int    `json:"limit"`
	Page     int    `json:"page"`
	Order    string `json:"sort"`
}

// GetShopCategoryList 获取商家分类列表
func GetShopCategoryList(c *gin.Context) {
	var searchdata ShopCategorySearch
	c.BindJSON(&searchdata)

	limit := searchdata.Limit
	if limit <= 0 {
		limit = 10
	}
	page := searchdata.Page
	if page <= 0 {
		page = 1
	}

	search := models.ShopCategory{
		Name:     searchdata.Name,
		Code:     searchdata.Code,
		ParentID: searchdata.ParentID,
		Status:   searchdata.Status,
	}

	listdata, err := models.GetShopCategoryList(limit, page, &search, searchdata.Order)
	if err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "获取商家分类列表失败",
			"data":    err.Error(),
		})
		return
	}

	listnum := models.GetShopCategoryTotal(&search)

	result := make(map[string]interface{})
	result["page"] = page
	result["limit"] = limit
	result["totalnum"] = listnum
	result["listdata"] = listdata

	c.JSON(200, gin.H{
		"code":    200,
		"message": "数据获取成功",
		"data":    result,
	})
}

// AddShopCategory 新增商家分类
func AddShopCategory(c *gin.Context) {
	var category models.ShopCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "参数绑定失败",
			"data":    err.Error(),
		})
		return
	}

	if category.Name == "" {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "分类名称不能为空",
			"data":    "",
		})
		return
	}

	err := models.AddShopCategory(&category)
	if err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "新增失败",
			"data":    err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "新增成功",
		"data":    "",
	})
}

// EditShopCategory 修改商家分类
func EditShopCategory(c *gin.Context) {
	var category models.ShopCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "参数绑定失败",
			"data":    err.Error(),
		})
		return
	}

	if category.ID == 0 {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "缺少 ID",
			"data":    "",
		})
		return
	}

	err := models.UpdateShopCategory(category.ID, &category)
	if err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "修改失败",
			"data":    err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "修改成功",
		"data":    "",
	})
}

// DelShopCategory 删除商家分类
func DelShopCategory(c *gin.Context) {
	var req struct {
		ID uint64 `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "参数绑定失败",
			"data":    err.Error(),
		})
		return
	}

	if req.ID == 0 {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "缺少 ID",
			"data":    "",
		})
		return
	}

	err := models.DeleteShopCategory(req.ID)
	if err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "删除失败",
			"data":    err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "删除成功",
		"data":    "",
	})
}
