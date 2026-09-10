package utils

import (
	"encoding/json"
	"fmt"
	"time"
)

// SmsCodeInfo 验证码缓存结构体
type SmsCodeInfo struct {
	Mobile     string `json:"mobile"`      // 手机号
	Code       string `json:"code"`        // 验证码
	Scene      string `json:"scene"`       // 场景: register, login, reset_pwd
	CreatedAt  int64  `json:"created_at"`  // 创建时间戳 (秒)
	RetryCount int    `json:"retry_count"` // 校验失败重试次数 (防暴力破解)
}

// ==================== 1. 验证码读写与防暴力破解逻辑 ====================

// SetSmsCodeToRedis 将验证码结构体转为 JSON 存入 Redis (默认 ttlSeconds 为 180 秒 = 3分钟)
func SetSmsCodeToRedis(mobile, code, scene string, ttlSeconds int) error {
	info := SmsCodeInfo{
		Mobile:     mobile,
		Code:       code,
		Scene:      scene,
		CreatedAt:  time.Now().Unix(),
		RetryCount: 0,
	}

	// 1. 序列化为 JSON
	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("json marshal error: %v", err)
	}

	// 2. 存入 Redis (调用你的 InsertRedisKeyExpire)
	redisKey := fmt.Sprintf("sms:code:%s:%s", scene, mobile)
	InsertRedisKeyExpire(redisKey, string(data), ttlSeconds)

	return nil
}

// GetSmsCodeFromRedis 从 Redis 读取 JSON 并反序列化为结构体
func GetSmsCodeFromRedis(mobile, scene string) (*SmsCodeInfo, error) {
	redisKey := fmt.Sprintf("sms:code:%s:%s", scene, mobile)

	// 调用你的 GetValueByKey
	valStr := GetValueByKey(redisKey)
	if valStr == "" {
		return nil, nil // 验证码不存在或已过期
	}

	// 反序列化
	var info SmsCodeInfo
	err := json.Unmarshal([]byte(valStr), &info)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal error: %v", err)
	}

	return &info, nil
}

// VerifyAndHandleSmsCode 校验验证码，并自动处理重试次数上限 (输错 3 次自动从 Redis 清除)
// 返回值：(是否校验通过, 错误原因说明/剩余次数)
func VerifyAndHandleSmsCode(mobile, scene, inputCode string) (bool, string) {
	redisKey := fmt.Sprintf("sms:code:%s:%s", scene, mobile)

	// 1. 获取验证码
	info, err := GetSmsCodeFromRedis(mobile, scene)
	if err != nil || info == nil {
		return false, "验证码已过期或不存在，请重新发送"
	}

	// 2. 匹配验证码
	if info.Code == inputCode {
		// 验证成功，调用你的 DelRedisKey 清除验证码
		DelRedisKey(redisKey)
		return true, ""
	}

	// 3. 验证码匹配失败，错误次数 +1
	info.RetryCount++

	// 超过最大失败次数 (3次)，物理删除验证码
	if info.RetryCount >= 3 {
		DelRedisKey(redisKey)
		return false, "验证码错误次数过多，已自动失效，请重新发送"
	}

	// 没达到上限，更新 Redis 里的 RetryCount (保留剩余 TTL)
	data, _ := json.Marshal(info)
	remainingTTL := int((info.CreatedAt + 180) - time.Now().Unix())
	if remainingTTL > 0 {
		InsertRedisKeyExpire(redisKey, string(data), remainingTTL)
	} else {
		DelRedisKey(redisKey)
	}

	return false, fmt.Sprintf("验证码错误，还可尝试 %d 次", 3-info.RetryCount)
}

// DeleteSmsCodeFromRedis 手动删除 Redis 验证码 (如注册/登录成功后)
func DeleteSmsCodeFromRedis(mobile, scene string) {
	redisKey := fmt.Sprintf("sms:code:%s:%s", scene, mobile)
	DelRedisKey(redisKey)
}

// ==================== 2. 防刷与频控逻辑 ====================

// CheckSmsSendLimit 检查 1 分钟 (60 秒) 内是否重复发送
func CheckSmsSendLimit(mobile, scene string) bool {
	limitKey := fmt.Sprintf("sms:limit:%s:%s", scene, mobile)
	// 调用你的 CheckRedisExits
	return CheckRedisExits(limitKey) == 1
}

// SetSmsSendLimit 设置 1 分钟发送冷却标记
func SetSmsSendLimit(mobile, scene string) {
	limitKey := fmt.Sprintf("sms:limit:%s:%s", scene, mobile)
	InsertRedisKeyExpire(limitKey, "1", 60)
}

// CheckAndIncrSmsDailyCount 检查并累加单日最大发送条数 (自然日 23:59:59 自动重置)
// 返回值：(是否允许发送, 当前累积发送次数, 错误信息)
func CheckAndIncrSmsDailyCount(mobile string, maxDailyLimit int) (bool, int, error) {
	todayStr := time.Now().Format("2006-01-02")
	dailyKey := fmt.Sprintf("sms:daily:%s:%s", todayStr, mobile)

	// 1. 读取当前已发送次数
	countStr := GetValueByKey(dailyKey)
	currentCount := 0
	if countStr != "" {
		fmt.Sscanf(countStr, "%d", &currentCount)
	}

	// 2. 如果已达到或超出限制，拒绝发送
	if currentCount >= maxDailyLimit {
		return false, currentCount, nil
	}

	// 3. 计算距离今晚 23:59:59 的剩余秒数
	now := time.Now()
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	ttlSeconds := int(endOfDay.Unix() - now.Unix())
	if ttlSeconds <= 0 {
		ttlSeconds = 1
	}

	// 4. 次数 +1 并设置自然日 TTL
	newCount := currentCount + 1
	InsertRedisKeyExpire(dailyKey, fmt.Sprintf("%d", newCount), ttlSeconds)

	return true, newCount, nil
}
