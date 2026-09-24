package models

import (
	"fmt"
	"strings"
	"time"
	"tiyu/global"
	// "reflect"
)

type Authrule struct {
	Id         int64     `json:"id"`
	Pid        int64     `json:"pid"`
	Type       string    `json:"type"`
	Icon       string    `json:"icon"`
	Pathname   string    `json:"pathname"`
	Title      string    `json:"title"`
	Remark     string    `json:"remark"`
	Ismenu     int       `json:"ismenu" xorm:"not null default 1 comment('是否启用 默认1 菜单 0 文件') TINYINT"`
	Created    time.Time `json:"createtime" xorm:"created int"`
	Updated    time.Time `json:"updatetime" xorm:"updated int"`
	Deletetime int       `json:"deletetime"`
	Weigh      int       `json:"weigh"`
	Status     string    `json:"status" xorm:"varchar(40)"`
	Component  string    `json:"component"`
}
type Treerule struct {
	Id        int64
	Pid       int64
	Type      string `json:"type"`
	Icon      string `json:"icon"`
	Pathname  string `json:"pathname"`
	Component string `json:"component"`
	Title     string `json:"title"`
	Remark    string
	Ismenu    int `json:"ismenu"`
	Weigh     int
	Status    string `json:"status"`
	Children  []*Treerule
}

func (a *Authrule) TableName() string {
	return "auth_rule"
}

// 获取树状数据
func Getruletree() []*Treerule {
	m := new(Authrule)
	//不new一个新的，采用结构体，外部无法访问()getruletreee() []*tt这样子，只能new一个，然后去访问
	return m.Treelist(0)

}

// 全部菜单
func (m *Authrule) Treelist(pid int64) []*Treerule {
	// menus := new(Authrule)
	// 	var a []Authrule
	var menus []Authrule
	global.Dorm.Where("pid = ?", pid).Where("(deletetime = ? OR deletetime IS NULL)", 0).Find(&menus)
	treelist := []*Treerule{}
	for _, v := range menus {
		child := v.Treelist(v.Id)
		node := &Treerule{
			Id:        v.Id,
			Pid:       v.Pid,
			Type:      v.Type,
			Icon:      v.Icon,
			Pathname:  v.Pathname,
			Component: v.Component,
			Title:     v.Title,
			Remark:    v.Remark,
			Ismenu:    v.Ismenu,
			Weigh:     v.Weigh,
			Status:    v.Status,
		}
		node.Children = child
		treelist = append(treelist, node)
	}
	return treelist

}

// 获取用户树状数据
func Getruleadmintree(Rules string) []*Treerule {
	m := new(Authrule)
	//不new一个新的，采用结构体，外部无法访问()getruletreee() []*tt这样子，只能new一个，然后去访问
	return m.Treelistgroup(0, Rules)

}

// 传参得某个权限的菜单集合
func (m *Authrule) Treelistgroup(pid int64, Rules string) []*Treerule {
	fmt.Println(Rules)
	//    a := new(Authrule)
	// // 	var a []Authrule
	var menus []*Authrule
	ids := strings.Split(Rules, ",") //转成数组用global.Dorm in
	// 	// ids:= string.Join(Rules,",")
	// Where("pid = ?", v.Id)
	global.Dorm.Where("pid = ?", pid).Where("status = ?", "normal").In("id", ids).Find(&menus)
	treelist := []*Treerule{}
	for _, v := range menus {
		child := v.Treelistgroup(v.Id, Rules)
		node := &Treerule{
			Id:        v.Id,
			Pid:       v.Pid,
			Type:      v.Type,
			Icon:      v.Icon,
			Pathname:  v.Pathname,
			Component: v.Component,
			Title:     v.Title,
			Remark:    v.Remark,
			Ismenu:    v.Ismenu,
			Weigh:     v.Weigh,
			Status:    v.Status,
		}
		node.Children = child
		treelist = append(treelist, node)
	}
	return treelist

}

func GetRulesList(limit int, pagesize int, search *Authrule, order string) []*Authrule {
	var page int
	listdata := []*Authrule{}
	if pagesize-1 < 1 {
		page = 0
	} else {
		page = pagesize - 1
	}
	if limit <= 6 {
		limit = 6
	}

	session := global.Dorm.Table("auth_rule")

	// 关键修改：千万别漏了 search != nil 判断！
	if search != nil {
		if search.Id > 0 {
			session = session.And("id = ?", search.Id)
		}
		if search.Title != "" {
			session = session.And("title LIKE ?", "%"+search.Title+"%")
		}
		if search.Component != "" {
			session = session.And("component LIKE ?", "%"+search.Component+"%")
		}
		if search.Pathname != "" {
			session = session.And("pathname LIKE ?", "%"+search.Pathname+"%")
		}
	}

	var byorder string = "id ASC"
	if order != "" {
		byorder = order // 允许传入自定义排序规则，如 "weigh desc, id asc"
	}

	session.OrderBy(byorder).Limit(limit, limit*page).Find(&listdata)
	return listdata
}

func GetRulestotal(search *Authrule) int64 {
	var num int64
	session := global.Dorm.Table("auth_rule")
	if search.Id > 0 {
		session = session.And("id", search.Id)
	}
	if search.Title != "" {
		name := "%" + search.Title + "%"
		session = session.And("title LIKE ?", name)
	}
	if search.Component != "" {
		name := "%" + search.Component + "%"
		session = session.And("component LIKE ?", name)
	}
	if search.Pathname != "" {
		name := "%" + search.Pathname + "%"
		session = session.And("pathname LIKE ?", name)
	}
	a := new(Authrule)
	total, err := session.Count(a)
	if err == nil {
		num = total
	}
	return num
}

func DeleteRules(id int64) int {
	a := new(Authrule)
	outnum, _ := global.Dorm.ID(id).Delete(a)
	return int(outnum)
}

// 新增权限规则
func AddRules(a *Authrule) error {
	_, err := global.Dorm.Insert(a)
	return err
}

// 修改权限规则
// 修改权限规则
func EditRules(a *Authrule) error {
	_, err := global.Dorm.ID(a.Id).AllCols().Update(a)
	return err
}
