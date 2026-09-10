package api

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"tiyu/models"
	"tiyu/services"

	"github.com/gin-gonic/gin"
)

// UploadImagesHandler 批量上传多张图片到默认云存储
func UploadImagesHandler(c *gin.Context) {
	uploadImages(c)
}

// PublicUploadImagesHandler 提供给外部系统的公开图片上传接口，不依赖会员登录。
func PublicUploadImagesHandler(c *gin.Context) {
	uploadImages(c)
}

func uploadImages(c *gin.Context) {
	// 1. 获取 multipart 表单
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "请选择要上传的图片"})
		return
	}

	// 获取前端传过来的多图数组（同时兼容 "files" 和 "file" 字段名）
	files := form.File["files"]
	if len(files) == 0 {
		files = form.File["file"]
	}

	if len(files) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "未找到要上传的图片文件"})
		return
	}

	// 2. 查询默认云存储配置
	storageConfig, err := models.GetDefaultStorageConfig()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "系统未设定默认云存储或配置无效"})
		return
	}

	// 3. 实例化对应的云存储驱动
	driver, err := services.NewStorageDriver(storageConfig.Provider, storageConfig.Config)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	// 4. 循环遍历并逐个上传
	var uploadedUrls []string
	var uploadedFiles []gin.H

	for _, file := range files {
		ext := filepath.Ext(file.Filename)
		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), ext)
		savePath := fmt.Sprintf("uploads/%s/%s", time.Now().Format("20060102"), filename)

		// 执行云端上传
		fileUrl, err := driver.Upload(file, savePath)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "msg": fmt.Sprintf("图片 [%s] 上传失败: %s", file.Filename, err.Error())})
			return
		}

		uploadedUrls = append(uploadedUrls, fileUrl)
		uploadedFiles = append(uploadedFiles, gin.H{
			"url":      fileUrl,
			"provider": storageConfig.Provider,
			"name":     file.Filename,
		})
	}

	// 5. 返回批量上传成功后的数据
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "批量上传成功",
		"data": gin.H{
			"urls":  uploadedUrls,
			"files": uploadedFiles,
		},
	})
}
