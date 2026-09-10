package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// jwt 签名密钥
const jwtSecret = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&*()"

// ==========================================
// 1. 原有 Admin 相关的 JWT 逻辑（后台管理系统）
// ==========================================

// 生成jwt身份标识 (Admin)
func CreateJsonWebToken(userName string) (token string) {
	keyInfo := jwtSecret
	info := map[string]interface{}{}
	info["userName"] = userName
	dataByte, _ := json.Marshal(info)
	var dataStr = string(dataByte)
	t := time.Now().Unix()
	exTime := t + 60000*30
	data := jwt.StandardClaims{Subject: dataStr, ExpiresAt: exTime}
	tokenInfo := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	token, _ = tokenInfo.SignedString([]byte(keyInfo))
	tMap := make(map[string]interface{})
	tMap["token"] = token
	tMap["expireTime"] = exTime
	jsonStr, _ := json.Marshal(tMap)
	timer := 60 * 30
	InsertRedisKeyExpire("loginAdmin_"+userName, string(jsonStr), timer)
	return
}

// 效验token是否过期
func CheckTokenExpired(token string) error {
	keyInfo := jwtSecret
	tokenInfo, _ := jwt.Parse(token, func(token *jwt.Token) (i interface{}, e error) {
		return keyInfo, nil
	})
	err := tokenInfo.Claims.Valid()
	if err != nil {
		fmt.Println("jwt失效 ", err.Error())
	}
	return err
}

// 解析jwt获取到当前登录用户的唯一标识 返回{"userName":"admin"} json字符串
func GetLoginUserName(token string) string {
	keyInfo := jwtSecret
	tokenInfo, _ := jwt.Parse(token, func(token *jwt.Token) (i interface{}, e error) {
		return keyInfo, nil
	})
	jwtMap := tokenInfo.Claims.(jwt.MapClaims)
	fmt.Print(jwtMap["sub"].(string))
	return jwtMap["sub"].(string)
}

// 根据用户唯一标识从redis取出存在里面的json字符串
func GetRedisMapByUserName(userName string) (tokenMap map[string]interface{}) {
	tokenMap = make(map[string]interface{})
	str := "loginAdmin_" + userName

	value := GetValueByKey(str)
	err := json.Unmarshal([]byte(value), &tokenMap)
	if err != nil {
		fmt.Println("JSON To Map error", err)
	}
	return
}

// 验证令牌有效期，相差不足20分钟，自动刷新缓存
func VerifyToken(tMap map[string]interface{}) {
	MILLIS_MINUTE_TEN := 1000 * 60 * 20
	var expireTime = int64(tMap["expireTime"].(float64))
	var currentTime = time.Now().Unix()
	if expireTime-currentTime <= int64(MILLIS_MINUTE_TEN) {
		CreateJsonWebToken(tMap["userName"].(string))
	}
}

// ParseToken 解析并校验 JWT（签名 + 过期时间），成功返回签发时写入的管理员 ID。
func ParseToken(token string) (string, error) {
	tk, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || tk == nil {
		return "", fmt.Errorf("token 无效")
	}
	claims, ok := tk.Claims.(jwt.MapClaims)
	if !ok || !tk.Valid {
		return "", fmt.Errorf("token 无效或已过期")
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return "", fmt.Errorf("token 缺少用户信息")
	}
	info := map[string]interface{}{}
	if err := json.Unmarshal([]byte(sub), &info); err != nil {
		return "", fmt.Errorf("token 用户信息解析失败")
	}
	adminId, _ := info["userName"].(string)
	if adminId == "" {
		return "", fmt.Errorf("token 缺少用户标识")
	}
	return adminId, nil
}

// JWTAuth 统一鉴权中间件 (Admin 后台使用)
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.Request.Header.Get("Authorization"))
		if strings.HasPrefix(token, "Bearer ") {
			token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer"))
		}
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未登录，请先登录",
			})
			return
		}
		adminId, err := ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "登录凭证无效，请重新登录",
			})
			return
		}
		if CheckRedisExits("loginAdmin_"+adminId) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "登录已过期，请重新登录",
			})
			return
		}
		c.Set("adminId", adminId)
		c.Next()
	}
}

// ==========================================
// 2. 新增：前台/SSO 用户凭证方法 (通用扩展)
// ==========================================

// UserClaims 定义前台用户 Token 载荷
type UserClaims struct {
	UserID int64  `json:"user_id"`
	Mobile string `json:"mobile"`
	jwt.StandardClaims
}

// GenerateToken 生成前台用户 JWT Token (兼容 SSOLogin 调用)
func GenerateToken(userID int64, mobile string) (string, error) {
	exTime := time.Now().Add(24 * time.Hour).Unix() // 前台用户 Token 默认 24 小时有效

	claims := UserClaims{
		UserID: userID,
		Mobile: mobile,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: exTime,
			IssuedAt:  time.Now().Unix(),
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := tokenClaims.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	// 写入 Redis 缓存登录态 (保留 24 小时)
	redisKey := "loginUser_" + strconv.FormatInt(userID, 10)
	tMap := map[string]interface{}{
		"token":      tokenStr,
		"user_id":    userID,
		"mobile":     mobile,
		"expireTime": exTime,
	}
	jsonStr, _ := json.Marshal(tMap)
	InsertRedisKeyExpire(redisKey, string(jsonStr), 86400)

	return tokenStr, nil
}

// ParseUserToken 解析前台用户 Token
func ParseUserToken(tokenStr string) (*UserClaims, error) {
	tk, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || tk == nil {
		return nil, fmt.Errorf("token 无效")
	}

	if claims, ok := tk.Claims.(*UserClaims); ok && tk.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("token 验证失败")
}

// UserJWTAuth 前台用户统一鉴权中间件 (前台/App/小程序使用)
// UserJWTAuth 前台用户统一鉴权中间件 (前台/App/小程序/跨系统使用)
func UserJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 优先从 Header 提取 pgtoken，如果不存在则从 Authorization 中提取
		token := strings.TrimSpace(c.Request.Header.Get("pgtoken"))
		if token == "" {
			token = strings.TrimSpace(c.Request.Header.Get("Authorization"))
			if strings.HasPrefix(token, "Bearer ") {
				token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer"))
			}
		}

		// 2. 如果都没有，返回 401
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "请先登录 (未检测到 pgtoken)",
			})
			return
		}

		// 3. 解析 Token
		claims, err := ParseUserToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "登录已失效，请重新登录",
			})
			return
		}

		// 4. 检查 Redis 中的登录态
		redisKey := "loginUser_" + strconv.FormatInt(claims.UserID, 10)
		if CheckRedisExits(redisKey) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "登录信息已过期，请重新登录",
			})
			return
		}

		// 5. 注入上下文变量 (方便 Controller 使用)
		c.Set("userId", claims.UserID)
		c.Set("userMobile", claims.Mobile)
		c.Next()
	}
}
