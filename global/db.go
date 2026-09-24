// global/db.go
package global

import (
	"github.com/go-xorm/xorm"
)

// Dorm 全局 XORM 数据库连接句柄（抽离到最底层，供所有业务包共享）
var Dorm *xorm.Engine
