package controllers

import (
	"strconv"
	"strings"
	"time"
	"tiyu/models"

	"github.com/gin-gonic/gin"
)

// 规则查询参数结构体
type Rulesserch struct {
	Id        int64  `json:"id"`
	Pid       int64  `json:"pid"`
	Title     string `json:"title"`
	Pathname  string `json:"pathname"`
	Component string `json:"component"`
	Limit     int    `json:"limit"`
	Page      int    `json:"page"`
	Order     string `json:"order"`
}

// 规则列表展示结构体
type Rulestable struct {
	Id        int64     `json:"id"`
	Pid       int64     `json:"pid"`
	Title     string    `json:"title"`
	Icon      string    `json:"icon"`
	Pathname  string    `json:"pathname"`
	Component string    `json:"component"`
	Type      string    `json:"type"`
	Ismenu    int       `json:"ismenu"`
	Weigh     int       `json:"weigh"`
	Status    string    `json:"status"`
	Created   time.Time `json:"createtime"`
}

// 规则树形节点结构体（专门给前端下拉树/动态菜单组件使用）
type RulesTreeNode struct {
	Id        int64            `json:"id"`
	Pid       int64            `json:"pid"`
	Title     string           `json:"title"`
	Icon      string           `json:"icon"`
	Pathname  string           `json:"pathname"`
	Component string           `json:"component"`
	Type      string           `json:"type"`
	Ismenu    int              `json:"ismenu"`
	Weigh     int              `json:"weigh"`
	Status    string           `json:"status"`
	Created   time.Time        `json:"createtime"`
	Children  []*RulesTreeNode `json:"children,omitempty"` // 为空时自动忽略 JSON 输出
}

// 请求根据 rules 获取菜单树的参数结构
type GetMenuByRulesReq struct {
	Rules string `json:"rules"` // 例如 "*" 或 "1,2,3,4"
}

// 1. 获取规则分页/条件列表 (适用于 Table 表格)
func Getruleslist(c *gin.Context) {
	var searchdata Rulesserch
	c.ShouldBindJSON(&searchdata)

	limit := searchdata.Limit
	page := searchdata.Page
	order := searchdata.Order
	result := make(map[string]interface{})

	search := &models.Authrule{
		Id:        searchdata.Id,
		Pid:       searchdata.Pid,
		Title:     searchdata.Title,
		Pathname:  searchdata.Pathname,
		Component: searchdata.Component,
	}

	listdata := models.GetRulesList(limit, page, search, order)
	tabledata := []*Rulestable{}
	for _, v := range listdata {
		node := &Rulestable{
			Id:        v.Id,
			Pid:       v.Pid,
			Type:      v.Type,
			Icon:      v.Icon,
			Pathname:  v.Pathname,
			Component: v.Component,
			Title:     v.Title,
			Ismenu:    v.Ismenu,
			Weigh:     v.Weigh,
			Status:    v.Status,
			Created:   v.Created,
		}
		tabledata = append(tabledata, node)
	}
	listnum := models.GetRulestotal(search)

	result["page"] = page
	result["totalnum"] = listnum
	result["limit"] = limit
	if listdata == nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "获取列表数据失败",
			"data":    "",
		})
		return
	} else {
		result["listdata"] = tabledata
		c.JSON(200, gin.H{
			"code":    200,
			"message": "数据获取成功",
			"data":    result,
		})
		return
	}
}

// 2. 获取所有权限规则树形结构 (适用于角色配置时的 ElTree / ElTreeSelect 下拉组件)
func GetRulesTree(c *gin.Context) {
	// 获取所有规则列表 (传入 limit=0, page=0 代表查询全量数据)
	listdata := models.GetRulesList(0, 0, nil, "weigh desc, id asc")
	if listdata == nil {
		c.JSON(200, gin.H{
			"code":    200,
			"message": "数据获取成功",
			"data":    []interface{}{},
		})
		return
	}

	// 第一步：构建全部节点的 Map 字典
	nodeMap := make(map[int64]*RulesTreeNode)
	for _, v := range listdata {
		nodeMap[v.Id] = &RulesTreeNode{
			Id:        v.Id,
			Pid:       v.Pid,
			Type:      v.Type,
			Icon:      v.Icon,
			Pathname:  v.Pathname,
			Component: v.Component,
			Title:     v.Title,
			Ismenu:    v.Ismenu,
			Weigh:     v.Weigh,
			Status:    v.Status,
			Created:   v.Created,
			Children:  make([]*RulesTreeNode, 0),
		}
	}

	// 第二步：根据 Pid 关联父子节点，生成树形结构
	var tree []*RulesTreeNode
	for _, v := range listdata {
		node := nodeMap[v.Id]
		if v.Pid == 0 {
			// Pid 为 0，说明是根节点
			tree = append(tree, node)
		} else {
			// 如果有父节点，将当前节点追加到父节点的 Children 数组中
			if parentNode, exists := nodeMap[v.Pid]; exists {
				parentNode.Children = append(parentNode.Children, node)
			} else {
				// 兜底逻辑：如果找不到匹配的父节点，直接作为顶级节点展示，避免数据丢失
				tree = append(tree, node)
			}
		}
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "数据获取成功",
		"data":    tree,
	})
}

// 3. 根据组别 Rules 权限集获取对应的动态菜单树 (用于系统左侧侧边栏/用户路由)
// 3. 根据组别 Rules 权限集获取对应的动态菜单树 (用于系统左侧侧边栏/用户路由)
// GetMenuByRules 根据组别 Rules 权限集获取对应的动态菜单树
func GetMenuByRules(c *gin.Context) {
	var req GetMenuByRulesReq

	// 1. 优先从 Query 获取，如果没拿到再尝试解析 JSON Body
	req.Rules = c.Query("rules")
	if req.Rules == "" && c.Request.ContentLength > 0 {
		_ = c.ShouldBindJSON(&req)
	}

	// 2. 传递 limit=0 或 10000 确保获取全量规则列表
	listdata := models.GetRulesList(10000, 1, &models.Authrule{}, "weigh desc, id asc")
	if len(listdata) == 0 {
		c.JSON(200, gin.H{
			"code":    200,
			"message": "菜单获取成功",
			"data":    []interface{}{},
		})
		return
	}

	// 3. 解析 rules 权限集合
	rulesStr := strings.TrimSpace(req.Rules)
	isSuper := rulesStr == "*" || rulesStr == ""
	ruleMap := make(map[int64]bool)

	if !isSuper {
		idStrs := strings.Split(rulesStr, ",")
		for _, idStr := range idStrs {
			trimmed := strings.TrimSpace(idStr)
			if trimmed == "" {
				continue
			}
			id, err := strconv.ParseInt(trimmed, 10, 64)
			if err == nil {
				ruleMap[id] = true
			}
		}
	}

	// 4. 构建全量 Node 字典 (注意：权限配置树不要过滤 Ismenu == 0，否则会导致目录/子节点丢失)
	nodeMap := make(map[int64]*RulesTreeNode)
	for _, v := range listdata {
		// 权限过滤：非超管时校验 ID 是否在规则中
		if !isSuper {
			if _, hasPermission := ruleMap[v.Id]; !hasPermission {
				continue
			}
		}

		// 注意：如果是在【用户角色组】分配权限时使用此接口，不能过滤 Ismenu == 0
		// 因为“目录”或“按钮”可能不是 Ismenu=1，过滤会导致下级节点找不到父节点而断裂。

		nodeMap[v.Id] = &RulesTreeNode{
			Id:        v.Id,
			Pid:       v.Pid,
			Type:      v.Type,
			Icon:      v.Icon,
			Pathname:  v.Pathname,
			Component: v.Component,
			Title:     v.Title,
			Ismenu:    v.Ismenu,
			Weigh:     v.Weigh,
			Status:    v.Status,
			Created:   v.Created,
			Children:  make([]*RulesTreeNode, 0),
		}
	}

	// 5. 按照原始有序数组 (listdata) 构建 Tree 结构，彻底解决 map 乱序导致的层级丢失问题
	var tree []*RulesTreeNode
	for _, v := range listdata {
		// 只处理已在 nodeMap 中的有效节点
		node, exists := nodeMap[v.Id]
		if !exists {
			continue
		}

		// 如果 Pid 为 0，或者父节点不在权限范围/字典中，作为顶级节点展示
		if node.Pid == 0 || nodeMap[node.Pid] == nil {
			tree = append(tree, node)
		} else {
			// 将子节点正确挂载到父节点的 Children 列表中
			parentNode := nodeMap[node.Pid]
			parentNode.Children = append(parentNode.Children, node)
		}
	}

	c.JSON(200, gin.H{
		"code":    200,
		"message": "菜单获取成功",
		"data":    tree,
	})
}

// 4. 删除权限规则
func DelRules(c *gin.Context) {
	var req struct {
		Id int64 `json:"id"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "参数错误",
			"data":    err.Error(),
		})
		return
	}
	if req.Id == 0 {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "缺少 ID",
			"data":    "",
		})
		return
	}

	outnum := models.DeleteRules(req.Id)
	c.JSON(200, gin.H{
		"code":    200,
		"message": "删除成功",
		"data":    outnum,
	})
}

// 5. 新增权限规则
func AddRules(c *gin.Context) {
	var rule models.Authrule
	if err := c.BindJSON(&rule); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "参数错误",
			"data":    err.Error(),
		})
		return
	}

	if err := models.AddRules(&rule); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "新增失败",
			"data":    err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"code":    200,
		"message": "新增成功",
		"data":    "",
	})
}

// 6. 修改权限规则
func EditRules(c *gin.Context) {
	var rule models.Authrule
	if err := c.BindJSON(&rule); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "参数错误",
			"data":    err.Error(),
		})
		return
	}
	if rule.Id == 0 {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "缺少 ID",
			"data":    "",
		})
		return
	}

	if err := models.EditRules(&rule); err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "修改失败",
			"data":    err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"code":    200,
		"message": "修改成功",
		"data":    "",
	})
}
