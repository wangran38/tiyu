package controllers

import (
	"net/http"

	"tiyu/models"

	"github.com/gin-gonic/gin"
)

// MemberTicketQueryRequest 票根列表查询请求参数结构体（controllers 层）。
// 该结构体与 models.MemberTicketQueryRequest 字段保持一一对应，
// 便于在 Handler 中完成参数绑定，再转换为 models 层结构体传入查询函数。
type MemberTicketQueryRequest struct {
	Limit          int    `json:"limit" form:"limit"`
	Page           int    `json:"page" form:"page"`
	Order          string `json:"order" form:"order"`
	ID             uint64 `json:"id" form:"id"`                                 // 票根主键ID
	UserID         uint64 `json:"user_id" form:"user_id"`                       // 会员ID
	TemplateID     uint64 `json:"template_id" form:"template_id"`               // 票根模板ID
	ChannelType    string `json:"channel_type" form:"channel_type"`             // 渠道类型
	TicketCategory string `json:"ticket_category" form:"ticket_category"`       // 票根分类
	TicketSN       string `json:"ticket_sn" form:"ticket_sn"`                   // 票根SN（模糊匹配）
	ExchangeStatus *int8  `json:"exchange_status" form:"exchange_status"`       // 兑换状态（指针：0-未兑换 / 1-已兑换）
	CouponID       uint64 `json:"coupon_id" form:"coupon_id"`                   // 关联优惠券ID

	// 创建时间段（对应 created_at 字段）
	CreatedStartTime string `json:"created_start_time" form:"created_start_time"` // 创建时间区间-开始
	CreatedEndTime   string `json:"created_end_time" form:"created_end_time"`     // 创建时间区间-结束

	// 兑换时间段（对应 exchanged_at 字段）
	ExchangedStartTime string `json:"exchanged_start_time" form:"exchanged_start_time"` // 兑换时间区间-开始
	ExchangedEndTime   string `json:"exchanged_end_time" form:"exchanged_end_time"`     // 兑换时间区间-结束
}

// MemberTicketResponse 票根响应结构体（可按需补充展示字段，如商户名、模板名等）。
// 当前仅继承 models.MemberTicket 所有字段，方便后续扩展。
type MemberTicketResponse struct {
	models.MemberTicket
}

// toModelQuery 将 controllers 层请求结构体转换为 models 层查询结构体。
func (r *MemberTicketQueryRequest) toModelQuery() *models.MemberTicketQueryRequest {
	return &models.MemberTicketQueryRequest{
		Limit:             r.Limit,
		Page:              r.Page,
		Order:             r.Order,
		ID:                r.ID,
		UserID:            r.UserID,
		TemplateID:        r.TemplateID,
		ChannelType:       r.ChannelType,
		TicketCategory:    r.TicketCategory,
		TicketSN:          r.TicketSN,
		ExchangeStatus:    r.ExchangeStatus,
		CouponID:          r.CouponID,
		CreatedStartTime:  r.CreatedStartTime,
		CreatedEndTime:    r.CreatedEndTime,
		ExchangedStartTime: r.ExchangedStartTime,
		ExchangedEndTime:   r.ExchangedEndTime,
	}
}

// GetMemberTicketListHandler 票根分页列表查询（后台通用接口）。
// 支持的查询条件：票根ID、用户ID、模板ID、渠道、分类、SN、兑换状态、优惠券ID、
// 创建时间段、兑换时间段等；同时返回分页信息 (page / limit) 与总数 (total)。
func GetMemberTicketListHandler(c *gin.Context) {
	var req MemberTicketQueryRequest
	// 同时支持 JSON 与 QueryString 绑定
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "参数错误",
			"data": err.Error(),
		})
		return
	}

	// 分页默认值兜底
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	// 转换为 models 层查询结构体
	modelReq := req.toModelQuery()

	// 调用 models 层封装的查询函数
	list := models.GetMemberTicketQueryList(req.Limit, req.Page, modelReq, req.Order)
	total := models.GetMemberTicketQueryTotal(modelReq)

	// 返回分页结果（与 Ordercontroller.go 风格保持一致）
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": gin.H{
			"list":  list,
			"total": total,
			"page":  req.Page,
			"limit": req.Limit,
		},
	})
}