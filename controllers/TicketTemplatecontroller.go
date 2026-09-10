package controllers

import (
	"tiyu/models"

	"github.com/gin-gonic/gin"
)

type TicketTemplateSearch struct {
	TemplateName string `json:"template_name"`
	Limit        int    `json:"limit"`
	Page         int    `json:"page"`
	Order        string `json:"sort"`
}

func GetTicketTemplatelist(c *gin.Context) {
	var searchdata TicketTemplateSearch
	if err := c.BindJSON(&searchdata); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "参数错误",
			"data":    err.Error(),
		})
		return
	}

	listdata := models.GetTicketTemplateList(searchdata.Limit, searchdata.Page, searchdata.TemplateName, searchdata.Order)
	result := map[string]interface{}{
		"page":     searchdata.Page,
		"totalnum": models.GetTicketTemplateTotal(searchdata.TemplateName),
		"limit":    searchdata.Limit,
		"listdata": listdata,
	}
	c.JSON(200, gin.H{
		"code":    200,
		"message": "数据获取成功",
		"data":    result,
	})
}

func AddTicketTemplate(c *gin.Context) {
	var template models.TicketTemplate
	if err := c.BindJSON(&template); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "参数错误",
			"data":    err.Error(),
		})
		return
	}
	if err := models.AddTicketTemplate(&template); err != nil {
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
		"data":    template,
	})
}

func EditTicketTemplate(c *gin.Context) {
	var template models.TicketTemplate
	if err := c.BindJSON(&template); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "参数错误",
			"data":    err.Error(),
		})
		return
	}
	if template.ID == 0 {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "缺少 ID",
			"data":    "",
		})
		return
	}
	if err := models.EditTicketTemplate(&template); err != nil {
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
		"data":    template,
	})
}

func DelTicketTemplate(c *gin.Context) {
	var req struct {
		ID uint64 `json:"id"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "参数错误",
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
	c.JSON(200, gin.H{
		"code":    200,
		"message": "删除成功",
		"data":    models.DelTicketTemplate(req.ID),
	})
}
