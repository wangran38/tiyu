package controllers

import (
	"time"

	"tiyu/models"

	"github.com/gin-gonic/gin"
)

type EventSearch struct {
	Title            string     `json:"title"`
	CategoryId       int64      `json:"category_id"`
	RegistrationFrom *time.Time `json:"registration_from"`
	RegistrationTo   *time.Time `json:"registration_to"`
	StartTime        *time.Time `json:"start_time"`
	EndTime          *time.Time `json:"end_time"`
	Limit            int        `json:"limit"`
	Page             int        `json:"page"`
	Order            string     `json:"sort"`
}

// 赛事活动列表
func GetEventlist(c *gin.Context) {
	var searchdata EventSearch
	if err := c.BindJSON(&searchdata); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "参数错误",
			"data":    err.Error(),
		})
		return
	}

	result := make(map[string]interface{})
	listdata := models.GetEventList(searchdata.Limit, searchdata.Page, searchdata.Title, searchdata.CategoryId, searchdata.RegistrationFrom, searchdata.RegistrationTo, searchdata.StartTime, searchdata.EndTime, searchdata.Order)
	listnum := models.GetEventTotal(searchdata.Title, searchdata.CategoryId, searchdata.RegistrationFrom, searchdata.RegistrationTo, searchdata.StartTime, searchdata.EndTime)

	result["page"] = searchdata.Page
	result["totalnum"] = listnum
	result["limit"] = searchdata.Limit
	result["listdata"] = listdata

	c.JSON(200, gin.H{
		"code":    200,
		"message": "数据获取成功",
		"data":    result,
	})
}

// 新增赛事活动
func AddEvent(c *gin.Context) {
	var event models.Event
	if err := c.BindJSON(&event); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "参数错误",
			"data":    err.Error(),
		})
		return
	}

	if err := models.AddEvent(&event); err != nil {
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
		"data":    event,
	})
}

// 修改赛事活动
func EditEvent(c *gin.Context) {
	var event models.Event
	if err := c.BindJSON(&event); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "参数错误",
			"data":    err.Error(),
		})
		return
	}
	if event.Id == 0 {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "缺少 ID",
			"data":    "",
		})
		return
	}

	if err := models.EditEvent(&event); err != nil {
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
		"data":    event,
	})
}

// 删除赛事活动
func DelEvent(c *gin.Context) {
	var req struct {
		Id int64 `json:"id"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "参数错误",
			"data":    err.Error(),
		})
		return
	}
	if req.Id == 0 {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "缺少 ID",
			"data":    "",
		})
		return
	}

	outnum := models.DelEvent(req.Id)
	c.JSON(200, gin.H{
		"code":    200,
		"message": "删除成功",
		"data":    outnum,
	})
}
