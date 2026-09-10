package models

import (
	"time"
)

// SmsConfig 云短信服务配置
type SmsConfig struct {
	Id        int64     `json:"id" xDorm:"pk autoincr 'id'"`
	Provider  string    `json:"provider" xorm:"varchar(50) notnull comment('服务商标识: tencent/aliyun')"`
	Name      string    `json:"name" xorm:"varchar(100) notnull comment('配置名称')"`
	Config    string    `json:"config" xorm:"text notnull comment('秘钥/模板/签名等 JSON 配置')"`
	Status    int       `json:"status" xorm:"notnull default 1 comment('状态: 1-启用, 2-禁用')"`
	IsDefault int       `json:"is_default" xorm:"notnull default 2 comment('是否默认: 1-默认, 2-否')"`
	Created   time.Time `json:"createtime" xorm:"created int"`
	Updated   time.Time `json:"updatetime" xorm:"updated int"`
}

func (s *SmsConfig) TableName() string {
	return "sms_configs"
}

// 分页列表
func GetSmsConfigList(limit int, page int, search *SmsConfig, order string) []*SmsConfig {
	offset := page - 1
	if offset < 0 {
		offset = 0
	}
	byorder := "id ASC"
	if order == "-id" {
		byorder = "id DESC"
	}
	listdata := []*SmsConfig{}
	session := Dorm.Table("sms_configs")

	if search.Id > 0 {
		session = session.And("id = ?", search.Id)
	}
	if search.Provider != "" {
		session = session.And("provider = ?", search.Provider)
	}
	if search.Name != "" {
		session = session.And("name LIKE ?", "%"+search.Name+"%")
	}
	if search.Status > 0 {
		session = session.And("status = ?", search.Status)
	}
	if search.IsDefault > 0 {
		session = session.And("is_default = ?", search.IsDefault)
	}

	session.OrderBy(byorder).Limit(limit, limit*offset).Find(&listdata)
	return listdata
}

// 统计总数
func GetSmsConfigTotal(search *SmsConfig) int64 {
	session := Dorm.Table("sms_configs")

	if search.Id > 0 {
		session = session.And("id = ?", search.Id)
	}
	if search.Provider != "" {
		session = session.And("provider = ?", search.Provider)
	}
	if search.Name != "" {
		session = session.And("name LIKE ?", "%"+search.Name+"%")
	}
	if search.Status > 0 {
		session = session.And("status = ?", search.Status)
	}
	if search.IsDefault > 0 {
		session = session.And("is_default = ?", search.IsDefault)
	}

	total, err := session.Count(new(SmsConfig))
	if err != nil {
		return 0
	}
	return total
}

// 获取默认且启用的短信配置 (用于发送短信接口)
func GetDefaultSmsConfig() (*SmsConfig, error) {
	s := new(SmsConfig)
	has, err := Dorm.Table("sms_configs").
		Where("is_default = ? AND status = ?", 1, 1).
		Get(s)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return s, nil
}

// 新增
func AddSmsConfig(s *SmsConfig) error {
	// 如果设为了默认配置，需要先清空其他默认项
	if s.IsDefault == 1 {
		ClearOtherSmsDefault(0)
	}
	_, err := Dorm.Insert(s)
	return err
}

// 修改
func EditSmsConfig(s *SmsConfig) error {
	// 如果设为了默认配置，需要先清空其他默认项
	if s.IsDefault == 1 {
		ClearOtherSmsDefault(s.Id)
	}
	_, err := Dorm.ID(s.Id).Update(s)
	return err
}

// 删除
func DelSmsConfig(id int64) int {
	s := new(SmsConfig)
	outnum, _ := Dorm.ID(id).Delete(s)
	return int(outnum)
}

// 清除其他短信记录的默认标志 (保证全局仅有一条记录 is_default=1)
func ClearOtherSmsDefault(excludeId int64) {
	if excludeId > 0 {
		Dorm.Table("sms_configs").Where("id != ?", excludeId).Cols("is_default").Update(&SmsConfig{IsDefault: 2})
	} else {
		Dorm.Table("sms_configs").Cols("is_default").Update(&SmsConfig{IsDefault: 2})
	}
}
