package services

import (
	"fmt"
	"mime/multipart"
)

// StorageDriver 定义统一的云存储接口
type StorageDriver interface {
	// Upload 上传文件到云存储，返回文件的网络访问 URL
	Upload(file *multipart.FileHeader, savePath string) (string, error)
}

// NewStorageDriver 实例化云存储驱动的工厂函数
// provider: 存储服务商名称，如 "cos"、"oss"、"kodo"
// configJson: 对应服务商的 JSON 配置字符串
func NewStorageDriver(provider string, configJson string) (StorageDriver, error) {
	switch provider {
	case "cos", "tencent": // 腾讯云 COS
		return NewCOSDriver(configJson)

	case "oss", "aliyun": // 阿里云 OSS (按需扩展)
		return NewOSSDriver(configJson)

	case "kodo", "qiniu": // 七牛云 Kodo (按需扩展)
		return NewKodoDriver(configJson)

	default:
		return nil, fmt.Errorf("暂不支持的云存储服务商: %s", provider)
	}
}

// ------------------------------------------------------------------
// 以下为存根函数（Stub），防止在未实现 OSS 或 Kodo 时编译报错。
// 如果你后续需要支持阿里云或七牛云，请把相应的实现写在独立的脚本中。
// ------------------------------------------------------------------

func NewOSSDriver(configJson string) (StorageDriver, error) {
	return nil, fmt.Errorf("阿里云 OSS 驱动尚未实现")
}

func NewKodoDriver(configJson string) (StorageDriver, error) {
	return nil, fmt.Errorf("七牛云 Kodo 驱动尚未实现")
}
