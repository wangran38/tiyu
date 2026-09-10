package controllers

import (
	"tiyu/models"

	"github.com/gin-gonic/gin"
)

type TicketClaimSearch struct {
	UserID         uint64 `json:"user_id"`
	ChannelType    string `json:"channel_type"`
	TicketCategory string `json:"ticket_category"`
	TemplateID     uint64 `json:"template_id"`
	ClaimStatus    int8   `json:"claim_status"`
	TicketSN       string `json:"ticket_sn"`
	PHash          string `json:"p_hash"`
	Keyword        string `json:"keyword"`
	Limit          int    `json:"limit"`
	Page           int    `json:"page"`
	Order          string `json:"sort"`
}

// 门票识别提交：保存图片地址和 OCR 服务返回的结构化识别结果。
func RecognizeTicket(c *gin.Context) {
	var claim models.TicketClaim
	if err := c.BindJSON(&claim); err != nil {
		c.JSON(200, gin.H{"code": 201, "message": "参数错误", "data": err.Error()})
		return
	}
	if claim.UserID == 0 || claim.UserImageURL == "" || claim.PHash == "" || claim.TicketSN == "" {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "缺少 user_id、user_image_url、p_hash 或 ticket_sn",
			"data":    "",
		})
		return
	}
	if claim.ClaimStatus == 0 {
		claim.ClaimStatus = 1
	}
	if err := models.AddTicketClaim(&claim); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "门票识别记录保存失败，可能已重复提交",
			"data":    err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"code":    200,
		"message": "门票识别成功",
		"data":    claim,
	})
}

func GetTicketClaimlist(c *gin.Context) {
	getTicketClaimList(c, "数据获取成功")
}

// 识别列表和防伪列表共用同一套筛选数据，保证后台两个视图口径一致。
func GetTicketRecognitionlist(c *gin.Context) {
	getTicketClaimList(c, "识别记录获取成功")
}

func GetTicketFraudlist(c *gin.Context) {
	getTicketClaimList(c, "防伪记录获取成功")
}

func getTicketClaimList(c *gin.Context, message string) {
	var searchdata TicketClaimSearch
	if err := c.BindJSON(&searchdata); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "参数错误",
			"data":    err.Error(),
		})
		return
	}

	search := &models.TicketClaim{
		UserID:         searchdata.UserID,
		ChannelType:    searchdata.ChannelType,
		TicketCategory: searchdata.TicketCategory,
		TemplateID:     searchdata.TemplateID,
		ClaimStatus:    searchdata.ClaimStatus,
		TicketSN:       searchdata.TicketSN,
		PHash:          searchdata.PHash,
		Title:          searchdata.Keyword,
	}
	result := map[string]interface{}{
		"page":     searchdata.Page,
		"totalnum": models.GetTicketClaimTotal(search),
		"limit":    searchdata.Limit,
		"listdata": models.GetTicketClaimList(searchdata.Limit, searchdata.Page, search, searchdata.Order),
	}
	c.JSON(200, gin.H{
		"code":    200,
		"message": message,
		"data":    result,
	})
}
