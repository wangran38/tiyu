package models

import (
	"errors"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Admin struct {
	Id           int64
	Username     string    `xDorm:"varchar(200)" json:"username"`
	Nickname     string    `xDorm:"varchar(200)" json:"nickname"`
	Salt         string    `xDorm:"varchar(200)" json:"salt"`
	Age          int       `xDorm:"int(2)" json:"age"`
	Avatar       string    `xDorm:"TEXT" json:"avatar"`
	Loginfailure int       `xDorm:"int(10)" json:"loginfailure"`
	Logintime    int       `xDorm:"int(10)" json:"logintime"`
	Loginip      string    `xDorm:"varchar(200)" json:"loginip"`
	Token        string    `xDorm:"varchar(59)"`
	Password     string    `xDorm:"varchar(200)"`
	Created      time.Time `xDorm:"created"`
	Updated      time.Time `xDorm:"updated"`
}

type Adminjson struct {
	Id        int64  `json:"id"`
	Gid       int64  `json:"gid"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Groupname string `json:"groupname"`
}
type AdminFlatJson struct {
	Id        int64  `xorm:"admin_id"`
	Gid       int64  `xorm:"group_id"`
	Username  string `xorm:"username"`
	Nickname  string `xorm:"nickname"`
	Avatar    string `xorm:"avatar"`
	Groupname string `xorm:"group_name"`
}

func (a *Admin) TableName() string {
	return "admin"
}

type AdminGroup struct {
	Admin     `xorm:"extends"`
	Authgroup struct {
		Id   int64  `xorm:"group_id"`   // 对应上面 g.id AS group_id
		Name string `xorm:"group_name"` // 对应上面 g.name AS group_name
	} `xorm:"extends"`
}

func (AdminGroup) TableName() string {
	return "admin"
}

// 根据用户名密码查询用户
func SelectUserByUserName(userName string) (*Admin, error) {
	a := new(Admin)
	has, err := Dorm.Where("username = ?", userName).Get(a)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("用户未找到！")
	}
	return a, nil
}

// 根据用户 id (string类型) 查询用户
func SelectAdminById(Id string) (*Admin, error) {
	a := new(Admin)
	id, _ := strconv.ParseInt(Id, 10, 64)
	has, err := Dorm.Where("id = ?", id).Get(a)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("用户未找到！")
	}
	return a, nil
}

// 根据用户 id (int64类型) 查询用户 (新增：供编辑用户时校验使用)
func SelectUserById(uid int64) (*Admin, error) {
	a := new(Admin)
	has, err := Dorm.Where("id = ?", uid).Get(a)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("用户未找到！")
	}
	return a, nil
}

// 添加管理员用户
func AddAdmin(a *Admin) error {
	_, err := Dorm.Insert(a)
	return err
}

// 分页列表
func GetUserList(limit int, pagesize int, search string, order string) []*Adminjson {
	page := pagesize - 1
	if page < 0 {
		page = 0
	}

	var adminlist []*Adminjson

	var byorder string = "a.id ASC"
	if order == "-id" {
		byorder = "a.id DESC"
	}

	// 将别名直接写成与结构体字段名一致
	query := Dorm.Table("admin").Alias("a").
		Select("a.id AS id, ac.gid AS gid, a.username AS username, a.nickname AS nickname, a.avatar AS avatar, g.name AS groupname").
		Join("INNER", []string{"auth_group_access", "ac"}, "ac.uid = a.id").
		Join("INNER", []string{"auth_group", "g"}, "g.id = ac.gid")

	if search != "" {
		query = query.Where("a.username like ?", "%"+search+"%")
	}

	query.OrderBy(byorder).Limit(limit, limit*page).Find(&adminlist)

	return adminlist
}

// 获取用户总数
func GetUsertotal(search string) int64 {
	var num int64 = 0
	a := new(Admin)
	if search != "" {
		total, err := Dorm.Cols("id", "username").Where("username like ?", "%"+search+"%").Count(a)
		if err == nil {
			num = total
		}
	} else {
		total, err := Dorm.Cols("id", "username").Count(a)
		if err == nil {
			num = total
		}
	}
	return num
}

// ======================= 新增/更新功能 (带事务) =======================

// UpdateAdminWithGroup 在事务中一并更新 admin 表数据和组别映射 auth_group_access 表
func UpdateAdminWithGroup(uid int64, username string, groupID int64) error {
	session := Dorm.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	// 1. 如果传了 username，更新 admin 表
	if username != "" {
		admin := &Admin{
			Username: username,
			Updated:  time.Now(),
		}
		_, err := session.Where("id = ?", uid).Cols("username", "updated").Update(admin)
		if err != nil {
			session.Rollback()
			return err
		}
	}

	// 2. 如果传了 groupID，更新/追加关联表 auth_group_access
	if groupID > 0 {
		access := &Authaccess{Gid: groupID}
		affected, err := session.Where("uid = ?", uid).Update(access)
		if err != nil {
			session.Rollback()
			return err
		}
		// 如果影响行数为 0，说明此前该用户没有关联记录，重新插入
		if affected == 0 {
			access.Uid = uid
			_, err = session.Insert(access)
			if err != nil {
				session.Rollback()
				return err
			}
		}
	}

	return session.Commit()
}

// UpdateAdminGroup 单独修改组别关联的方法
// func UpdateAdminGroup(uid int64, gid int64) error {
// 	access := &Authaccess{Gid: gid}
// 	affected, err := Dorm.Where("uid = ?", uid).Update(access)
// 	if err != nil {
// 		return err
// 	}
// 	if affected == 0 {
// 		access.Uid = uid
// 		_, err = Dorm.Insert(access)
// 	}
// 	return err
// }

// // UpdateAdminPassword 更新用户密码和 Salt
// func UpdateAdminPassword(uid int64, password string, salt string) error {
// 	admin := &Admin{
// 		Password: password,
// 		Salt:     salt,
// 		Updated:  time.Now(),
// 	}
// 	_, err := Dorm.Where("id = ?", uid).Cols("password", "salt", "updated").Update(admin)
// 	return err
// }

// DeleteAdminWithAccess 根据用户 ID 删除用户及对应的组别关联 (带事务)
func DeleteAdminWithAccess(uid int64) error {
	session := Dorm.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	// 1. 删除 admin 表中的记录
	admin := &Admin{Id: uid}
	_, err := session.Delete(admin)
	if err != nil {
		session.Rollback()
		return err
	}

	// 2. 删除 auth_group_access 表中的关联记录
	access := &Authaccess{Uid: uid}
	_, err = session.Where("uid = ?", uid).Delete(access)
	if err != nil {
		session.Rollback()
		return err
	}

	// 提交事务
	return session.Commit()
}
