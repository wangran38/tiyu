package utils

import (
	"fmt"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

var redisPool *redis.Pool

func init() {
	redisPool = InitRedis()
}

func InitRedis() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		MaxActive:   5,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "127.0.0.1:6379")
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

//新增key
func InsertRedis(key string, value string) error {
	conn := redisPool.Get()
	defer conn.Close()
	_, err := conn.Do("SET", key, value)
	if err != nil {
		fmt.Println("redis set error", err)
	}
	return err
}

//根据key取值
func GetValueByKey(key string) string {
	conn := redisPool.Get()
	defer conn.Close()
	data, err := redis.String(conn.Do("GET", key))
	if err != nil {
		log.Println(err)
	}
	return data
}

//新增key并设置过期时间
func InsertRedisKeyExpire(key string, value string, time int) error {
	conn := redisPool.Get()
	defer conn.Close()
	_, err := conn.Do("SETEX", key, time, value)
	if err != nil {
		fmt.Println("redis set error", err)
	}
	return err
}

//检测是否存在该key 返回0不存在 1存在
func CheckRedisExits(key string) int64 {
	conn := redisPool.Get()
	defer conn.Close()
	v, _ := conn.Do("EXISTS", key)
	return v.(int64)
}

//删除key
func DelRedisKey(key string) error {
	conn := redisPool.Get()
	defer conn.Close()
	_, err := conn.Do("DEL", key)
	return err
}
