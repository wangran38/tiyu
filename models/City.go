package models

//城市后端模型
import (
	"errors"
	"tiyu/global"

	_ "github.com/go-sql-driver/mysql"
)

type City struct {
	Id        int64
	Pid       int
	Shortname string `xorm:"varchar(200)" json:"shortname"`
	Name      string `xorm:"varchar(200)" json:"name"`
	Mergename string `xorm:"varchar(200)" json:"mergename"`
	Level     int    `json:"status" xorm:"not null default 1 comment('层级 0 1 2 省市区县') TINYINT"`
	Pinyin    string `xorm:"varchar(200)" json:"pingyin"`
	Code      string `xorm:"varchar(200)" json:"code"`
	Zip       string `xorm:"varchar(200)" json:"zip"`
	First     string `xorm:"varchar(200)" json:"first"`
	Lng       string `xorm:"varchar(200)" json:"lng"`
	Lat       string `xorm:"varchar(200)" json:"lat"`
}

func (a *City) TableName() string {
	return "area"
}

// 根据用户名密码查询用户
func SelectBycityid(Id int) (*City, error) {
	a := new(City)
	has, err := global.Dorm.Where("id = ?", Id).Get(a)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("暂无此条数据！")
	}
	return a, nil

}

// 分页列表：limit 每页条数，page 页码（从1开始），search 按名称模糊匹配，order 排序
func GetCityList(limit int, page int, search string, order string) []*City {
	offset := page - 1
	if offset < 0 {
		offset = 0
	}
	byorder := "id ASC"
	if order == "-id" {
		byorder = "id DESC"
	}
	listdata := []*City{}
	global.Dorm.Table("area").
		Where("name like ?", "%"+search+"%").
		OrderBy(byorder).
		Limit(limit, limit*offset).
		Find(&listdata)
	return listdata
}

func GetCityTotal(search string) int64 {
	a := new(City)
	total, err := global.Dorm.Cols("id", "name").Where("name like ?", "%"+search+"%").Count(a)
	if err != nil {
		return 0
	}
	return total
}

// 新增地区
func AddCity(a *City) error {
	_, err := global.Dorm.Insert(a)
	return err
}

// 修改地区
func EditCity(a *City) error {
	_, err := global.Dorm.ID(a.Id).Update(a)
	return err
}

// 删除地区
func DelCity(id int64) int {
	a := new(City)
	outnum, _ := global.Dorm.ID(id).Delete(a)
	return int(outnum)
}

// add
// GetCitiesByPid 根据 Pid 查询地区

// GetAllAreaList 获取 area 表中的所有记录（用于构建树）

// GetAllCitiesByLevel 根据层级获取城市 (1 为地级市)
func GetAllCitiesByLevel(level int) ([]*City, error) {
	cities := make([]*City, 0)
	err := global.Dorm.Table("area").
		Where("level = ?", level).
		OrderBy("first ASC, pinyin ASC").
		Find(&cities)
	return cities, err
}

// GetHotCities 获取热门城市列表
func GetHotCities() ([]*City, error) {
	cities := make([]*City, 0)
	hotNames := []string{"北京", "上海", "广州", "深圳", "成都", "杭州", "武汉"}
	err := global.Dorm.Table("area").
		In("name", hotNames).
		Find(&cities)
	return cities, err
}

// GetAllAreaList 获取全量地区列表（构建三级树）
func GetAllAreaList() ([]*City, error) {
	cities := make([]*City, 0)
	err := global.Dorm.Table("area").OrderBy("id ASC").Find(&cities)
	return cities, err
}

// GetCitiesByPid 根据 Pid 查询下级地区
func GetCitiesByPid(pid int64) ([]*City, error) {
	cities := make([]*City, 0)
	err := global.Dorm.Table("area").Where("pid = ?", pid).OrderBy("id ASC").Find(&cities)
	return cities, err
}
