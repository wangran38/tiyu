package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	imageUploadDir = "uploads"
	maxImageSize   = 10 << 20
)

// UploadQwenOCR 上传票据图片，保存图片地址后调用 Qwen OCR 返回识别结果。
func UploadQwenOCR(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 201, "message": "请使用 file 字段上传图片", "data": err.Error()})
		return
	}
	if fileHeader.Size <= 0 || fileHeader.Size > maxImageSize {
		c.JSON(http.StatusBadRequest, gin.H{"code": 201, "message": "图片大小必须在 1 字节到 10MB 之间", "data": ""})
		return
	}

	extension := strings.ToLower(filepath.Ext(fileHeader.Filename))
	allowedExtensions := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowedExtensions[extension] {
		c.JSON(http.StatusBadRequest, gin.H{"code": 201, "message": "仅支持 jpg、jpeg、png、webp 图片", "data": ""})
		return
	}

	if err := os.MkdirAll(imageUploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 201, "message": "创建图片目录失败", "data": err.Error()})
		return
	}

	fileName, err := newImageFileName(extension)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 201, "message": "生成图片名称失败", "data": err.Error()})
		return
	}
	filePath := filepath.Join(imageUploadDir, fileName)
	if err := c.SaveUploadedFile(fileHeader, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 201, "message": "保存图片失败", "data": err.Error()})
		return
	}

	imageURL := buildImageURL(c, fileName)
	imageDataURL, err := buildImageDataURL(filePath, extension)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    201,
			"message": "图片已保存，但读取图片失败",
			"data": gin.H{
				"image_url": imageURL,
				"error":     err.Error(),
			},
		})
		return
	}
	apiKey := os.Getenv("DASHSCOPE_API_KEY")
	if apiKey == "" {
		apiKey = DashScopeAPIKey
	}
	result, err := ParseTicketVision(c.Request.Context(), apiKey, imageDataURL, c.PostForm("sample_image_url"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"code":    201,
			"message": "图片已保存，但门票识别失败",
			"data": gin.H{
				"image_url": imageURL,
				"error":     err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "图片上传并识别成功",
		"data": gin.H{
			"image_url": imageURL,
			"ocr":       result,
		},
	})
}

func newImageFileName(extension string) (string, error) {
	randomBytes := make([]byte, 12)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), hex.EncodeToString(randomBytes), extension), nil
}

func buildImageDataURL(filePath, extension string) (string, error) {
	imageBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("读取图片文件失败: %w", err)
	}

	mimeType := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".webp": "image/webp",
	}[extension]
	if mimeType == "" {
		return "", fmt.Errorf("不支持的图片类型: %s", extension)
	}

	return "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(imageBytes), nil
}

func buildImageURL(c *gin.Context, fileName string) string {
	baseURL := strings.TrimRight(os.Getenv("PUBLIC_BASE_URL"), "/")
	if baseURL == "" {
		scheme := "http"
		if strings.EqualFold(c.Request.Header.Get("X-Forwarded-Proto"), "https") || c.Request.TLS != nil {
			scheme = "https"
		}
		host := c.Request.Header.Get("X-Forwarded-Host")
		if host == "" {
			host = c.Request.Host
		}
		baseURL = scheme + "://" + host
	}
	return baseURL + "/uploads/" + fileName
}
