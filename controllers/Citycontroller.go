package controllers

import (
	"tiyu/models"

	"github.com/gin-gonic/gin"
)

type Cityserch struct {
	Name  string `json:"name"`
	Limit int    `json:"limit"`
	Page  int    `json:"page"`
	Order string `json:"sort"`
}

// 地区列表
func Getcitylist(c *gin.Context) {
	var searchdata Cityserch
	c.BindJSON(&searchdata)

	limit := searchdata.Limit
	page := searchdata.Page
	name := searchdata.Name
	order := searchdata.Order

	result := make(map[string]interface{})

	listdata := models.GetCityList(limit, page, name, order)
	listnum := models.GetCityTotal(name)

	result["page"] = page
	result["totalnum"] = listnum
	result["limit"] = limit

	if listdata == nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "获取地区列表失败",
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

// 新增地区
func Addcity(c *gin.Context) {
	var city models.City
	c.BindJSON(&city)
	err := models.AddCity(&city)
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

// 修改地区
func Editcity(c *gin.Context) {
	var city models.City
	c.BindJSON(&city)
	if city.Id == 0 {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "缺少 ID",
			"data":    "",
		})
		return
	}
	err := models.EditCity(&city)
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

// 删除地区
func Delcity(c *gin.Context) {
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
	outnum := models.DelCity(req.Id)
	c.JSON(200, gin.H{
		"code":    200,
		"message": "删除成功",
		"data":    outnum,
	})
}
