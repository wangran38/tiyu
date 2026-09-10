package models

import "time"

// TicketClaim 票根提交解析与领券记录，负责防刷和重复领取拦截。
type TicketClaim struct {
	ID             uint64 `xorm:"pk autoincr bigint 'id'" json:"id"`
	UserID         uint64 `xorm:"bigint notnull index(idx_user_claim) 'user_id' comment('领券用户ID')" json:"user_id"`
	ChannelType    string `xorm:"varchar(16) notnull 'channel_type' comment('通道类型: CHANNEL_A(标准) / CHANNEL_B(模版)')" json:"channel_type"`
	TicketCategory string `xorm:"varchar(32) notnull unique(uk_category_sn) 'ticket_category' comment('票据大类(TRAIN/FLIGHT/SCENIC/THIRD_PARTY/LOCAL_EVENT)')" json:"ticket_category"`
	TemplateID     uint64 `xorm:"bigint default 0 'template_id' comment('通道B使用的模版ID(通道A为0)')" json:"template_id"`
	UserImageURL   string `xorm:"varchar(512) notnull 'user_image_url' comment('用户上传的图片URL')" json:"user_image_url"`

	PHash          string `xorm:"varchar(64) notnull unique(uk_phash) 'p_hash' comment('图片感知哈希(防同图重复刷)')" json:"p_hash"`
	TicketSN       string `xorm:"varchar(128) notnull unique(uk_category_sn) 'ticket_sn' comment('票据流水号/车次号/订单号')" json:"ticket_sn"`
	HolderIdentity string `xorm:"varchar(128) default '' index(idx_identity) 'holder_identity' comment('身份标识(实名/身份证MD5/手机号)')" json:"holder_identity"`

	Title      string  `xorm:"varchar(255) 'title' comment('识别提取的标题')" json:"title"`
	HolderName string  `xorm:"varchar(64) 'holder_name' comment('识别提取的持票人')" json:"holder_name"`
	EventDate  string  `xorm:"varchar(32) 'event_date' comment('行程/活动日期')" json:"event_date"`
	Amount     float64 `xorm:"decimal(10,2) default 0.00 'amount' comment('票面金额')" json:"amount"`
	Confidence float64 `xorm:"decimal(4,3) 'confidence' comment('大模型识别置信度')" json:"confidence"`
	OCRRawJSON string  `xorm:"text 'ocr_raw_json' comment('Qwen2-VL 返回的原始 JSON 备份')" json:"ocr_raw_json"`

	PackageID    uint64    `xorm:"bigint notnull 'package_id' comment('最终触发发放的礼包ID')" json:"package_id"`
	ClaimStatus  int8      `xorm:"tinyint default 1 'claim_status' comment('1:成功发券 2:人工审核 3:驳回')" json:"claim_status"`
	RejectReason string    `xorm:"varchar(255) default '' 'reject_reason' comment('驳回原因')" json:"reject_reason"`
	CreatedAt    time.Time `xorm:"created index 'created_at'" json:"created_at"`
}

func (TicketClaim) TableName() string {
	return "ticket_claims"
}

func GetTicketClaimList(limit int, page int, search *TicketClaim, order string) []*TicketClaim {
	offset := page - 1
	if offset < 0 {
		offset = 0
	}
	byorder := "created_at DESC"
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

	query := Dorm.Table("ticket_claims")
	if search.UserID > 0 {
		query = query.And("user_id = ?", search.UserID)
	}
	if search.ChannelType != "" {
		query = query.And("channel_type = ?", search.ChannelType)
	}
	if search.TicketCategory != "" {
		query = query.And("ticket_category = ?", search.TicketCategory)
	}
	if search.TemplateID > 0 {
		query = query.And("template_id = ?", search.TemplateID)
	}
	if search.ClaimStatus > 0 {
		query = query.And("claim_status = ?", search.ClaimStatus)
	}
	if search.TicketSN != "" {
		query = query.And("ticket_sn like ?", "%"+search.TicketSN+"%")
	}
	if search.PHash != "" {
		query = query.And("p_hash = ?", search.PHash)
	}
	if search.Title != "" {
		query = query.And("(title like ? OR holder_name like ?)", "%"+search.Title+"%", "%"+search.Title+"%")
	}

	listdata := []*TicketClaim{}
	query.OrderBy(byorder).Limit(limit, limit*offset).Find(&listdata)
	return listdata
}

func GetTicketClaimTotal(search *TicketClaim) int64 {
	query := Dorm.NewSession()
	if search.UserID > 0 {
		query = query.Where("user_id = ?", search.UserID)
	}
	if search.ChannelType != "" {
		query = query.And("channel_type = ?", search.ChannelType)
	}
	if search.TicketCategory != "" {
		query = query.And("ticket_category = ?", search.TicketCategory)
	}
	if search.TemplateID > 0 {
		query = query.And("template_id = ?", search.TemplateID)
	}
	if search.ClaimStatus > 0 {
		query = query.And("claim_status = ?", search.ClaimStatus)
	}
	if search.TicketSN != "" {
		query = query.And("ticket_sn like ?", "%"+search.TicketSN+"%")
	}
	if search.PHash != "" {
		query = query.And("p_hash = ?", search.PHash)
	}
	if search.Title != "" {
		query = query.And("(title like ? OR holder_name like ?)", "%"+search.Title+"%", "%"+search.Title+"%")
	}

	total, err := query.Count(new(TicketClaim))
	if err != nil {
		return 0
	}
	return total
}

func AddTicketClaim(claim *TicketClaim) error {
	_, err := Dorm.Insert(claim)
	return err
}
