package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"tiyu/models"
	"tiyu/services"

	"github.com/gin-gonic/gin"
)

// UploadImageHandler 上传图片到默认云存储
func UploadImageHandler(c *gin.Context) {
	// 1. 获取图片文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "请选择要上传的图片"})
		return
	}

	// 2. 查询默认云存储配置
	storageConfig, err := models.GetDefaultStorageConfig()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "系统未设定默认云存储或配置无效"})
		return
	}

	// 3. 产生云端的路径/文件名 (例如: uploads/20260831/1693468800_demo.png)
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), ext)
	savePath := fmt.Sprintf("uploads/%s/%s", time.Now().Format("20060102"), filename)

	// 4. 实例化对应的云存储驱动
	driver, err := services.NewStorageDriver(storageConfig.Provider, storageConfig.Config)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	// 5. 执行云端上传
	fileUrl, err := driver.Upload(file, savePath)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "云存储上传失败: " + err.Error()})
		return
	}

	// 6. 返回图片的完整网络 CDN 链接
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "上传成功",
		"data": gin.H{
			"url":      fileUrl,
			"provider": storageConfig.Provider,
			"name":     file.Filename,
		},
	})
}
