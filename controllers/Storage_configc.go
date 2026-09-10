package controllers

import (
	"tiyu/models"

	"github.com/gin-gonic/gin"
)

type StorageConfigSearch struct {
	Name      string `json:"name"`
	Provider  string `json:"provider"`
	Status    int8   `json:"status"`
	IsDefault int8   `json:"is_default"`
	Keyword   string `json:"keyword"`
	Limit     int    `json:"limit"`
	Page      int    `json:"page"`
	Order     string `json:"sort"`
}

type StorageConfigReq struct {
	ID        uint64      `json:"id"`
	Name      string      `json:"name"`
	Provider  string      `json:"provider"`
	IsDefault int8        `json:"is_default"`
	Status    int8        `json:"status"`
	Remark    string      `json:"remark"`
	Config    interface{} `json:"config"`
}

// 获取存储配置列表
func GetStorageConfiglist(c *gin.Context) {
	var searchdata StorageConfigSearch
	if err := c.BindJSON(&searchdata); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "参数错误",
			"data":    err.Error(),
		})
		return
	}

	search := &models.StorageConfig{
		Name:      searchdata.Name,
		Provider:  searchdata.Provider,
		Status:    searchdata.Status,
		IsDefault: searchdata.IsDefault,
	}
	if searchdata.Keyword != "" && search.Name == "" {
		search.Name = searchdata.Keyword
	}

	result := map[string]interface{}{
		"page":     searchdata.Page,
		"totalnum": models.GetStorageConfigTotal(search),
		"limit":    searchdata.Limit,
		"listdata": models.GetStorageConfigList(searchdata.Limit, searchdata.Page, search, searchdata.Order),
	}
	c.JSON(200, gin.H{
		"code":    200,
		"message": "获取存储配置列表成功",
		"data":    result,
	})
}

// 获取存储配置详情
func GetStorageConfigDetail(c *gin.Context) {
	var param struct {
		ID uint64 `json:"id"`
	}
	if err := c.BindJSON(&param); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "参数错误",
			"data":    err.Error(),
		})
		return
	}

	if param.ID == 0 {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "缺少 id 参数",
			"data":    "",
		})
		return
	}

	detail, err := models.GetStorageConfigByID(param.ID)
	if err != nil || detail == nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "数据不存在或获取失败",
			"data":    "",
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "获取详情成功",
		"data":    detail,
	})
}

// 新增存储配置
func AddStorageConfig(c *gin.Context) {
	var req StorageConfigReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "参数错误",
			"data":    err.Error(),
		})
		return
	}

	if req.Name == "" || req.Provider == "" {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "缺少配置名称(name)或服务商类型(provider)",
			"data":    "",
		})
		return
	}

	config := &models.StorageConfig{
		Name:      req.Name,
		Provider:  req.Provider,
		IsDefault: req.IsDefault,
		Status:    req.Status,
		Remark:    req.Remark,
	}

	if err := models.AddStorageConfig(config, req.Config); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "新增存储配置失败",
			"data":    err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "新增存储配置成功",
		"data":    config,
	})
}

// 修改存储配置
func UpdateStorageConfig(c *gin.Context) {
	var req StorageConfigReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "参数错误",
			"data":    err.Error(),
		})
		return
	}

	if req.ID == 0 {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "缺少记录 id",
			"data":    "",
		})
		return
	}

	exist, err := models.GetStorageConfigByID(req.ID)
	if err != nil || exist == nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "数据不存在或已被删除",
			"data":    "",
		})
		return
	}

	exist.Name = req.Name
	exist.Provider = req.Provider
	exist.IsDefault = req.IsDefault
	exist.Status = req.Status
	exist.Remark = req.Remark

	if err := models.UpdateStorageConfig(req.ID, exist, req.Config); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "更新存储配置失败",
			"data":    err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "修改存储配置成功",
		"data":    exist,
	})
}

// 删除存储配置
func DeleteStorageConfig(c *gin.Context) {
	var param struct {
		ID uint64 `json:"id"`
	}
	if err := c.BindJSON(&param); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "参数错误",
			"data":    err.Error(),
		})
		return
	}

	if param.ID == 0 {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "缺少 id 参数",
			"data":    "",
		})
		return
	}

	if err := models.DeleteStorageConfig(param.ID); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "删除失败",
			"data":    err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "删除成功",
		"data":    "",
	})
}

// 设为默认存储配置
func SetDefaultStorageConfig(c *gin.Context) {
	var param struct {
		ID uint64 `json:"id"`
	}
	if err := c.BindJSON(&param); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "参数错误",
			"data":    err.Error(),
		})
		return
	}

	if param.ID == 0 {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "缺少 id 参数",
			"data":    "",
		})
		return
	}

	if err := models.SetDefaultStorageConfig(param.ID); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "设置默认存储失败",
			"data":    err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "设置默认存储成功",
		"data":    "",
	})
}
