package shop

import "tiyu/global"

// GoodsRule 产品核销规则表（预订规则、不可用日期、退改规则等）
type GoodsRule struct {
	Id                      int64  `json:"id" xorm:"pk autoincr comment('规则ID')"`
	ProductId               int64  `json:"product_id" xorm:"unique notnull comment('关联团购产品ID')"`
	NeedAppointment         int    `json:"need_appointment" xorm:"notnull default 0 comment('预约规则: 0-免预约, 1-需提前预约, 2-指定场次/日期使用')"`
	AppointmentAdvanceHours int    `json:"appointment_advance_hours" xorm:"default 0 comment('需提前多少小时预约(例: 24代表提前1天)')"`
	NeedRealName            int    `json:"need_real_name" xorm:"notnull default 0 comment('是否需要出行人实名信息: 0-不需要, 1-需要')"`
	AppointmentPhone        string `json:"appointment_phone" xorm:"varchar(32) default '' comment('商家预约/咨询电话')"`
	LimitPerUser            int    `json:"limit_per_user" xorm:"default 0 comment('单用户限购张数，0为不限')"`
	LimitPerOrder           int    `json:"limit_per_order" xorm:"default 0 comment('单次下单/单桌限用张数，0为不限')"`
	UsableTimes             string `json:"usable_times" xorm:"json comment('每周可用时段 JSON，例如: [{\"weeks\":[1,2,3,4,5],\"start\":\"11:00\",\"end\":\"21:00\"}]')"`
	UnusableDates           string `json:"unusable_dates" xorm:"json comment('不可用日期 JSON 数组，如节假日: [\"2026-10-01\",\"2026-10-02\"]')"`
	UseInstructions         string `json:"use_instructions" xorm:"json comment('使用须知文本数组 JSON，如: [\"仅限堂食\",\"不与店里其他优惠同享\"]')"`
	RefundRule              int    `json:"refund_rule" xorm:"notnull default 1 comment('退改规则: 1-随时退/过期自动退, 2-条件退, 3-不可退')"`
	RefundAdvanceHours      int    `json:"refund_advance_hours" xorm:"default 0 comment('条件退款需提前的小时数')"`
}

func (r *GoodsRule) TableName() string {
	return "goods_rule"
}

// GetGoodsRuleByProductID 获取指定产品的使用规则
func GetGoodsRuleByProductID(productId int64) (*GoodsRule, error) {
	rule := new(GoodsRule)
	has, err := global.Dorm.Where("product_id = ?", productId).Get(rule)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return rule, nil
}

// DeleteGoodsRuleByProductID 按产品ID删除核销规则（编辑时用：先删后插）
func DeleteGoodsRuleByProductID(productId int64) error {
	_, err := global.Dorm.Where("product_id = ?", productId).Delete(new(GoodsRule))
	return err
}

// AddGoodsRule 新增核销规则
func AddGoodsRule(rule *GoodsRule) error {
	_, err := global.Dorm.Insert(rule)
	return err
}
