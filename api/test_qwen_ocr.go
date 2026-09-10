package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"tiyu/models"
	"tiyu/services"
	ticketstore "tiyu/storage/pebble"

	"github.com/gin-gonic/gin"
)

// ===== 1. 配置信息 =====
const (
	DashScopeAPIKey = "sk-ws-H.EYPRLMR.Qoqy.MEUCIHXVpop30uWLoU9ddJYGVypi9tP8wTIEzgF4ONU_gqIRAiEA5ilehZMGMnvVvbmLnTJfg9cT8rXiUEXXlzpJDY3pHdY"
	QwenModel       = "qwen-vl-max"
)

// ===== 2. 结构体定义 =====

type ContentItem struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

type ImageURL struct {
	URL string `json:"url"`
}

type Message struct {
	Role    string        `json:"role"`
	Content []ContentItem `json:"content"`
}

type QwenReq struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type OCRResult struct {
	ChannelType    string  `json:"channel_type"`
	TicketCategory string  `json:"ticket_category"`
	ThirdPartyName string  `json:"third_party_name"`
	IsValid        bool    `json:"is_valid"`
	IsHandwritten  bool    `json:"is_handwritten"`
	IsSameTemplate bool    `json:"is_same_template"`
	Confidence     float64 `json:"confidence"`
	Data           struct {
		Title      string  `json:"title"`
		HolderName string  `json:"holder_name"`
		EventDate  string  `json:"event_date"`
		TicketSN   string  `json:"ticket_sn"`
		Seat       string  `json:"seat"`
		Amount     float64 `json:"amount"`
	} `json:"data"`
	RejectReason string `json:"reject_reason"`
}

type VerifyTicketRequest struct {
	UserImgURL   string `json:"user_img_url" binding:"required"` // 用户上传的票根图片（来自 common/upload 返回的腾讯云 URL）
	TemplateID   int64  `json:"template_id"`                     // 模版ID（可选）
	SampleImgURL string `json:"sample_img_url"`                  // 标杆样例图URL（可选）
}

// ===== 3. AI 解析核心函数 =====
func ParseTicketVision(ctx context.Context, apiKey, userImgURL, sampleImgURL string) (*OCRResult, error) {
	prompt := `你是一个全能的文旅及交通票据智能审核专家。请分析上传的图片：

【通道分类规则】
- CHANNEL_A (标准票): 12306火车票/行程单、机票登机牌/行程单、国家统一增值税发票。
- CHANNEL_B (非标票): 大麦/猫眼/美团等电子票截图、景区门票/核销码、体育赛事入场券、演出门票、乡村/手撕/盖章票据。
- 体育赛事、演出、展览等有明确活动名称、日期或入场信息的票，即使是纸质票或拍摄方向旋转，也属于有效的第三方票，不要判定为 INVALID。

【提取要求】
1. 判断 channel_type ("CHANNEL_A" 或 "CHANNEL_B")
2. 判断 ticket_category：火车票返回“火车票”，飞机票返回“飞机票”，其他票据返回票面上的第三方平台或活动名称。同时将第三方平台或活动名称填入 third_party_name；火车票和飞机票的该字段为空。请先自动纠正图片的横竖方向后再识别文字。
3. 提取字段：
   - title: 票头或活动/景区名称
   - holder_name: 持票人/乘客姓名 (若无则为空)
   - event_date: 格式 YYYY-MM-DD hh:mm:ss (如 2026-05-01 10:00:00)
   - ticket_sn: 核心流水号、车次(如G1234)、航班号、订单号
   - amount: 数字金额 (如 100.50)
4. 若传入了两张图（第一张为后台参考模版，第二张为用户上传），请对比两者版式布局是否一致 (is_same_template: true/false)。
5. 判断票据是否为手绘、手写或非印刷制作 (is_handwritten)。只要票面主要内容是手绘、手写或非印刷填写形成的票据，就判定为无效票据 (is_valid: false)，并在 reject_reason 中说明原因；正常印刷票据即使有手写签名或少量手写补充，也不算手绘票。
6. 提取票面座位信息并填入 data.seat，例如“3排8座”“A区12排5号”；没有座位信息时返回空字符串。

请严格只返回以下 JSON 格式，绝不要添加任何 Markdown 格式符或额外解释：
{
  "channel_type": "CHANNEL_A",
  "ticket_category": "火车票",
  "third_party_name": "",
  "is_valid": true,
	"is_handwritten": false,
  "is_same_template": true,
  "confidence": 0.98,
  "data": {
    "title": "高铁电子客票",
    "holder_name": "张三",
    "event_date": "2026-05-01 10:00:00",
    "ticket_sn": "G1234",
	"seat": "3排8座",
    "amount": 150.50
  },
  "reject_reason": ""
}`

	var contents []ContentItem
	if sampleImgURL != "" {
		contents = append(contents, ContentItem{
			Type:     "image_url",
			ImageURL: &ImageURL{URL: sampleImgURL},
		})
	}
	contents = append(contents, ContentItem{
		Type:     "image_url",
		ImageURL: &ImageURL{URL: userImgURL},
	})
	contents = append(contents, ContentItem{
		Type: "text",
		Text: prompt,
	})

	reqBody := QwenReq{
		Model: QwenModel,
		Messages: []Message{
			{Role: "user", Content: contents},
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("json marshal error: %w", err)
	}

	client := &http.Client{Timeout: 45 * time.Second}
	var resp *http.Response
	for attempt := 0; attempt < 3; attempt++ {
		httpReq, requestErr := http.NewRequestWithContext(ctx, "POST", "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions", bytes.NewReader(jsonBytes))
		if requestErr != nil {
			return nil, requestErr
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err = client.Do(httpReq)
		if err == nil {
			break
		}
		var networkErr net.Error
		if !errors.As(err, &networkErr) || attempt == 2 {
			return nil, fmt.Errorf("http request failed: %w", err)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(attempt+1) * time.Second):
		}
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api responded code %d, body: %s", resp.StatusCode, string(respBytes))
	}

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBytes, &apiResp); err != nil || len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("unmarshal api choice failed: %w", err)
	}

	rawJSON := apiResp.Choices[0].Message.Content
	rawJSON = strings.ReplaceAll(rawJSON, "```json", "")
	rawJSON = strings.ReplaceAll(rawJSON, "```", "")
	rawJSON = strings.TrimSpace(rawJSON)

	var result OCRResult
	if err := json.Unmarshal([]byte(rawJSON), &result); err != nil {
		return nil, fmt.Errorf("unmarshal struct failed: %w, raw text: %s", err, rawJSON)
	}
	if result.IsHandwritten {
		result.IsValid = false
		if result.RejectReason == "" {
			result.RejectReason = "票据为手绘、手写或非印刷票据"
		}
	}

	switch strings.ToUpper(strings.TrimSpace(result.TicketCategory)) {
	case "TRAIN":
		result.TicketCategory = "火车票"
	case "FLIGHT":
		result.TicketCategory = "飞机票"
	case "THIRD_PARTY", "SCENIC", "LOCAL_EVENT":
		if result.ThirdPartyName == "" {
			result.ThirdPartyName = result.Data.Title
		}
		if result.ThirdPartyName != "" {
			result.TicketCategory = result.ThirdPartyName
		}
	}

	return &result, nil
}

// ===== 4. 票根核验与权益发放接口 =====
// ===== 4. 票根核验接口（测试版：暂去数据库操作） =====
// ===== 4. 票根核验与权益发放接口（完整合并版：后端自动接文件、上传云存储并交由AI核验） =====
func VerifyTicketHandler(c *gin.Context) {
	userIDValue, exists := c.Get("userId")
	userID, ok := userIDValue.(int64)
	if !exists || !ok || userID <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 401, "msg": "请先登录"})
		return
	}

	// 如果后续需要恢复登录拦截校验，可取消下方注释
	/*
		userIdVal, userExists := c.Get("user_id")
		if !userExists {
			c.JSON(http.StatusOK, gin.H{"code": 401, "msg": "未登录或用户信息不存在"})
			return
		}
		var userId int64
		switch v := userIdVal.(type) {
		case int64:
			userId = v
		case int:
			userId = int64(v)
		case float64:
			userId = int64(v)
		default:
			c.JSON(http.StatusOK, gin.H{"code": 401, "msg": "用户ID类型解析异常"})
			return
		}
	*/

	// 1. 获取前端传过来的图片文件（同时兼容 "file" 和 "files" 字段名）
	file, err := c.FormFile("file")
	if err != nil {
		file, err = c.FormFile("files")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "未找到需要核验的票根图片文件"})
			return
		}
	}
	imageHash, err := perceptualHashUploadedFile(file)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "票根图片读取失败: " + err.Error()})
		return
	}

	// 2. 获取其他可选表单参数（如 template_id, sample_img_url）
	sampleImgURL := c.PostForm("sample_img_url")

	// 3. 查询默认云存储配置
	storageConfig, err := models.GetDefaultStorageConfig()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "系统未设定默认云存储或配置无效"})
		return
	}

	// 4. 实例化对应的云存储驱动
	driver, err := services.NewStorageDriver(storageConfig.Provider, storageConfig.Config)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "云存储驱动初始化失败: " + err.Error()})
		return
	}

	// 5. 生成云端保存路径并上传
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), ext)
	savePath := fmt.Sprintf("uploads/ticket_verify/%s/%s", time.Now().Format("20060102"), filename)

	userImgURL, err := driver.Upload(file, savePath)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "票根图片上传云端失败: " + err.Error()})
		return
	}

	// 6. 获取 AI API Key
	apiKey := os.Getenv("DASHSCOPE_API_KEY")
	if apiKey == "" {
		apiKey = DashScopeAPIKey
	}

	// 7. 调用 AI 视觉解析核心函数（传入后端刚刚生成的云存储公网 URL）
	ocrRes, err := ParseTicketVision(c.Request.Context(), apiKey, userImgURL, sampleImgURL)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "票根AI识别异常: " + err.Error()})
		return
	}

	if !ocrRes.IsValid {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "票根核验未通过: " + ocrRes.RejectReason,
			"data": gin.H{
				"user_img_url": userImgURL,
				"ocr_result":   ocrRes,
			},
		})
		return
	}

	ocrJSON, err := json.Marshal(ocrRes)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "票根识别结果序列化失败: " + err.Error()})
		return
	}
	duplicateKind, err := ticketstore.SaveIfAbsent(ticketstore.Record{
		TicketCategory: ocrRes.TicketCategory,
		TicketSN:       ocrRes.Data.TicketSN,
		Seat:           ocrRes.Data.Seat,
		HolderName:     ocrRes.Data.HolderName,
		EventDate:      ocrRes.Data.EventDate,
		ImageHash:      imageHash,
		UserImageURL:   userImgURL,
		OCRJSON:        ocrJSON,
		CreatedAt:      time.Now(),
	})
	if errors.Is(err, ticketstore.ErrDuplicate) {
		message := "对不起，你拍摄的票根已经存在重复"
		if duplicateKind == "ticket_sn" {
			message = "已经有人使用该票根"
		}
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": message})
		return
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "票根识别数据保存失败: " + err.Error()})
		return
	}

	recognizedAt := time.Now()
	memberTicket := &models.MemberTicket{
		UserID:         uint64(userID),
		ChannelType:    ocrRes.ChannelType,
		TicketCategory: ocrRes.TicketCategory,
		UserImageURL:   userImgURL,
		PHash:          imageHash,
		TicketSN:       ocrRes.Data.TicketSN,
		Title:          ocrRes.Data.Title,
		HolderName:     ocrRes.Data.HolderName,
		EventDate:      ocrRes.Data.EventDate,
		Seat:           ocrRes.Data.Seat,
		Amount:         ocrRes.Data.Amount,
		Confidence:     ocrRes.Confidence,
		IsHandwritten:  ocrRes.IsHandwritten,
		OCRRawJSON:     string(ocrJSON),
		ExchangeStatus: 0,
		RecognizedAt:   recognizedAt,
	}
	if err := models.AddMemberTicket(memberTicket); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "会员票根记录保存失败: " + err.Error()})
		return
	}

	// 8. 返回核验成功响应
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "【测试模式】票根AI核验成功（后端已自动上传云存储并识别）",
		"data": gin.H{
			"user_img_url": userImgURL,
			"ocr_result":   ocrRes,
		},
	})
}

func perceptualHashUploadedFile(file *multipart.FileHeader) (string, error) {
	reader, err := file.Open()
	if err != nil {
		return "", err
	}
	defer reader.Close()

	decoded, _, err := image.Decode(reader)
	if err != nil {
		return "", fmt.Errorf("decode image: %w", err)
	}
	if decoded.Bounds().Empty() {
		return "", fmt.Errorf("image has empty dimensions")
	}

	var hash uint64
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			left := grayscaleAt(decoded, x, y, 9, 8)
			right := grayscaleAt(decoded, x+1, y, 9, 8)
			hash <<= 1
			if left > right {
				hash |= 1
			}
		}
	}
	return strconv.FormatUint(hash, 16), nil
}

func grayscaleAt(source image.Image, x, y, width, height int) uint8 {
	bound := source.Bounds()
	sourceX := bound.Min.X + x*(bound.Dx()-1)/(width-1)
	sourceY := bound.Min.Y + y*(bound.Dy()-1)/(height-1)
	gray := color.GrayModel.Convert(source.At(sourceX, sourceY)).(color.Gray)
	return gray.Y
}
