package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type JSONSlice []string

func (value JSONSlice) Value() (driver.Value, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

func (value *JSONSlice) Scan(src interface{}) error {
	if src == nil {
		*value = nil
		return nil
	}

	var data []byte
	switch source := src.(type) {
	case []byte:
		data = source
	case string:
		data = []byte(source)
	default:
		return fmt.Errorf("cannot scan %T into JSONSlice", src)
	}
	return json.Unmarshal(data, value)
}

type TicketTemplate struct {
	ID               uint64     `json:"id" xorm:"pk autoincr BIGINT comment('模版ID')"`
	TemplateName     string     `json:"template_name" xorm:"varchar(128) notnull comment('模版名称(如:2026村BA观赛票)')"`
	Category         string     `json:"category" xorm:"varchar(32) notnull index(idx_category_status) comment('票据大类(THIRD_PARTY/LOCAL_EVENT/SCENIC)')"`
	SampleImageURLs  JSONSlice  `json:"sample_image_urls" xorm:"json notnull comment('运营上传的标杆样例图URL列表')"`
	RequiredKeywords string     `json:"required_keywords" xorm:"varchar(255) default '' comment('必须包含的文本关键字(逗号分隔)')"`
	CheckIsTorn      bool       `json:"check_is_torn" xorm:"tinyint(1) default 0 comment('是否强制校验手撕痕迹')"`
	PackageID        uint64     `json:"package_id" xorm:"BIGINT notnull comment('绑定发券的大礼包ID(CouponPackage)')"`
	Status           int8       `json:"status" xorm:"tinyint default 1 index(idx_category_status) comment('状态(1:启用 0:停用)')"`
	CreatedAt        time.Time  `json:"created_at" xorm:"created"`
	UpdatedAt        time.Time  `json:"updated_at" xorm:"updated"`
	DeletedAt        *time.Time `json:"-" xorm:"datetime index"`
}

func (t *TicketTemplate) TableName() string {
	return "ticket_template"
}

func GetTicketTemplateList(limit int, page int, search string, order string) []*TicketTemplate {
	offset := page - 1
	if offset < 0 {
		offset = 0
	}
	byorder := "id ASC"
	switch order {
	case "id":
		byorder = "id ASC"
	case "-id":
		byorder = "id DESC"
	case "created_at":
		byorder = "created_at ASC"
	case "-created_at":
		byorder = "created_at DESC"
	}
	listdata := []*TicketTemplate{}
	Dorm.Table("ticket_template").
		Where("deleted_at IS NULL").
		Where("template_name like ?", "%"+search+"%").
		OrderBy(byorder).
		Limit(limit, limit*offset).
		Find(&listdata)
	return listdata
}

func GetTicketTemplateTotal(search string) int64 {
	template := new(TicketTemplate)
	total, err := Dorm.Where("deleted_at IS NULL").And("template_name like ?", "%"+search+"%").Count(template)
	if err != nil {
		return 0
	}
	return total
}

func AddTicketTemplate(template *TicketTemplate) error {
	_, err := Dorm.Insert(template)
	return err
}

func EditTicketTemplate(template *TicketTemplate) error {
	_, err := Dorm.ID(template.ID).Where("deleted_at IS NULL").Update(template)
	return err
}

func DelTicketTemplate(id uint64) int {
	deletedAt := time.Now()
	outnum, _ := Dorm.ID(id).Where("deleted_at IS NULL").Cols("deleted_at").Update(&TicketTemplate{DeletedAt: &deletedAt})
	return int(outnum)
}
