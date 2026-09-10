package controllers

import (
	"encoding/json"
	"tiyu/models"

	"github.com/gin-gonic/gin"
)

// ================= 请求结构体 (DTO) =================

type SmsConfigSearch struct {
	Name      string `json:"name"`
	Provider  string `json:"provider"`
	Status    int    `json:"status"`
	IsDefault int    `json:"is_default"`
	Limit     int    `json:"limit"`
	Page      int    `json:"page"`
	Order     string `json:"sort"`
}

type SmsConfigDelReq struct {
	Id int64 `json:"id" binding:"required"`
}

// 新增短信配置请求参数
type SmsConfigAddReq struct {
	Provider  string `json:"provider" binding:"required"` // 服务商标识: tencent/aliyun
	Name      string `json:"name" binding:"required"`     // 配置名称
	Config    string `json:"config" binding:"required"`   // 秘钥/模板/签名等 JSON 字符串
	Status    int    `json:"status"`                      // 状态：1启用 2禁用
	IsDefault int    `json:"is_default"`                  // 是否默认：1默认 2否
}

// 编辑/修改短信配置请求参数
type SmsConfigEditReq struct {
	Id        int64  `json:"id" binding:"required"`       // 要修改的主键ID
	Provider  string `json:"provider" binding:"required"` // 服务商标识
	Name      string `json:"name" binding:"required"`     // 配置名称
	Config    string `json:"config" binding:"required"`   // 秘钥/模板/签名等 JSON 字符串
	Status    int    `json:"status"`                      // 状态
	IsDefault int    `json:"is_default"`                  // 是否默认
}

// 内部验证 JSON 结构的辅助结构体
type smsConfigDetail struct {
	SecretID   string `json:"secret_id"`
	SecretKey  string `json:"secret_key"`
	AppID      string `json:"app_id"`
	SignName   string `json:"sign_name"`
	TemplateID string `json:"template_id"`
}

// ================= 控制器方法 =================

// 获取短信配置列表 (POST)
func GetSmsConfigList(c *gin.Context) {
	var searchdata SmsConfigSearch
	c.BindJSON(&searchdata)

	result := make(map[string]interface{})

	limit := searchdata.Limit
	page := searchdata.Page
	order := searchdata.Order

	search := &models.SmsConfig{
		Name:      searchdata.Name,
		Provider:  searchdata.Provider,
		Status:    searchdata.Status,
		IsDefault: searchdata.IsDefault,
	}

	listdata := models.GetSmsConfigList(limit, page, search, order)
	listnum := models.GetSmsConfigTotal(search)

	result["page"] = page
	result["totalnum"] = listnum
	result["limit"] = limit

	if listdata == nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "获取短信配置列表失败",
			"data":    "",
		})
		return
	} else {
		result["listdata"] = listdata
		c.JSON(200, gin.H{
			"code":    200,
			"message": "数据获取成功",
			"data":    result,
		})
		return
	}
}

// 删除短信配置 (POST)
func DelSmsConfig(c *gin.Context) {
	var req SmsConfigDelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "参数格式错误",
			"data":    "",
		})
		return
	}

	affectedRows := models.DelSmsConfig(req.Id)
	if affectedRows <= 0 {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "删除失败，记录不存在或已被删除",
			"data":    "",
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "删除成功",
		"data":    "",
	})
}

// 新增短信配置 (POST)
func AddSmsConfig(c *gin.Context) {
	var req SmsConfigAddReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "参数格式错误或必填项缺失",
			"data":    "",
		})
		return
	}

	// 校验 JSON 配置字符串格式合法性
	var detail smsConfigDetail
	if err := json.Unmarshal([]byte(req.Config), &detail); err != nil {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "Config 格式不符合 JSON 标准",
			"data":    "",
		})
		return
	}

	// 默认状态填充
	if req.Status == 0 {
		req.Status = 1
	}
	if req.IsDefault == 0 {
		req.IsDefault = 2
	}

	// 实例化结构体并传入 models.AddSmsConfig
	smsConfig := &models.SmsConfig{
		Provider:  req.Provider,
		Name:      req.Name,
		Config:    req.Config,
		Status:    req.Status,
		IsDefault: req.IsDefault,
	}

	err := models.AddSmsConfig(smsConfig)
	if err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "添加失败: " + err.Error(),
			"data":    "",
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "添加成功",
		"data":    "",
	})
}

// 修改/编辑短信配置 (POST)
func EditSmsConfig(c *gin.Context) {
	var req SmsConfigEditReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "参数格式错误或必填项缺失",
			"data":    "",
		})
		return
	}

	// 校验 JSON 配置字符串格式合法性
	var detail smsConfigDetail
	if err := json.Unmarshal([]byte(req.Config), &detail); err != nil {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "Config 格式不符合 JSON 标准",
			"data":    "",
		})
		return
	}

	// 实例化结构体并传入 models.EditSmsConfig
	smsConfig := &models.SmsConfig{
		Id:        req.Id,
		Provider:  req.Provider,
		Name:      req.Name,
		Config:    req.Config,
		Status:    req.Status,
		IsDefault: req.IsDefault,
	}

	err := models.EditSmsConfig(smsConfig)
	if err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "修改失败: " + err.Error(),
			"data":    "",
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "修改成功",
		"data":    "",
	})
}
