package services

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// COSConfig 匹配数据库中存储的 JSON 参数
type COSConfig struct {
	Bucket    string `json:"bucket"`
	Domain    string `json:"domain"`
	Region    string `json:"region"`
	SecretID  string `json:"secret_id"`
	SecretKey string `json:"secret_key"`
}

type COSDriver struct {
	cfg COSConfig
}

// NewCOSDriver 创建腾讯云 COS 驱动实例
func NewCOSDriver(configJson string) (*COSDriver, error) {
	var cfg COSConfig
	if err := json.Unmarshal([]byte(configJson), &cfg); err != nil {
		return nil, fmt.Errorf("解析 COS 配置失败: %v", err)
	}
	return &COSDriver{cfg: cfg}, nil
}

// Upload 实现 StorageDriver 接口的 Upload 方法
func (d *COSDriver) Upload(file *multipart.FileHeader, savePath string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("打开本地文件失败: %v", err)
	}
	defer src.Close()

	// 1. 构建腾讯云官方标准的通讯域名 (解决 invalid bucket format 报错)
	officialBaseURL := fmt.Sprintf("https://%s.cos.%s.myqcloud.com", d.cfg.Bucket, d.cfg.Region)

	u, err := url.Parse(officialBaseURL)
	if err != nil {
		return "", fmt.Errorf("构建 COS 官方 URL 失败: %v", err)
	}

	// 2. 初始化 COS 客户端
	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  d.cfg.SecretID,
			SecretKey: d.cfg.SecretKey,
		},
	})

	// 3. 上传图片到腾讯云
	_, err = client.Object.Put(context.Background(), savePath, src, nil)
	if err != nil {
		return "", fmt.Errorf("上传至腾讯云 COS 失败: %v", err)
	}

	// 4. 拼接最终返回给前端的 CDN 自定义域名地址
	domain := strings.TrimRight(d.cfg.Domain, "/")
	savePath = strings.TrimLeft(savePath, "/")

	return fmt.Sprintf("%s/%s", domain, savePath), nil
}
