package models

import (
	_ "database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	// _ "tiyu/db"
	"github.com/cockroachdb/pebble"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
)

// 读数据
var Dorm *xorm.Engine

// //写数据
// var DB_Write *xDorm.Engine

// var engine *xDorm.Engine
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
	//Dorm, err = xorm.NewEngine("mysql", "tiyu:aksaMDdERmZeMWCD@tcp(localhost:3306)/tiyu?charset=utf8mb4")
	Dorm, err = xorm.NewEngine("mysql", "root:root@tcp(localhost:3306)/tiyu?charset=utf8mb4")
	// engine, err := xDorm.NewEngine("mysql", "2343432:122222@/(http://127.0.0.1:3306)/tiyu?charset=utf8")
	// db, err = xDorm.NewEngine("mysql", "username:password@tcp(host:3306)/dbname?charset=utf8")
	if err != nil {
		fmt.Println(err)
		fmt.Println(err.Error())
		return
	} else {
		Dorm.Ping()
		//是否显示sql语句
		Dorm.ShowSQL(true)
		if err = Dorm.Sync2(
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
		); err != nil {
			fmt.Println(err)
		} else {
			// Dorm.ShowWarn(true)
			// engine.ShowWarn(true)
			fmt.Print("自动生成表成功！")
		}

	}
	Dorm.SetMaxOpenConns(50)
	// 设置连接池的空闲数大小
	Dorm.SetMaxIdleConns(5)
	// 设置空闲连接最大时长
	// engine.SetConnMaxIdleTime(600) xDorm没有这个功能
	// 设置连接最大存活时长，必须小于mysql的wait_timeout
	// mysql wait_timeout单位是s，time.Duration 默认是纳秒，后面*1000000000，转化成s
	// 设置为4h
	Dorm.SetConnMaxLifetime(59 * time.Second)
	// fmt.Println(err)
}

// GetDorm 获取数据库引擎（供脚本使用）
func GetDorm() *xorm.Engine {
	return Dorm
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
