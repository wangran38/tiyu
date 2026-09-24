package shop

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"tiyu/global"
)

var (
	ErrTicketDiscountNotEnabled = errors.New("该商品未开启票根专属优惠")
	ErrTicketCategoryNotMatch   = errors.New("该票根类型不在当前商品支持的优惠票种范围内")
	ErrTicketCityNotMatch       = errors.New("票根到达城市与商品限定的城市不匹配")
	ErrTicketDateExpired        = errors.New("票根已超出优惠要求的有效天数")
	ErrTicketAlreadyExchanged   = errors.New("票根已被使用或兑换")
)

// GoodsTicketDiscount 票根识别联动优惠配置（凭交通票/门票/观演票 Enjoy 专属折扣）
type GoodsTicketDiscount struct {
	Id                    int64   `json:"id" xorm:"pk autoincr comment('主键ID')"`
	ProductId             int64   `json:"product_id" xorm:"unique notnull comment('关联团购产品ID')"`
	IsEnabled             int     `json:"is_enabled" xorm:"notnull default 0 comment('是否开启票根专属优惠: 0-关闭, 1-开启')"`
	AllowTicketCategories string  `json:"allow_ticket_categories" xorm:"varchar(255) comment('支持的票种分类，多选逗号分隔，例: \"火车票,飞机票,景区门票,体育赛事门票\"')"`
	MatchDestinationCity  string  `json:"match_destination_city" xorm:"varchar(64) default '' comment('限定到达城市，如限定为\"杭州\"到达票')"`
	DiscountType          int     `json:"discount_type" xorm:"notnull default 1 comment('优惠类型: 1-凭票直接立减固定金额, 2-凭票打折')"`
	DiscountValue         float64 `json:"discount_value" xorm:"decimal(10,2) notnull default 0.00 comment('优惠数值: 立减金额(元) 或 折扣率(如0.85代表85折)')"`
	TicketValidDays       int     `json:"ticket_valid_days" xorm:"default 3 comment('票根时效要求: 仅限票面时间在近X天内的票根')"`
}

func (td *GoodsTicketDiscount) TableName() string {
	return "goods_ticket_discount"
}

// GetGoodsTicketDiscountByProductID 获取产品绑定的票根规则
func GetGoodsTicketDiscountByProductID(productId int64) (*GoodsTicketDiscount, error) {
	td := new(GoodsTicketDiscount)
	has, err := global.Dorm.Where("product_id = ? AND is_enabled = 1", productId).Get(td)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return td, nil
}

// GetGoodsTicketDiscountRawByProductID 获取产品绑定的票根规则原始记录（不过滤 is_enabled）。
// 用于编辑回显：即使优惠已关闭，前端也要能拿到原配置展示"已关闭"状态。
func GetGoodsTicketDiscountRawByProductID(productId int64) (*GoodsTicketDiscount, error) {
	td := new(GoodsTicketDiscount)
	has, err := global.Dorm.Where("product_id = ?", productId).Get(td)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return td, nil
}

// DeleteGoodsTicketDiscountByProductID 按产品ID删除票根优惠配置（编辑时用：先删后插）
func DeleteGoodsTicketDiscountByProductID(productId int64) error {
	_, err := global.Dorm.Where("product_id = ?", productId).Delete(new(GoodsTicketDiscount))
	return err
}

// AddGoodsTicketDiscount 新增票根优惠配置
func AddGoodsTicketDiscount(td *GoodsTicketDiscount) error {
	_, err := global.Dorm.Insert(td)
	return err
}

// ValidateAndCalculate 校验票根信息是否符合要求并计算优惠金额
// 拆分参数解耦，避免循环引用 models 包
func (td *GoodsTicketDiscount) ValidateAndCalculate(
	exchangeStatus int8,
	ticketCategory string,
	ticketMainCategory string,
	title string,
	ocrRawJSON string,
	eventDate string,
	originalTotal float64,
) (float64, error) {
	// 1. 检查规则是否开启
	if td == nil || td.IsEnabled != 1 {
		return 0, ErrTicketDiscountNotEnabled
	}

	// 2. 检查票根使用状态 (0: 未兑换, 1: 已兑换)
	if exchangeStatus == 1 {
		return 0, ErrTicketAlreadyExchanged
	}

	// 3. 校验票种分类 (AllowTicketCategories)
	if td.AllowTicketCategories != "" {
		allowedCategories := strings.Split(td.AllowTicketCategories, ",")
		matched := false
		for _, cat := range allowedCategories {
			cat = strings.TrimSpace(cat)
			if cat != "" && (cat == ticketCategory || cat == ticketMainCategory) {
				matched = true
				break
			}
		}
		if !matched {
			return 0, fmt.Errorf("%w: 当前票根分类为 [%s]", ErrTicketCategoryNotMatch, ticketCategory)
		}
	}

	// 4. 校验到达城市 (MatchDestinationCity)
	if td.MatchDestinationCity != "" {
		targetCity := strings.TrimSpace(td.MatchDestinationCity)
		cleanTarget := strings.TrimSuffix(targetCity, "市")

		titleMatch := strings.Contains(title, cleanTarget)
		rawMatch := strings.Contains(ocrRawJSON, cleanTarget)

		if !titleMatch && !rawMatch {
			return 0, fmt.Errorf("%w: 需为到达 [%s] 的票根", ErrTicketCityNotMatch, targetCity)
		}
	}

	// 5. 校验票面时间时效 (TicketValidDays)
	if td.TicketValidDays > 0 && eventDate != "" {
		eventTime, err := parseEventDate(eventDate)
		if err == nil {
			now := time.Now()
			validDuration := time.Duration(td.TicketValidDays*24) * time.Hour
			diff := now.Sub(eventTime)
			if diff < 0 {
				diff = -diff
			}

			if diff > validDuration {
				return 0, fmt.Errorf("%w: 票面时间为 %s，需在近 %d 天内", ErrTicketDateExpired, eventDate, td.TicketValidDays)
			}
		}
	}

	// 6. 计算优惠折抵金额
	var discountAmount float64 = 0.0

	switch td.DiscountType {
	case 1: // 凭票立减固定金额
		discountAmount = td.DiscountValue
	case 2: // 凭票打折 (如 0.85 代表 85 折，优惠额 = 原价 * (1 - 0.85))
		if td.DiscountValue > 0 && td.DiscountValue < 1 {
			discountAmount = originalTotal * (1.0 - td.DiscountValue)
		}
	}

	// 防超扣
	if discountAmount > originalTotal {
		discountAmount = originalTotal
	}

	discountAmount = float64(int(discountAmount*100+0.5)) / 100.0

	return discountAmount, nil
}

// 辅助解析票面时间
func parseEventDate(dateStr string) (time.Time, error) {
	cleanStr := strings.TrimSpace(dateStr)
	cleanStr = regexp.MustCompile(`[年/.]`).ReplaceAllString(cleanStr, "-")
	cleanStr = strings.ReplaceAll(cleanStr, "月", "-")
	cleanStr = strings.ReplaceAll(cleanStr, "日", "")

	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}

	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, cleanStr, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("unsupported date format")
}
