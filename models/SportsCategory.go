package models

import (
	"time"
)

// SportsCategory 体育赛事分类
type SportsCategory struct {
	Id      int64     `json:"id"`
	Name    string    `json:"name" xorm:"varchar(60) notnull comment('分类名称 如 足球/篮球')"`
	Icon    string    `json:"icon" xorm:"varchar(60) comment('分类图标 emoji 或 URL')"`
	Rules   string    `json:"rules" xorm:"text comment('比赛规则 富文本')"`
	Remark  string    `json:"remark" xorm:"varchar(255) comment('备注说明')"`
	Weigh   int       `json:"weigh" xorm:"not null default 0 comment('排序权重')"`
	Status  string    `json:"status" xorm:"varchar(40) notnull default 'normal' comment('状态 normal/hidden')"`
	Created time.Time `json:"createtime" xorm:"created int"`
	Updated time.Time `json:"updatetime" xorm:"updated int"`
}

func (a *SportsCategory) TableName() string {
	return "sports_category"
}

// 分页列表
func GetSportsCategoryList(limit int, page int, search string, order string) []*SportsCategory {
	offset := page - 1
	if offset < 0 {
		offset = 0
	}
	byorder := "id ASC"
	if order == "-id" {
		byorder = "id DESC"
	}
	listdata := []*SportsCategory{}
	Dorm.Table("sports_category").
		Where("name like ?", "%"+search+"%").
		OrderBy(byorder).
		Limit(limit, limit*offset).
		Find(&listdata)
	return listdata
}

func GetSportsCategoryTotal(search string) int64 {
	a := new(SportsCategory)
	total, err := Dorm.Where("name like ?", "%"+search+"%").Count(a)
	if err != nil {
		return 0
	}
	return total
}

// 新增
func AddSportsCategory(a *SportsCategory) error {
	_, err := Dorm.Insert(a)
	return err
}

// 修改
func EditSportsCategory(a *SportsCategory) error {
	_, err := Dorm.ID(a.Id).Update(a)
	return err
}

// 删除
func DelSportsCategory(id int64) int {
	a := new(SportsCategory)
	outnum, _ := Dorm.ID(id).Delete(a)
	return int(outnum)
}
