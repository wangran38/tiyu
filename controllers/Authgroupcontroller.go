package controllers

import (
	"tiyu/models"

	"github.com/gin-gonic/gin"
)

// ================= 请求结构体 (DTO) =================

type Groupserch struct {
	Name  string `json:"name"`
	Limit int    `json:"limit"`
	Page  int    `json:"page"`
	Order string `json:"sort"`
}

type GroupDelReq struct {
	Id int64 `json:"id" binding:"required"`
}

// 新增组别请求参数（字段已与 Authgroup 模型统一）
type GroupAddReq struct {
	Name   string `json:"name" binding:"required"` // 组别名称
	Pid    int64  `json:"pid"`                     // 父级ID，默认为0 (顶层)
	Status int    `json:"status"`                  // 状态：1启用 0禁用
	Rules  string `json:"rules"`                   // 关联权限规则节点
}

// 编辑/修改组别请求参数
type GroupEditReq struct {
	Id     int64  `json:"id" binding:"required"`   // 要修改的主键ID
	Name   string `json:"name" binding:"required"` // 组别名称
	Pid    int64  `json:"pid"`                     // 父级ID
	Status int    `json:"status"`                  // 状态
	Rules  string `json:"rules"`                   // 关联权限规则节点
}

// ================= 控制器方法 =================

// 获取组别列表
func Getgrouplist(c *gin.Context) {
	var searchdata Groupserch
	c.BindJSON(&searchdata)

	result := make(map[string]interface{})

	limit := searchdata.Limit
	page := searchdata.Page
	name := searchdata.Name
	order := searchdata.Order

	listdata, err := models.GetgroupList(limit, page, name, order)
	listnum := models.Getgrouptotal(name)

	result["page"] = page
	result["totalnum"] = listnum
	result["limit"] = limit

	if err != nil || listdata == nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "获取菜单失败1",
			"data":    "",
		})
		return
	} else {
		result["listdata"] = listdata
		c.JSON(200, gin.H{
			"code":    200,
			"message": "数据获取成功1",
			"data":    result,
		})
		return
	}
}

// 获取组别树（用于选择器下拉框）
func Getgrouptree(c *gin.Context) {
	treeData, err := models.GetGroupTree()

	if err != nil || treeData == nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "获取组别树失败",
			"data":    "",
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "数据获取成功",
		"data":    treeData,
	})
}

// 删除组别 (POST)
func Delgroup(c *gin.Context) {
	var req GroupDelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "参数格式错误",
			"data":    "",
		})
		return
	}

	// 修正：接收 (int, error) 两个返回值
	affectedRows, err := models.Delgroup(req.Id)
	if err != nil || affectedRows <= 0 {
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

// 新增组别 (POST)
func Addgroup(c *gin.Context) {
	var req GroupAddReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "参数格式错误或必填项缺失",
			"data":    "",
		})
		return
	}

	// 实例化结构体并传入 models.Addgroup
	group := &models.Authgroup{
		Name:   req.Name,
		Pid:    req.Pid,
		Status: req.Status,
		Rules:  req.Rules,
	}

	err := models.Addgroup(group)
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

// 修改/编辑组别 (POST)
func Editgroup(c *gin.Context) {
	var req GroupEditReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "参数格式错误或必填项缺失",
			"data":    "",
		})
		return
	}

	// 校验不能将自身设为父节点
	if req.Id == int64(req.Pid) {
		c.JSON(200, gin.H{
			"code":    400,
			"message": "上级节点不能选择自身",
			"data":    "",
		})
		return
	}

	// 实例化结构体并传入 models.Editgroup
	group := &models.Authgroup{
		Id:     req.Id,
		Name:   req.Name,
		Pid:    req.Pid,
		Status: req.Status,
		Rules:  req.Rules,
	}

	err := models.Editgroup(group)
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
