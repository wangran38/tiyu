package models

import (
	"errors"
	"time"
	"tiyu/global"
)

var (
	ErrMemberTicketNotFound = errors.New("member ticket not found or already exchanged")
	ErrCouponNotAvailable   = errors.New("coupon not available or sold out")
)

// MemberTicket 会员识别成功后的票根记录，用于“我的票根”和后续优惠券兑换。
type MemberTicket struct {
	ID                 uint64 `xorm:"pk autoincr bigint 'id'" json:"id"`
	UserID             uint64 `xorm:"bigint notnull index(idx_member_ticket_user) 'user_id'" json:"user_id"`
	ChannelType        string `xorm:"varchar(16) notnull 'channel_type'" json:"channel_type"`
	TicketCategory     string `xorm:"varchar(64) notnull 'ticket_category'" json:"ticket_category"`
	TicketMainCategory string `xorm:"varchar(64) notnull 'ticket_main_category'" json:"ticket_main_category"`
	TemplateID         uint64 `xorm:"bigint default 0 'template_id'" json:"template_id"`
	UserImageURL       string `xorm:"varchar(512) notnull 'user_image_url'" json:"user_image_url"`
	PHash              string `xorm:"varchar(64) notnull index(idx_member_ticket_phash) 'p_hash'" json:"p_hash"`
	TicketSN           string `xorm:"varchar(128) notnull index(idx_member_ticket_sn) 'ticket_sn'" json:"ticket_sn"`

	Title         string  `xorm:"varchar(255) 'title'" json:"title"`
	HolderName    string  `xorm:"varchar(64) 'holder_name'" json:"holder_name"`
	EventDate     string  `xorm:"varchar(32) 'event_date'" json:"event_date"`
	Seat          string  `xorm:"varchar(64) 'seat'" json:"seat"`
	Amount        float64 `xorm:"decimal(10,2) default 0.00 'amount'" json:"amount"`
	Confidence    float64 `xorm:"decimal(4,3) 'confidence'" json:"confidence"`
	IsHandwritten bool    `xorm:"notnull default 0 'is_handwritten'" json:"is_handwritten"`
	OCRRawJSON    string  `xorm:"text 'ocr_raw_json'" json:"ocr_raw_json"`

	ExchangeStatus int8       `xorm:"tinyint notnull default 0 index 'exchange_status' comment('0:未兑换 1:已兑换')" json:"exchange_status"`
	CouponID       uint64     `xorm:"bigint default 0 'coupon_id'" json:"coupon_id"`
	ExchangedAt    *time.Time `xorm:"datetime null 'exchanged_at'" json:"exchanged_at"`
	RecognizedAt   time.Time  `xorm:"notnull index 'recognized_at'" json:"recognized_at"`
	CreatedAt      time.Time  `xorm:"created 'created_at'" json:"created_at"`
}

func (MemberTicket) TableName() string {
	return "member_tickets"
}

// AddMemberTicket 保存一条识别成功的会员票根。
func AddMemberTicket(ticket *MemberTicket) error {
	_, err := global.Dorm.Insert(ticket)
	return err
}

// GetMemberTicketList 获取指定会员的票根分页列表。
func GetMemberTicketList(limit, page int, search *MemberTicket, order string) []*MemberTicket {
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	byOrder := "recognized_at DESC"
	switch order {
	case "id":
		byOrder = "id ASC"
	case "-id":
		byOrder = "id DESC"
	case "recognized_at":
		byOrder = "recognized_at ASC"
	}

	query := global.Dorm.Table("member_tickets")
	if search != nil {
		if search.UserID > 0 {
			query = query.And("user_id = ?", search.UserID)
		}
		if search.ExchangeStatus > 0 {
			query = query.And("exchange_status = ?", search.ExchangeStatus)
		}
		if search.TicketCategory != "" {
			query = query.And("ticket_category = ?", search.TicketCategory)
		}
	}

	list := []*MemberTicket{}
	query.OrderBy(byOrder).Limit(limit, limit*(page-1)).Find(&list)
	return list
}

// GetMemberTicketTotal 获取指定会员的票根总数。
func GetMemberTicketTotal(search *MemberTicket) int64 {
	query := global.Dorm.Table("member_tickets")
	if search != nil {
		if search.UserID > 0 {
			query = query.And("user_id = ?", search.UserID)
		}
		if search.ExchangeStatus > 0 {
			query = query.And("exchange_status = ?", search.ExchangeStatus)
		}
		if search.TicketCategory != "" {
			query = query.And("ticket_category = ?", search.TicketCategory)
		}
	}

	total, err := query.Count(new(MemberTicket))
	if err != nil {
		return 0
	}
	return total
}

// ExchangeMemberTicket 将未兑换票根标记为已兑换，并记录优惠券和兑换时间。
func ExchangeMemberTicket(id, userID, couponID uint64) (bool, error) {
	now := time.Now()
	affected, err := global.Dorm.ID(id).
		Where("user_id = ? AND exchange_status = 0", userID).
		Cols("exchange_status", "coupon_id", "exchanged_at").
		Update(&MemberTicket{
			ExchangeStatus: 1,
			CouponID:       couponID,
			ExchangedAt:    &now,
		})
	return affected > 0, err
}

// RedeemMemberTicket 使用会员票根兑换优惠券，并原子更新票根和优惠券库存。
func RedeemMemberTicket(ticketID, userID, couponID uint64) (*MemberTicket, *Coupon, error) {
	session := global.Dorm.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return nil, nil, err
	}
	rollback := true
	defer func() {
		if rollback {
			_ = session.Rollback()
		}
	}()

	ticket := new(MemberTicket)
	has, err := session.ID(ticketID).Where("user_id = ? AND exchange_status = 0", userID).Get(ticket)
	if err != nil {
		return nil, nil, err
	}
	if !has {
		return nil, nil, ErrMemberTicketNotFound
	}

	coupon := new(Coupon)
	has, err = session.ID(couponID).Get(coupon)
	if err != nil {
		return nil, nil, err
	}
	if !has || coupon.Status != 1 || coupon.ReceiveCount >= coupon.TotalCount || time.Now().Before(coupon.StartTime) || time.Now().After(coupon.EndTime) {
		return nil, nil, ErrCouponNotAvailable
	}

	affected, err := session.ID(couponID).
		Where("status = 1 AND receive_count < total_count").
		Incr("receive_count", 1).
		Update(new(Coupon))
	if err != nil {
		return nil, nil, err
	}
	if affected == 0 {
		return nil, nil, ErrCouponNotAvailable
	}

	now := time.Now()
	affected, err = session.ID(ticketID).
		Where("user_id = ? AND exchange_status = 0", userID).
		Cols("exchange_status", "coupon_id", "exchanged_at").
		Update(&MemberTicket{
			ExchangeStatus: 1,
			CouponID:       couponID,
			ExchangedAt:    &now,
		})
	if err != nil {
		return nil, nil, err
	}
	if affected == 0 {
		return nil, nil, ErrMemberTicketNotFound
	}

	if err := session.Commit(); err != nil {
		return nil, nil, err
	}
	rollback = false
	ticket.ExchangeStatus = 1
	ticket.CouponID = couponID
	ticket.ExchangedAt = &now
	coupon.ReceiveCount++
	return ticket, coupon, nil
}

// MemberTicketQueryRequest 会员票根列表查询请求结构体（支持兑换时间段 + 创建时间段过滤）。
// 用法：传入此结构体作为 search 参数，配合 GetMemberTicketQueryList / GetMemberTicketQueryTotal 使用。
// 时间字段约定：使用字符串（如 "2024-05-01 00:00:00" 或 RFC3339 字符串），为空则忽略该条件。
type MemberTicketQueryRequest struct {
	Limit          int    `json:"limit" form:"limit"`
	Page           int    `json:"page" form:"page"`
	Order          string `json:"order" form:"order"`
	ID             uint64 `json:"id" form:"id"`                           // 票根主键ID
	UserID         uint64 `json:"user_id" form:"user_id"`                 // 会员ID
	TemplateID     uint64 `json:"template_id" form:"template_id"`         // 票根模板ID
	ChannelType    string `json:"channel_type" form:"channel_type"`       // 渠道类型
	TicketCategory string `json:"ticket_category" form:"ticket_category"` // 票根分类
	TicketSN       string `json:"ticket_sn" form:"ticket_sn"`             // 票根SN（支持模糊匹配）
	ExchangeStatus *int8  `json:"exchange_status" form:"exchange_status"` // 兑换状态（用指针支持 0:未兑换 / 1:已兑换 的精确过滤）
	CouponID       uint64 `json:"coupon_id" form:"coupon_id"`             // 关联优惠券ID

	// 创建时间段：对应票根 created_at / recognized_at 字段（按 created_at 过滤）。
	CreatedStartTime string `json:"created_start_time" form:"created_start_time"` // 创建时间区间-开始
	CreatedEndTime   string `json:"created_end_time" form:"created_end_time"`     // 创建时间区间-结束

	// 兑换时间段：对应票根 exchanged_at 字段，未兑换的记录该字段为 NULL，需注意空值过滤。
	ExchangedStartTime string `json:"exchanged_start_time" form:"exchanged_start_time"` // 兑换时间区间-开始
	ExchangedEndTime   string `json:"exchanged_end_time" form:"exchanged_end_time"`     // 兑换时间区间-结束
}

// GetMemberTicketQueryList 获取票根分页列表（适配 MemberTicketQueryRequest 全字段过滤）。
func GetMemberTicketQueryList(limit, page int, search *MemberTicketQueryRequest, order string) []*MemberTicket {
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	byOrder := "created_at DESC"
	switch order {
	case "id":
		byOrder = "id ASC"
	case "-id":
		byOrder = "id DESC"
	case "created_at":
		byOrder = "created_at ASC"
	case "-created_at":
		byOrder = "created_at DESC"
	case "recognized_at":
		byOrder = "recognized_at ASC"
	case "-recognized_at":
		byOrder = "recognized_at DESC"
	case "exchanged_at":
		byOrder = "exchanged_at ASC"
	case "-exchanged_at":
		byOrder = "exchanged_at DESC"
	}

	query := global.Dorm.Table("member_tickets")
	if search != nil {
		// 1. 主键 ID
		if search.ID > 0 {
			query = query.And("id = ?", search.ID)
		}
		// 2. 会员ID
		if search.UserID > 0 {
			query = query.And("user_id = ?", search.UserID)
		}
		// 3. 票根模板ID
		if search.TemplateID > 0 {
			query = query.And("template_id = ?", search.TemplateID)
		}
		// 4. 渠道类型（精确匹配）
		if search.ChannelType != "" {
			query = query.And("channel_type = ?", search.ChannelType)
		}
		// 5. 票根分类（精确匹配）
		if search.TicketCategory != "" {
			query = query.And("ticket_category = ?", search.TicketCategory)
		}
		// 6. 票根SN（模糊匹配）
		if search.TicketSN != "" {
			query = query.And("ticket_sn like ?", "%"+search.TicketSN+"%")
		}
		// 7. 兑换状态（通过指针判断，支持 0 和 1 精确过滤）
		if search.ExchangeStatus != nil {
			query = query.And("exchange_status = ?", *search.ExchangeStatus)
		}
		// 8. 关联优惠券ID
		if search.CouponID > 0 {
			query = query.And("coupon_id = ?", search.CouponID)
		}
		// 9. 创建时间区间（对应 created_at 字段）
		if search.CreatedStartTime != "" {
			query = query.And("created_at >= ?", search.CreatedStartTime)
		}
		if search.CreatedEndTime != "" {
			query = query.And("created_at <= ?", search.CreatedEndTime)
		}
		// 10. 兑换时间区间（对应 exchanged_at 字段）
		if search.ExchangedStartTime != "" {
			query = query.And("exchanged_at >= ?", search.ExchangedStartTime)
		}
		if search.ExchangedEndTime != "" {
			query = query.And("exchanged_at <= ?", search.ExchangedEndTime)
		}
	}

	list := []*MemberTicket{}
	query.OrderBy(byOrder).Limit(limit, limit*(page-1)).Find(&list)
	return list
}

// GetMemberTicketQueryTotal 获取票根总数（适配 MemberTicketQueryRequest 全字段过滤）。
func GetMemberTicketQueryTotal(search *MemberTicketQueryRequest) int64 {
	query := global.Dorm.Table("member_tickets")
	if search != nil {
		if search.ID > 0 {
			query = query.And("id = ?", search.ID)
		}
		if search.UserID > 0 {
			query = query.And("user_id = ?", search.UserID)
		}
		if search.TemplateID > 0 {
			query = query.And("template_id = ?", search.TemplateID)
		}
		if search.ChannelType != "" {
			query = query.And("channel_type = ?", search.ChannelType)
		}
		if search.TicketCategory != "" {
			query = query.And("ticket_category = ?", search.TicketCategory)
		}
		if search.TicketSN != "" {
			query = query.And("ticket_sn like ?", "%"+search.TicketSN+"%")
		}
		if search.ExchangeStatus != nil {
			query = query.And("exchange_status = ?", *search.ExchangeStatus)
		}
		if search.CouponID > 0 {
			query = query.And("coupon_id = ?", search.CouponID)
		}
		// 创建时间区间
		if search.CreatedStartTime != "" {
			query = query.And("created_at >= ?", search.CreatedStartTime)
		}
		if search.CreatedEndTime != "" {
			query = query.And("created_at <= ?", search.CreatedEndTime)
		}
		// 兑换时间区间
		if search.ExchangedStartTime != "" {
			query = query.And("exchanged_at >= ?", search.ExchangedStartTime)
		}
		if search.ExchangedEndTime != "" {
			query = query.And("exchanged_at <= ?", search.ExchangedEndTime)
		}
	}

	total, err := query.Count(new(MemberTicket))
	if err != nil {
		return 0
	}
	return total
}

// GetMemberTicketByIDAndUserID 根据票根ID和用户ID获取未兑换/有效的票根
func GetMemberTicketByIDAndUserID(id uint64, userID uint64) (*MemberTicket, error) {
	ticket := new(MemberTicket)
	has, err := global.Dorm.Where("id = ? AND user_id = ?", id, userID).Get(ticket)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return ticket, nil
}
