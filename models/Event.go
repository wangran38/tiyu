package models

import (
	"time"
	"tiyu/global"
)

// Event 赛事活动
type Event struct {
	Id               int64     `json:"id"`
	Title            string    `json:"title" xorm:"varchar(120) notnull comment('活动主题')"`
	Covers           string    `json:"covers" xorm:"text comment('封面图片 JSON 数组，最多6张')"`
	Posters          string    `json:"posters" xorm:"text comment('长页海报 JSON 数组，最多20张')"`
	CategoryId       int64     `json:"category_id" xorm:"bigint default 0 comment('体育分类ID')"`
	RegistrationFrom time.Time `json:"registration_from" xorm:"datetime comment('报名开始时间')"`
	RegistrationTo   time.Time `json:"registration_to" xorm:"datetime comment('报名结束时间')"`
	StartTime        time.Time `json:"start_time" xorm:"datetime comment('开始时间')"`
	EndTime          time.Time `json:"end_time" xorm:"datetime comment('结束时间')"`
	Fee              string    `json:"fee" xorm:"varchar(50) comment('活动费用，文本格式，例：免费/￥50')"`
	Venue            string    `json:"venue" xorm:"varchar(255) comment('活动地点')"`
	Address          string    `json:"address" xorm:"varchar(255) comment('详细地址')"`
	Description      string    `json:"description" xorm:"text comment('赛事介绍')"`
	SignupSettings   string    `json:"signup_settings" xorm:"text comment('报名项设置 JSON')"`
	DisplaySettings  string    `json:"display_settings" xorm:"text comment('显示设置 JSON')"`
	Tags             string    `json:"tags" xorm:"varchar(255) comment('活动标签，逗号分隔')"`
	Documents        string    `json:"documents" xorm:"text comment('资质/许可 JSON 或附件列表')"`
	RequireReview    bool      `json:"require_review" xorm:"bool default false comment('是否需要审核')"`
	Organizer        string    `json:"organizer" xorm:"varchar(255) comment('主办单位名称')"`
	ContactPerson    string    `json:"contact_person" xorm:"varchar(100) comment('联系人')"`
	ContactPhone     string    `json:"contact_phone" xorm:"varchar(50) comment('联系电话')"`
	Status           string    `json:"status" xorm:"varchar(40) notnull default 'normal' comment('状态 normal/hidden/ended')"`
	Weigh            int       `json:"weigh" xorm:"not null default 0 comment('排序权重')"`
	Created          time.Time `json:"createtime" xorm:"created int"`
	Updated          time.Time `json:"updatetime" xorm:"updated int"`
}

func (e *Event) TableName() string {
	return "events"
}

// 分页列表
func GetEventList(limit int, page int, search string, categoryId int64, registrationFrom *time.Time, registrationTo *time.Time, startTime *time.Time, endTime *time.Time, order string) []*Event {
	offset := page - 1
	if offset < 0 {
		offset = 0
	}
	byorder := "weigh DESC, created DESC"
	switch order {
	case "weigh":
		byorder = "weigh ASC, created DESC"
	case "-weigh":
		byorder = "weigh DESC, created DESC"
	case "createtime":
		byorder = "created ASC"
	case "-createtime":
		byorder = "created DESC"
	case "id":
		byorder = "id ASC"
	case "-id":
		byorder = "id DESC"
	}
	query := global.Dorm.Table("events")
	if search != "" {
		query = query.Where("title like ?", "%"+search+"%")
	}
	if categoryId > 0 {
		query = query.And("category_id = ?", categoryId)
	}
	if registrationFrom != nil {
		query = query.And("registration_from >= ?", *registrationFrom)
	}
	if registrationTo != nil {
		query = query.And("registration_to <= ?", *registrationTo)
	}
	if startTime != nil {
		query = query.And("start_time >= ?", *startTime)
	}
	if endTime != nil {
		query = query.And("end_time <= ?", *endTime)
	}
	listdata := []*Event{}
	query.OrderBy(byorder).Limit(limit, limit*offset).Find(&listdata)
	return listdata
}

func GetEventTotal(search string, categoryId int64, registrationFrom *time.Time, registrationTo *time.Time, startTime *time.Time, endTime *time.Time) int64 {
	a := new(Event)
	query := global.Dorm.NewSession().Where("1 = 1")
	if search != "" {
		query = query.Where("title like ?", "%"+search+"%")
	}
	if categoryId > 0 {
		query = query.And("category_id = ?", categoryId)
	}
	if registrationFrom != nil {
		query = query.And("registration_from >= ?", *registrationFrom)
	}
	if registrationTo != nil {
		query = query.And("registration_to <= ?", *registrationTo)
	}
	if startTime != nil {
		query = query.And("start_time >= ?", *startTime)
	}
	if endTime != nil {
		query = query.And("end_time <= ?", *endTime)
	}
	total, err := query.Count(a)
	if err != nil {
		return 0
	}
	return total
}

// 新增
func AddEvent(a *Event) error {
	_, err := global.Dorm.Insert(a)
	return err
}

// 修改
func EditEvent(a *Event) error {
	_, err := global.Dorm.ID(a.Id).Update(a)
	return err
}

// 删除
func DelEvent(id int64) int {
	a := new(Event)
	outnum, _ := global.Dorm.ID(id).Delete(a)
	return int(outnum)
}
