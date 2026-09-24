package models

import (
	"time"
	"tiyu/global"
)

// User 会员表（前台用户，区别于后台 admin 表）
type User struct {
	Id          int64     `json:"id" xorm:"pk autoincr bigint 'id'"`
	Openid      string    `json:"openid" xorm:"varchar(64) index 'openid' comment('微信 OpenID')"`
	Unionid     string    `json:"unionid" xorm:"varchar(64) index 'unionid' comment('微信 UnionID')"`
	Username    string    `json:"username" xorm:"varchar(60) comment('用户名')"`
	Password    string    `json:"-" xorm:"varchar(64) comment('密码 MD5')"`
	Salt        string    `json:"-" xorm:"varchar(20) comment('密码盐')"`
	Nickname    string    `json:"nickname" xorm:"varchar(60) comment('昵称')"`
	Avatar      string    `json:"avatar" xorm:"varchar(512) comment('头像 URL')"`
	Realname    string    `json:"realname" xorm:"varchar(60) comment('真实姓名')"`
	Mobile      string    `json:"mobile" xorm:"varchar(20) index 'mobile' comment('手机号')"`
	Email       string    `json:"email" xorm:"varchar(100) comment('邮箱')"`
	Gender      int       `json:"gender" xorm:"tinyint default 0 comment('性别 0未知 1男 2女')"`
	Birthday    string    `json:"birthday" xorm:"varchar(10) comment('生日 YYYY-MM-DD')"`
	Point       int       `json:"point" xorm:"notnull default 0 comment('积分')"`
	Level       int       `json:"level" xorm:"notnull default 1 comment('会员等级')"`
	LastLoginIp string    `json:"last_login_ip" xorm:"varchar(60) comment('最后登录 IP')"`
	LastLoginAt int64     `json:"last_login_at" xorm:"bigint comment('最后登录时间戳')"`
	Status      string    `json:"status" xorm:"varchar(40) notnull default 'normal' comment('状态 normal/hidden/banned')"`
	Created     time.Time `json:"created_at" xorm:"created int"`
	Updated     time.Time `json:"updated_at" xorm:"updated int"`
}

func (User) TableName() string {
	return "users"
}

// 分页列表
func GetMemberList(limit int, page int, search *User, order string) []*User {
	offset := page - 1
	if offset < 0 {
		offset = 0
	}
	byorder := "id ASC"
	if order == "-id" {
		byorder = "id DESC"
	}
	listdata := []*User{}
	session := global.Dorm.Table("users")
	if search.Id > 0 {
		session = session.And("id = ?", search.Id)
	}
	if search.Username != "" {
		session = session.And("username LIKE ?", "%"+search.Username+"%")
	}
	if search.Mobile != "" {
		session = session.And("mobile LIKE ?", "%"+search.Mobile+"%")
	}
	if search.Nickname != "" {
		session = session.And("nickname LIKE ?", "%"+search.Nickname+"%")
	}
	if search.Status != "" {
		session = session.And("status = ?", search.Status)
	}
	session.OrderBy(byorder).Limit(limit, limit*offset).Find(&listdata)
	return listdata
}

func GetMemberTotal(search *User) int64 {
	session := global.Dorm.Table("users")
	if search.Id > 0 {
		session = session.And("id = ?", search.Id)
	}
	if search.Username != "" {
		session = session.And("username LIKE ?", "%"+search.Username+"%")
	}
	if search.Mobile != "" {
		session = session.And("mobile LIKE ?", "%"+search.Mobile+"%")
	}
	if search.Nickname != "" {
		session = session.And("nickname LIKE ?", "%"+search.Nickname+"%")
	}
	if search.Status != "" {
		session = session.And("status = ?", search.Status)
	}
	total, err := session.Count(new(User))
	if err != nil {
		return 0
	}
	return total
}

// 新增
func AddMember(a *User) error {
	_, err := global.Dorm.Insert(a)
	return err
}

// 修改
// 修改会员信息
func EditMember(a *User) error {
	_, err := global.Dorm.ID(a.Id).AllCols().Update(a)
	return err
}

// 删除
func DelMember(id int64) int {
	outnum, _ := global.Dorm.ID(id).Delete(new(User))
	return int(outnum)
}

// 根据用户名查询（登录校验）
func SelectMemberByUsername(username string) (*User, error) {
	u := new(User)
	has, err := global.Dorm.Where("username = ?", username).Get(u)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return u, nil
}

// 根据 ID 查询（积分/等级调整用）
func SelectMemberById(id int64) (*User, error) {
	u := new(User)
	has, err := global.Dorm.ID(id).Get(u)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return u, nil
}

// FindOrCreateUserByMobile 按手机号查询，不存在则自动注册 (XORM 标准实现)
func FindOrCreateUserByMobile(mobile string, nickname string) (*User, error) {
	u := new(User)
	// 1. 查询手机号是否存在
	has, err := global.Dorm.Where("mobile = ?", mobile).Get(u)
	if err != nil {
		return nil, err
	}

	// 2. 如果已存在，直接返回用户对象
	if has {
		return u, nil
	}

	// 3. 不存在，则自动创建新用户
	// 如果传入的 nickname 为空，默认用手机号后 4 位掩码生成默认昵称
	if nickname == "" && len(mobile) >= 11 {
		nickname = "用户_" + mobile[7:]
	} else if nickname == "" {
		nickname = "新用户"
	}

	newUser := &User{
		Mobile:   mobile,
		Username: mobile, // 默认用户名设置为手机号
		Nickname: nickname,
		Status:   "normal", // 注意：结构体中 Status 为 string 类型，需填 'nglobal.Dormal'
		Point:    0,
		Level:    1,
	}

	// 4. 插入数据库
	_, err = global.Dorm.Insert(newUser)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}
