package models

import (
	_ "database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
	"tiyu/global"
	"tiyu/models/shop"

	// "tiyu/models/shop"

	// _ "tiyu/db"
	"github.com/cockroachdb/pebble"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/joho/godotenv"
)

// 读数据
// var Dorm *xorm.Engine

var err error

// PebbleDataDir 是票根 KV 数据库目录，可通过 TICKET_PEBBLE_PATH 覆盖。
var PebbleDataDir = func() string {
	if path := os.Getenv("TICKET_PEBBLE_PATH"); path != "" {
		return path
	}
	return filepath.Join("pbdata")
}()

var (
	PebbleDB      *pebble.DB
	pebbleOnce    sync.Once
	pebbleInitErr error
)

func init() {
	// 尝试加载根目录下的 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("Notice: No .env file found, relying on system environment variables.")
	}

	// 从环境变量读取 MySQL 配置，如果未找到则使用默认值
	dbUser := getEnv("DB_USER", "root")
	dbPass := getEnv("DB_PASS", "root")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "tiyu")

	// 拼接 DSN 连接串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4", dbUser, dbPass, dbHost, dbPort, dbName)

	global.Dorm, err = xorm.NewEngine("mysql", dsn)
	if err != nil {
		fmt.Println("Database connection error:", err)
		return
	}

	if err = global.Dorm.Ping(); err != nil {
		fmt.Println("Database ping error:", err)
		return
	}

	// 是否显示sql语句
	global.Dorm.ShowSQL(true)
	if err = global.Dorm.Sync2(
		new(Admin),
		new(User),
		new(Authgroup),
		new(Authrule),
		new(Authaccess),
		new(City),
		new(SportsCategory),
		new(UserFlow),
		new(Event),
		new(TicketTemplate),
		new(TicketClaim),
		new(MemberTicket),
		new(PublicTicketRecognition),
		new(Order),
		new(ShopCategory),
		new(Shop),
		new(Coupon),
		new(StorageConfig),
		new(SmsConfig),
		// ---- 本地生活/商家团购（shop包） ----
		new(shop.GoodsProduct),
		new(shop.GoodsSku),
		new(shop.GoodsItem),
		new(shop.GoodsRule),
		new(shop.GoodsTicketDiscount),
		new(shop.GoodsDailyCalendar),
		new(OrderItem),
	); err != nil {
		fmt.Println("Sync table error:", err)
	} else {
		fmt.Print("自动生成表成功！")
	}

	global.Dorm.SetMaxOpenConns(50)
	// 设置连接池的空闲数大小
	global.Dorm.SetMaxIdleConns(5)
	// 设置连接最大存活时长，必须小于mysql的wait_timeout
	global.Dorm.SetConnMaxLifetime(59 * time.Second)
}

// 辅助函数：获取环境变量，带默认值
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// GetDorm 获取数据库引擎（供脚本使用）
func GetDorm() *xorm.Engine {
	return global.Dorm
}

// GetPebbleDB 获取票根 Pebble KV 连接，首次调用时自动打开。
func GetPebbleDB() (*pebble.DB, error) {
	pebbleOnce.Do(func() {
		if err := os.MkdirAll(PebbleDataDir, 0755); err != nil {
			pebbleInitErr = fmt.Errorf("create pebble data directory: %w", err)
			return
		}
		PebbleDB, pebbleInitErr = pebble.Open(PebbleDataDir, &pebble.Options{})
	})
	return PebbleDB, pebbleInitErr
}

// ClosePebbleDB 关闭票根 Pebble KV 连接，服务退出时调用。
func ClosePebbleDB() error {
	if PebbleDB == nil {
		return nil
	}
	return PebbleDB.Close()
}
