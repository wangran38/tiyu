package controllers

import (
	"tiyu/models"

	"github.com/gin-gonic/gin"
)

type SportsCategoryserch struct {
	Name  string `json:"name"`
	Limit int    `json:"limit"`
	Page  int    `json:"page"`
	Order string `json:"sort"`
}

// 列表
func GetSportsCategorylist(c *gin.Context) {
	var searchdata SportsCategoryserch
	c.BindJSON(&searchdata)

	limit := searchdata.Limit
	page := searchdata.Page
	name := searchdata.Name
	order := searchdata.Order

	result := make(map[string]interface{})

	listdata := models.GetSportsCategoryList(limit, page, name, order)
	listnum := models.GetSportsCategoryTotal(name)

	result["page"] = page
	result["totalnum"] = listnum
	result["limit"] = limit

	if listdata == nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "获取赛事分类列表失败",
			"data":    "",
		})
		return
	}

	result["listdata"] = listdata
	c.JSON(200, gin.H{
		"code":    200,
		"message": "数据获取成功",
		"data":    result,
	})
}

// 新增
func AddSportsCategory(c *gin.Context) {
	var m models.SportsCategory
	c.BindJSON(&m)
	err := models.AddSportsCategory(&m)
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

// 修改
func EditSportsCategory(c *gin.Context) {
	var m models.SportsCategory
	c.BindJSON(&m)
	if m.Id == 0 {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "缺少 ID",
			"data":    "",
		})
		return
	}
	err := models.EditSportsCategory(&m)
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

// 删除
func DelSportsCategory(c *gin.Context) {
	var req struct {
		Id int64 `json:"id"`
	}
	c.BindJSON(&req)
	if req.Id == 0 {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "缺少 ID",
			"data":    "",
		})
		return
	}
	outnum := models.DelSportsCategory(req.Id)
	c.JSON(200, gin.H{
		"code":    200,
		"message": "删除成功",
		"data":    outnum,
	})
}
