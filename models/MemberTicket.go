package models

import (
	"errors"
	"time"
)

var (
	ErrMemberTicketNotFound = errors.New("member ticket not found or already exchanged")
	ErrCouponNotAvailable   = errors.New("coupon not available or sold out")
)

// MemberTicket 会员识别成功后的票根记录，用于“我的票根”和后续优惠券兑换。
type MemberTicket struct {
	ID             uint64 `xorm:"pk autoincr bigint 'id'" json:"id"`
	UserID         uint64 `xorm:"bigint notnull index(idx_member_ticket_user) 'user_id'" json:"user_id"`
	ChannelType    string `xorm:"varchar(16) notnull 'channel_type'" json:"channel_type"`
	TicketCategory string `xorm:"varchar(64) notnull 'ticket_category'" json:"ticket_category"`
	TemplateID     uint64 `xorm:"bigint default 0 'template_id'" json:"template_id"`
	UserImageURL   string `xorm:"varchar(512) notnull 'user_image_url'" json:"user_image_url"`
	PHash          string `xorm:"varchar(64) notnull index(idx_member_ticket_phash) 'p_hash'" json:"p_hash"`
	TicketSN       string `xorm:"varchar(128) notnull index(idx_member_ticket_sn) 'ticket_sn'" json:"ticket_sn"`

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
	_, err := Dorm.Insert(ticket)
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

	query := Dorm.Table("member_tickets")
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
	query := Dorm.Table("member_tickets")
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
	affected, err := Dorm.ID(id).
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
	session := Dorm.NewSession()
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
