package utils

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

//生成jwt身份标识
func CreateJsonWebToken(userName string) (token string) {
	keyInfo := "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&*()"
	//将部分用户信息保存到map转化为json
	info := map[string]interface{}{}
	info["userName"] = userName
	dataByte, _ := json.Marshal(info)
	var dataStr = string(dataByte)
	//获取当前时间戳
	t := time.Now().Unix()
	//设置过期时间 30分钟
	exTime := t + 60000*30
	//使用Claim保存json
	//data := jwt.StandardClaims{Subject:dataStr,ExpiresAt:int64(time.Now().Add(time.Hour * 72).Unix())}
	data := jwt.StandardClaims{Subject: dataStr, ExpiresAt: exTime}
	tokenInfo := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	//生成token字符串
	token, _ = tokenInfo.SignedString([]byte(keyInfo))
	tMap := make(map[string]interface{})
	//把token存进map
	tMap["token"] = token
	//过期时间
	tMap["expireTime"] = exTime
	json, _ := json.Marshal(tMap)
	//设置redis有效期
	timer := 60 * 30
	InsertRedisKeyExpire("loginAdmin_"+userName, string(json), timer)
	return
}

//效验token是否过期
func CheckTokenExpired(token string) error {
	keyInfo := "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&*()"
	tokenInfo, _ := jwt.Parse(token, func(token *jwt.Token) (i interface{}, e error) {
		return keyInfo, nil
	})
	//效验jwt令牌是否失效
	err := tokenInfo.Claims.Valid()
	if err != nil {
		print("jwt失效 ", err.Error())
	}
	return err
}

//解析jwt获取到当前登录用户的唯一标识  返回{"userName":"admin"} json字符串
func GetLoginUserName(token string) string {
	keyInfo := "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&*()"
	tokenInfo, _ := jwt.Parse(token, func(token *jwt.Token) (i interface{}, e error) {
		return keyInfo, nil
	})
	jwtMap := tokenInfo.Claims.(jwt.MapClaims)
	fmt.Print(jwtMap["sub"].(string))
	return jwtMap["sub"].(string)
}

//根据用户唯一标识从redis取出存在里面的json字符串
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

//验证令牌有效期，相差不足20分钟，自动刷新缓存
func VerifyToken(tMap map[string]interface{}) {
	//得到20分钟时间戳
	MILLIS_MINUTE_TEN := 1000 * 60 * 20
	//获取过期时间
	var expireTime = int64(tMap["expireTime"].(float64))
	var currentTime = time.Now().Unix()
	//判断
	if expireTime-currentTime <= int64(MILLIS_MINUTE_TEN) {
		//重新生成token
		CreateJsonWebToken(tMap["userName"].(string))
	}
}
