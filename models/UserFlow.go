package models

import "time"

// UserFlow 会员流水（积分/等级变化记录）
type UserFlow struct {
	Id       int64     `json:"id"`
	UserId   int64     `json:"user_id" xorm:"notnull default 0 comment('会员 ID') index"`
	Username string    `json:"username" xorm:"varchar(60) comment('会员用户名冗余')"`
	Type     string    `json:"type" xorm:"varchar(30) notnull comment('类型 point/level/status')"`
	Change   int       `json:"change" xorm:"notnull default 0 comment('变化量 增为正/减为负')"`
	Before   int       `json:"before" xorm:"notnull default 0 comment('变化前值')"`
	After    int       `json:"after" xorm:"notnull default 0 comment('变化后值')"`
	Reason   string    `json:"reason" xorm:"varchar(255) comment('变化原因')"`
	Operator string    `json:"operator" xorm:"varchar(60) comment('操作人 后台管理员账号')"`
	Remark   string    `json:"remark" xorm:"varchar(255) comment('备注')"`
	Created  time.Time `json:"created_at" xorm:"created int"`
}

func (a *UserFlow) TableName() string {
	return "user_flow"
}

// 分页查询（支持 userId/type 时间范围搜索）
func GetUserFlowList(limit int, page int, search *UserFlow, order string) []*UserFlow {
	offset := page - 1
	if offset < 0 {
		offset = 0
	}
	byorder := "id ASC"
	if order == "-id" {
		byorder = "id DESC"
	}
	listdata := []*UserFlow{}
	session := Dorm.Table("user_flow")
	if search.UserId > 0 {
		session = session.And("user_id = ?", search.UserId)
	}
	if search.Username != "" {
		session = session.And("username LIKE ?", "%"+search.Username+"%")
	}
	if search.Type != "" {
		session = session.And("type = ?", search.Type)
	}
	session.OrderBy(byorder).Limit(limit, limit*offset).Find(&listdata)
	return listdata
}

func GetUserFlowTotal(search *UserFlow) int64 {
	session := Dorm.Table("user_flow")
	if search.UserId > 0 {
		session = session.And("user_id = ?", search.UserId)
	}
	if search.Username != "" {
		session = session.And("username LIKE ?", "%"+search.Username+"%")
	}
	if search.Type != "" {
		session = session.And("type = ?", search.Type)
	}
	total, err := session.Count(new(UserFlow))
	if err != nil {
		return 0
	}
	return total
}

// 写入一条流水
func AddUserFlow(f *UserFlow) error {
	_, err := Dorm.Insert(f)
	return err
}
