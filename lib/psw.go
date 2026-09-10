package lib

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"time"
)

func Md5(buf []byte) string {
	hash := md5.New()
	hash.Write(buf)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func Password(len int, pwdO string) (pwd string, salt string) {
	salt = GetRandomString(len)
	defaultPwd := "988cj" + "988cj.com"
	if pwdO != "" {
		defaultPwd = pwdO + "988cj.com"
	}
	pwd = Md5([]byte(defaultPwd + salt))
	return pwd, salt
}

//生成文件用户的密钥
// func Autotoken(phonenum string) string {
// 	sh := "sh"
// 	// ctime := strconv.Formatlnt(int64, creatime)
// 	str := Md5([]byte(phonenum + sh))
// 	return string(str)
// }

// 生成32位MD5
// func MD5(text string) string{
//    ctx := md5.New()
//    ctx.Write([]byte(text))
//    return hex.EncodeToString(ctx.Sum(nil))
// }

//生成随机字符串
func GetRandomString(lens int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < lens; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}
