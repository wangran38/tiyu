package utils

import (
	"encoding/json"
	"fmt"
	"strings"
	"tiyu/models"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

// TencentConfigDetail 腾讯云 Config JSON 解析结构
type TencentConfigDetail struct {
	SecretId    string `json:"secret_id"`
	SecretKey   string `json:"secret_key"`
	SmsSdkAppId string `json:"sms_sdk_app_id"`
	SignName    string `json:"sign_name"`
	TemplateId  string `json:"template_id"`
}

// SendDynamicSmsCode 自动获取数据库默认配置并发送短信
func SendDynamicSmsCode(mobile, scene, code string) error {
	// 1. 从数据库获取启用且默认的短信配置
	smsConfig, err := models.GetDefaultSmsConfig()
	if err != nil {
		return fmt.Errorf("读取短信配置异常: %v", err)
	}
	if smsConfig == nil {
		return fmt.Errorf("系统未开启或未配置默认短信服务商")
	}

	// 2. 根据服务商类型分发
	switch smsConfig.Provider {
	case "tencent":
		var detail TencentConfigDetail
		if err := json.Unmarshal([]byte(smsConfig.Config), &detail); err != nil {
			return fmt.Errorf("解析腾讯云短信配置失败: %v", err)
		}
		return sendTencentSms(detail, mobile, code)

	case "aliyun":
		// 如果后续扩展阿里云可以在此补充
		return fmt.Errorf("暂不支持阿里云短信")

	default:
		return fmt.Errorf("不支持的短信服务商: %s", smsConfig.Provider)
	}
}

// sendTencentSms 底层调用腾讯云 SDK
// sendTencentSms 底层调用腾讯云 SDK
func sendTencentSms(cfg TencentConfigDetail, mobile, code string) error {
	// 自动清理首尾空格
	secretId := strings.TrimSpace(cfg.SecretId)
	secretKey := strings.TrimSpace(cfg.SecretKey)

	credential := common.NewCredential(secretId, secretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqMethod = "POST"
	cpf.HttpProfile.Endpoint = "sms.tencentcloudapi.com"

	client, err := sms.NewClient(credential, "ap-guangzhou", cpf)
	if err != nil {
		return fmt.Errorf("客户端初始化失败: %v", err)
	}

	req := sms.NewSendSmsRequest()
	req.SmsSdkAppId = common.StringPtr(cfg.SmsSdkAppId)
	req.SignName = common.StringPtr(cfg.SignName)
	req.TemplateId = common.StringPtr(cfg.TemplateId)
	req.PhoneNumberSet = common.StringPtrs([]string{"+86" + mobile})
	// 假设模板参数: {1}为验证码，{2}为有效时间（5分钟）
	req.TemplateParamSet = common.StringPtrs([]string{code, "5"})

	resp, err := client.SendSms(req)
	if err != nil {
		return fmt.Errorf("SDK 发送失败: %v", err)
	}

	if len(resp.Response.SendStatusSet) > 0 {
		status := resp.Response.SendStatusSet[0]
		if *status.Code != "Ok" {
			return fmt.Errorf("腾讯云返回错误 [%s]: %s", *status.Code, *status.Message)
		}
	}
	return nil
}
