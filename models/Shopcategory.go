package models

import (
	"time"
)

// ShopCategory 入驻商家分类表
type ShopCategory struct {
	ID uint64 `xorm:"pk autoincr bigint 'id'" json:"id"`
	// 将 uint64 修改为 int64，以便接收前端传来的 -1（代表不限父级/全部）
	ParentID    int64  `xorm:"bigint default 0 index 'parent_id' comment('父级分类ID，0为顶级分类')" json:"parent_id"`
	Name        string `xorm:"varchar(64) notnull 'name' comment('分类名称')" json:"name"`
	Code        string `xorm:"varchar(32) notnull unique 'code' comment('分类编码')" json:"code"`
	Icon        string `xorm:"varchar(512) default '' 'icon' comment('分类图标URL')" json:"icon"`
	Sort        int32  `xorm:"integer default 0 index 'sort' comment('排序权重(越大越靠前)')" json:"sort"`
	Status      int8   `xorm:"tinyint default 1 index 'status' comment('状态 1:启用 2:禁用')" json:"status"`
	Description string `xorm:"varchar(255) default '' 'description' comment('分类描述')" json:"description"`

	CreatedAt time.Time `xorm:"created index 'created_at'" json:"created_at"`
	UpdatedAt time.Time `xorm:"updated 'updated_at'" json:"updated_at"`
	DeletedAt time.Time `xorm:"deleted index 'deleted_at'" json:"deleted_at"` // 启用 xorm 软删除
}

func (ShopCategory) TableName() string {
	return "shop_categories"
}

// GetShopCategoryList 分页查询商家分类列表
func GetShopCategoryList(limit int, page int, search *ShopCategory, order string) ([]*ShopCategory, error) {
	offset := page - 1
	if offset < 0 {
		offset = 0
	}
	byorder := "sort DESC, id ASC"
	switch order {
	case "id":
		byorder = "id ASC"
	case "-id":
		byorder = "id DESC"
	case "sort":
		byorder = "sort ASC"
	case "-sort":
		byorder = "sort DESC"
	case "created_at":
		byorder = "created_at ASC"
	case "-created_at":
		byorder = "created_at DESC"
	}

	query := Dorm.Table("shop_categories")

	// -------------------- 约定逻辑 --------------------
	// 1. ParentID == -1：查全部分类 (不拼 parent_id 条件)
	// 2. ParentID >= 0：查指定父级的分类 (如 0 表示仅查顶级，>0 表示查对应子级)
	if search.ParentID >= 0 {
		query = query.And("parent_id = ?", search.ParentID)
	}
	// --------------------------------------------------

	if search.Status > 0 {
		query = query.And("status = ?", search.Status)
	}
	if search.Code != "" {
		query = query.And("code = ?", search.Code)
	}
	if search.Name != "" {
		query = query.And("name like ?", "%"+search.Name+"%")
	}

	listdata := []*ShopCategory{}
	err := query.OrderBy(byorder).Limit(limit, limit*offset).Find(&listdata)
	return listdata, err
}

// GetShopCategoryTotal 获取符合条件的分类总条数
func GetShopCategoryTotal(search *ShopCategory) int64 {
	session := Dorm.NewSession()
	defer session.Close()

	// -------------------- 约定逻辑 --------------------
	// ParentID >= 0 时拼条件；-1 时跳过 parent_id 约束查总数
	if search.ParentID >= 0 {
		session = session.And("parent_id = ?", search.ParentID)
	}
	// --------------------------------------------------

	if search.Status > 0 {
		session = session.And("status = ?", search.Status)
	}
	if search.Code != "" {
		session = session.And("code = ?", search.Code)
	}
	if search.Name != "" {
		session = session.And("name like ?", "%"+search.Name+"%")
	}

	total, err := session.Count(new(ShopCategory))
	if err != nil {
		return 0
	}
	return total
}

// AddShopCategory 新增商家分类
func AddShopCategory(category *ShopCategory) error {
	_, err := Dorm.Insert(category)
	return err
}

// UpdateShopCategory 更新商家分类
func UpdateShopCategory(id uint64, category *ShopCategory) error {
	_, err := Dorm.ID(id).Update(category)
	return err
}

// DeleteShopCategory 删除商家分类（触发软删除）
func DeleteShopCategory(id uint64) error {
	_, err := Dorm.ID(id).Delete(new(ShopCategory))
	return err
}

// GetAllActiveShopCategories 获取所有启用状态的商家分类（用于构建分类树）
func GetAllActiveShopCategories() ([]*ShopCategory, error) {
	categories := make([]*ShopCategory, 0)
	err := Dorm.Table("shop_categories").
		Where("status = ?", 1).
		OrderBy("sort DESC, id ASC").
		Find(&categories)
	return categories, err
}
