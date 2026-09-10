package models

import (
	"fmt"
	"math"
	"time"
)

// Shop 店铺主表 (O2O线下商家信息表)
type Shop struct {
	ID         int64  `xorm:"pk autoincr bigint 'id'" json:"id"`
	MerchantNo string `xorm:"varchar(32) notnull unique 'merchant_no' comment('商家业务编号(如 M202608280001)')" json:"merchant_no"`
	UserID     int64  `xorm:"bigint index 'user_id' comment('绑定的平台掌柜/主账号ID')" json:"user_id"`
	CategoryID int64  `xorm:"bigint index 'category_id' comment('主营分类ID(关联 shop_categories)')" json:"category_id"`
	CityID     int64  `xorm:"bigint index 'city_id' comment('城市ID(关联 city 表 Id)')" json:"city_id"`
	Status     int8   `xorm:"tinyint default 0 index 'status' comment('状态 0:待审核 1:营业中 2:休息中 3:已冻结 4:审核驳回')" json:"status"`

	// --- 基础信息 ---
	Name         string `xorm:"varchar(128) notnull 'name' comment('门店名称')" json:"name"`
	Logo         string `xorm:"varchar(512) default '' 'logo' comment('门店Logo图片URL')" json:"logo"`
	CoverImages  string `xorm:"text 'cover_images' comment('门头照/环境图URL列表(JSON数组)')" json:"cover_images"`
	ContactName  string `xorm:"varchar(64) notnull 'contact_name' comment('负责人姓名')" json:"contact_name"`
	ContactPhone string `xorm:"varchar(20) notnull 'contact_phone' comment('负责人联系电话')" json:"contact_phone"`
	ServicePhone string `xorm:"varchar(20) default '' 'service_phone' comment('对外客服/订座电话')" json:"service_phone"`
	Description  string `xorm:"text 'description' comment('商家简介(长文本)')" json:"description"`
	Discounts    string `xorm:"text 'discounts' comment('优惠内容/促销活动(长文本)')" json:"discounts"`

	// --- 线下地理位置信息 (O2O 核心) ---
	ProvinceCode string  `xorm:"varchar(12) default '' 'province_code' comment('省份行政区划代码')" json:"province_code"`
	CityCode     string  `xorm:"varchar(12) default '' 'city_code' comment('城市行政区划代码')" json:"city_code"`
	DistrictCode string  `xorm:"varchar(12) default '' 'district_code' comment('区县行政区划代码')" json:"district_code"`
	Address      string  `xorm:"varchar(255) notnull 'address' comment('详细地址')" json:"address"`
	Longitude    float64 `xorm:"decimal(10,7) notnull index 'longitude' comment('经度')" json:"longitude"`
	Latitude     float64 `xorm:"decimal(10,7) notnull index 'latitude' comment('纬度')" json:"latitude"`
	GeoHash      string  `xorm:"varchar(12) default '' index 'geo_hash' comment('GeoHash空间索引字串')" json:"geo_hash"`

	// --- 运营与配送配置 ---
	OpeningHours   string  `xorm:"varchar(255) default '' 'opening_hours' comment('营业时间段配置(如 09:00-22:00)')" json:"opening_hours"`
	DeliveryType   int8    `xorm:"tinyint default 1 'delivery_type' comment('配送模式 1:商家自送 2:平台专送 3:仅到店')" json:"delivery_type"`
	DeliveryRadius float64 `xorm:"decimal(5,2) default 3.00 'delivery_radius' comment('配送半径(公里)')" json:"delivery_radius"`
	MinOrderAmount float64 `xorm:"decimal(10,2) default 0.00 'min_order_amount' comment('起送金额(元)')" json:"min_order_amount"`
	AvgCost        float64 `xorm:"decimal(10,2) default 0.00 'avg_cost' comment('人均消费(元)')" json:"avg_cost"`

	// --- 抽成与手续费配置 ---
	CommissionRate     float64 `xorm:"decimal(5,2) default 0.00 'commission_rate' comment('平台抽成/佣金比例(%, 如5.50表示5.5%)')" json:"commission_rate"`
	ServiceFee         float64 `xorm:"decimal(10,2) default 0.00 'service_fee' comment('固定平台服务手续费/单(元)')" json:"service_fee"`
	TransactionFeeRate float64 `xorm:"decimal(5,2) default 0.00 'transaction_fee_rate' comment('支付通道/交易手续费率(%, 如0.60表示0.6%)')" json:"transaction_fee_rate"`

	// --- 额外扩展参数 (不映射到 DB 列，用于经纬度搜索) ---
	UserLng    float64 `xorm:"-" json:"user_lng,omitempty"`
	UserLat    float64 `xorm:"-" json:"user_lat,omitempty"`
	RadiusKm   float64 `xorm:"-" json:"radius_km,omitempty"`
	DistanceKm float64 `xorm:"-" json:"distance_km,omitempty"`

	// --- 基础审计时间戳 ---
	CreatedAt time.Time `xorm:"created index 'created_at'" json:"created_at"`
	UpdatedAt time.Time `xorm:"updated 'updated_at'" json:"updated_at"`
	DeletedAt time.Time `xorm:"index 'deleted_at'" json:"deleted_at"` // 👈 彻底去掉了 deleted 关键字
}

func (Shop) TableName() string {
	return "shops"
}

// ----------------------------------------------------
// 查询与统计逻辑
// ----------------------------------------------------

// GetShopList 条件与分页查询店铺列表
func GetShopList(limit int, page int, search *Shop, startTime string, endTime string, order string) ([]*Shop, error) {
	offset := page - 1
	if offset < 0 {
		offset = 0
	}

	byorder := "id DESC"
	hasLocation := search.UserLng >= -180 && search.UserLng <= 180 && search.UserLat >= -90 && search.UserLat <= 90 && (search.UserLng != 0 || search.UserLat != 0)
	switch order {
	case "id":
		byorder = "id ASC"
	case "-id":
		byorder = "id DESC"
	case "created_at":
		byorder = "created_at ASC"
	case "-created_at":
		byorder = "created_at DESC"
	case "commission_rate":
		byorder = "commission_rate ASC"
	case "-commission_rate":
		byorder = "commission_rate DESC"
	case "distance":
		if hasLocation {
			byorder = "distance_km ASC"
		}
	case "-distance", "distance_desc":
		if hasLocation {
			byorder = "distance_km DESC"
		}
	case "distance_asc":
		if hasLocation {
			byorder = "distance_km ASC"
		}
	}

	query := Dorm.Table("shops")

	if search.Status >= 0 {
		query = query.And("status = ?", search.Status)
	}

	if search.CategoryID > 0 {
		query = query.And("category_id = ?", search.CategoryID)
	}
	if search.CityID > 0 {
		query = query.And("city_id = ?", search.CityID)
	}
	if search.UserID > 0 {
		query = query.And("user_id = ?", search.UserID)
	}
	if search.MerchantNo != "" {
		query = query.And("merchant_no = ?", search.MerchantNo)
	}
	if search.CityCode != "" {
		query = query.And("city_code = ?", search.CityCode)
	}
	if search.Name != "" {
		query = query.And("name like ?", "%"+search.Name+"%")
	}

	if startTime != "" && endTime != "" {
		query = query.And("created_at BETWEEN ? AND ?", startTime, endTime)
	}

	// 地理位置筛选
	if hasLocation {
		distanceExpr := fmt.Sprintf(
			"( 6371 * acos( cos( radians(%f) ) * cos( radians(latitude) ) * cos( radians(longitude) - radians(%f) ) + sin( radians(%f) ) * sin( radians(latitude) ) ) )",
			search.UserLat, search.UserLng, search.UserLat,
		)

		if search.RadiusKm > 0 {
			query = query.And(fmt.Sprintf("%s <= ?", distanceExpr), search.RadiusKm)
		}

		query = query.Select(fmt.Sprintf("shops.*, %s AS distance_km", distanceExpr))
	}

	listdata := []*Shop{}
	err := query.OrderBy(byorder).Limit(limit, limit*offset).Find(&listdata)
	if err == nil && hasLocation {
		for _, shop := range listdata {
			shop.DistanceKm = calculateDistanceKm(search.UserLat, search.UserLng, shop.Latitude, shop.Longitude)
		}
	}
	return listdata, err
}

func calculateDistanceKm(userLat, userLng, shopLat, shopLng float64) float64 {
	const earthRadiusKm = 6371.0
	lat1 := userLat * math.Pi / 180
	lat2 := shopLat * math.Pi / 180
	deltaLat := (shopLat - userLat) * math.Pi / 180
	deltaLng := (shopLng - userLng) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	return earthRadiusKm * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// GetShopTotal 获取符合条件的店铺总条数
func GetShopTotal(search *Shop, startTime string, endTime string) int64 {
	session := Dorm.NewSession()
	defer session.Close()

	if search.Status >= 0 {
		session = session.And("status = ?", search.Status)
	}
	if search.CategoryID > 0 {
		session = session.And("category_id = ?", search.CategoryID)
	}
	if search.CityID > 0 {
		session = session.And("city_id = ?", search.CityID)
	}
	if search.UserID > 0 {
		session = session.And("user_id = ?", search.UserID)
	}
	if search.MerchantNo != "" {
		session = session.And("merchant_no = ?", search.MerchantNo)
	}
	if search.CityCode != "" {
		session = session.And("city_code = ?", search.CityCode)
	}
	if search.Name != "" {
		session = session.And("name like ?", "%"+search.Name+"%")
	}

	if startTime != "" && endTime != "" {
		session = session.And("created_at BETWEEN ? AND ?", startTime, endTime)
	}

	hasLocation := search.UserLng >= -180 && search.UserLng <= 180 && search.UserLat >= -90 && search.UserLat <= 90 && (search.UserLng != 0 || search.UserLat != 0)
	if hasLocation && search.RadiusKm > 0 {
		distanceExpr := fmt.Sprintf(
			"( 6371 * acos( cos( radians(%f) ) * cos( radians(latitude) ) * cos( radians(longitude) - radians(%f) ) + sin( radians(%f) ) * sin( radians(latitude) ) ) )",
			search.UserLat, search.UserLng, search.UserLat,
		)
		session = session.And(fmt.Sprintf("%s <= ?", distanceExpr), search.RadiusKm)
	}

	total, err := session.Count(new(Shop))
	if err != nil {
		return 0
	}
	return total
}

// ----------------------------------------------------
// 单条与基础 CRUD 方法
// ----------------------------------------------------

func GetShopByID(id int64) (*Shop, error) {
	shop := new(Shop)
	has, err := Dorm.ID(id).Get(shop)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return shop, nil
}

func GetShopByUserID(userID int64) (*Shop, error) {
	shop := new(Shop)
	has, err := Dorm.Where("user_id = ?", userID).Get(shop)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return shop, nil
}

func AddShop(shop *Shop) error {
	_, err := Dorm.Insert(shop)
	return err
}

func UpdateShop(id int64, shop *Shop) error {
	_, err := Dorm.ID(id).AllCols().Update(shop)
	return err
}

func DeleteShop(id int64) error {
	_, err := Dorm.ID(id).Delete(new(Shop))
	return err
}
