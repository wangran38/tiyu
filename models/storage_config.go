package models

import (
	"encoding/json"
	"errors"
	"time"
)

type StorageConfig struct {
	ID        int64     `xorm:"pk autoincr 'id'" json:"id"`
	Name      string    `xorm:"varchar(100) notnull 'name'" json:"name"`
	Provider  string    `xorm:"varchar(32) notnull 'provider'" json:"provider"` // tencent / aliyun / qiniu / local
	Config    string    `xorm:"text notnull 'config'" json:"config"`            // JSON 字符串
	IsDefault int8      `xorm:"tinyint default 0 'is_default'" json:"is_default"`
	Status    int8      `xorm:"tinyint default 1 'status'" json:"status"`
	Remark    string    `xorm:"varchar(255) default '' 'remark'" json:"remark"`
	CreatedAt time.Time `xorm:"created 'created_at'" json:"created_at"`
	UpdatedAt time.Time `xorm:"updated 'updated_at'" json:"updated_at"`
}

func (StorageConfig) TableName() string {
	return "storage_configs"
}

// GetStorageConfigList 获取存储配置列表
func GetStorageConfigList(limit, page int, search *StorageConfig, order string) []*StorageConfig {
	list := make([]*StorageConfig, 0)
	session := Dorm.Table("storage_configs")

	if search.Provider != "" {
		session.And("provider = ?", search.Provider)
	}
	if search.Name != "" {
		session.And("name LIKE ?", "%"+search.Name+"%")
	}
	if search.Status > 0 {
		session.And("status = ?", search.Status)
	}
	if search.IsDefault > 0 {
		session.And("is_default = ?", search.IsDefault)
	}

	if order != "" {
		session.OrderBy(order)
	} else {
		session.OrderBy("id DESC")
	}

	if limit > 0 {
		offset := (page - 1) * limit
		session.Limit(limit, offset)
	}

	session.Find(&list)
	return list
}

// GetStorageConfigTotal 获取符合条件的总记录数
func GetStorageConfigTotal(search *StorageConfig) int64 {
	session := Dorm.Table("storage_configs")

	if search.Provider != "" {
		session.And("provider = ?", search.Provider)
	}
	if search.Name != "" {
		session.And("name LIKE ?", "%"+search.Name+"%")
	}
	if search.Status > 0 {
		session.And("status = ?", search.Status)
	}
	if search.IsDefault > 0 {
		session.And("is_default = ?", search.IsDefault)
	}

	total, err := session.Count(&StorageConfig{})
	if err != nil {
		return 0
	}
	return total
}

// GetStorageConfigByID 根据 ID 获取配置
func GetStorageConfigByID(id uint64) (*StorageConfig, error) {
	config := new(StorageConfig)
	has, err := Dorm.ID(id).Get(config)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return config, nil
}

// AddStorageConfig 新增存储配置
func AddStorageConfig(config *StorageConfig, rawConfig interface{}) error {
	if rawConfig != nil {
		bytes, err := json.Marshal(rawConfig)
		if err != nil {
			return errors.New("配置参数序列化失败")
		}
		config.Config = string(bytes)
	}

	session := Dorm.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	// 如果设置为了默认配置，需要将其他的置为非默认
	if config.IsDefault == 1 {
		if _, err := session.Table("storage_configs").Where("1=1").Update(map[string]interface{}{"is_default": 0}); err != nil {
			session.Rollback()
			return err
		}
	}

	if _, err := session.Insert(config); err != nil {
		session.Rollback()
		return err
	}

	return session.Commit()
}

// UpdateStorageConfig 更新存储配置
func UpdateStorageConfig(id uint64, config *StorageConfig, rawConfig interface{}) error {
	if rawConfig != nil {
		bytes, err := json.Marshal(rawConfig)
		if err != nil {
			return errors.New("配置参数序列化失败")
		}
		config.Config = string(bytes)
	}

	session := Dorm.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	if config.IsDefault == 1 {
		if _, err := session.Table("storage_configs").Where("id != ?", id).Update(map[string]interface{}{"is_default": 0}); err != nil {
			session.Rollback()
			return err
		}
	}

	if _, err := session.ID(id).AllCols().Update(config); err != nil {
		session.Rollback()
		return err
	}

	return session.Commit()
}

// DeleteStorageConfig 删除存储配置
func DeleteStorageConfig(id uint64) error {
	_, err := Dorm.ID(id).Delete(&StorageConfig{})
	return err
}

// SetDefaultStorageConfig 设置单个配置为默认
func SetDefaultStorageConfig(id uint64) error {
	session := Dorm.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	// 将所有记录取消默认
	if _, err := session.Table("storage_configs").Where("1=1").Update(map[string]interface{}{"is_default": 0}); err != nil {
		session.Rollback()
		return err
	}

	// 将当前 ID 设为默认与启用
	if _, err := session.Table("storage_configs").Where("id = ?", id).Update(map[string]interface{}{
		"is_default": 1,
		"status":     1,
	}); err != nil {
		session.Rollback()
		return err
	}

	return session.Commit()
}

// GetDefaultStorageConfig 获取当前启用的默认存储配置
func GetDefaultStorageConfig() (*StorageConfig, error) {
	config := new(StorageConfig)
	has, err := Dorm.Table("storage_configs").
		Where("is_default = ? AND status = ?", 1, 1).
		Get(config)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("未找到有效的默认存储配置")
	}
	return config, nil
}
