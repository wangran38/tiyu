package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"
	"tiyu/models"
	ticketstore "tiyu/storage/pebble"

	"github.com/gin-gonic/gin"
)

// OCRRequest 公开票根识别请求参数。
// UserImageURL: 用户上传的票根图片 URL（必填）。
// SampleImageURL: 参考样例图片 URL（可选，用于辅助比对模板）。
// UserID: 当前提交识别的用户 ID（可选，若提供则写入公共识别记录表）。
type OCRRequest struct {
	UserID         uint64 `form:"user_id"`
	UserImageURL   string `form:"user_image_url" binding:"required"`
	SampleImageURL string `form:"sample_image_url"`
}

// RecognizeTicketByURLHandler 提供给外部系统的公开票根识别接口，不依赖会员登录。
func RecognizeTicketByURLHandler(c *gin.Context) {
	var request OCRRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    201,
			"message": "参数错误",
			"data":    err.Error(),
		})
		return
	}

	apiKey := os.Getenv("DASHSCOPE_API_KEY")
	if apiKey == "" {
		apiKey = DashScopeAPIKey
	}
	result, err := ParseTicketVision(c.Request.Context(), apiKey, request.UserImageURL, request.SampleImageURL)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"code":    201,
			"message": "门票识别失败",
			"data":    err.Error(),
		})
		return
	}

	if request.UserID > 0 {
		ocrJSON, err := json.Marshal(result)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    201,
				"message": "识别结果序列化失败",
				"data":    err.Error(),
			})
			return
		}

		if _, err := ticketstore.SaveIfAbsent(ticketstore.Record{
			TicketCategory: result.TicketCategory,
			TicketSN:       result.Data.TicketSN,
			Seat:           result.Data.Seat,
			HolderName:     result.Data.HolderName,
			EventDate:      result.Data.EventDate,
			UserImageURL:   request.UserImageURL,
			OCRJSON:        ocrJSON,
			CreatedAt:      time.Now(),
		}); err != nil && !errors.Is(err, ticketstore.ErrDuplicate) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    201,
				"message": "写入 Pebble 数据库失败",
				"data":    err.Error(),
			})
			return
		}

		publicRecord := &models.PublicTicketRecognition{
			UserID:         request.UserID,
			ChannelType:    result.ChannelType,
			TicketCategory: result.TicketCategory,
			ThirdPartyName: result.ThirdPartyName,
			UserImageURL:   request.UserImageURL,
			PHash:          "",
			TicketSN:       result.Data.TicketSN,
			Title:          result.Data.Title,
			HolderName:     result.Data.HolderName,
			EventDate:      result.Data.EventDate,
			Seat:           result.Data.Seat,
			Amount:         result.Data.Amount,
			Confidence:     result.Confidence,
			IsHandwritten:  result.IsHandwritten,
			OCRRawJSON:     string(ocrJSON),
			RecognizedAt:   time.Now(),
		}
		if err := models.AddPublicTicketRecognition(publicRecord); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    201,
				"message": "写入 MySQL 公共识别表失败",
				"data":    err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "门票识别成功",
		"data":    result,
	})
}

// TestQwenOCR 用于通过 HTTP 测试 Qwen 票据识别，不保存领券记录。
func TestQwenOCR(c *gin.Context) {
	RecognizeTicketByURLHandler(c)
}
