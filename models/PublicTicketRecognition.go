package models

import (
	"time"
)

// PublicTicketRecognition 用于保存公开识别接口（非会员登录）识别出来的票根记录。
// 这个表独立于原来的 member_tickets，避免影响现有会员票根、兑换、订单等业务流程。
type PublicTicketRecognition struct {
	ID             uint64 `xorm:"pk autoincr bigint 'id'" json:"id"`
	UserID         uint64 `xorm:"bigint notnull index(idx_public_ticket_user) 'user_id'" json:"user_id"`
	ChannelType    string `xorm:"varchar(16) notnull 'channel_type'" json:"channel_type"`
	TicketCategory string `xorm:"varchar(64) notnull 'ticket_category'" json:"ticket_category"`
	ThirdPartyName string `xorm:"varchar(128) 'third_party_name'" json:"third_party_name"`
	UserImageURL   string `xorm:"varchar(512) notnull 'user_image_url'" json:"user_image_url"`
	PHash          string `xorm:"varchar(64) default '' 'p_hash'" json:"p_hash"`
	TicketSN       string `xorm:"varchar(128) notnull index(idx_public_ticket_sn) 'ticket_sn'" json:"ticket_sn"`

	Title         string  `xorm:"varchar(255) 'title'" json:"title"`
	HolderName    string  `xorm:"varchar(64) 'holder_name'" json:"holder_name"`
	EventDate     string  `xorm:"varchar(32) 'event_date'" json:"event_date"`
	Seat          string  `xorm:"varchar(64) 'seat'" json:"seat"`
	Amount        float64 `xorm:"decimal(10,2) default 0.00 'amount'" json:"amount"`
	Confidence    float64 `xorm:"decimal(4,3) 'confidence'" json:"confidence"`
	IsHandwritten bool    `xorm:"notnull default 0 'is_handwritten'" json:"is_handwritten"`
	OCRRawJSON    string  `xorm:"text 'ocr_raw_json'" json:"ocr_raw_json"`

	RecognizedAt time.Time `xorm:"notnull index 'recognized_at'" json:"recognized_at"`
	CreatedAt    time.Time `xorm:"created 'created_at'" json:"created_at"`
}

func (PublicTicketRecognition) TableName() string {
	return "public_ticket_recognitions"
}

// AddPublicTicketRecognition 保存公开识别接口成功识别的票根记录。
func AddPublicTicketRecognition(record *PublicTicketRecognition) error {
	_, err := Dorm.Insert(record)
	return err
}
