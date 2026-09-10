package api

import (
	"fmt"
	"net/http"
	"strconv"

	"tiyu/models"

	"github.com/gin-gonic/gin"
)

// ShopCategoryTreeNode 分类树节点结构体
type ShopCategoryTreeNode struct {
	ID          uint64                  `json:"id"`
	ParentID    int64                   `json:"parent_id"`
	Name        string                  `json:"name"`
	Code        string                  `json:"code"`
	Icon        string                  `json:"icon"`
	Sort        int32                   `json:"sort"`
	Status      int8                    `json:"status"`
	Description string                  `json:"description"`
	Children    []*ShopCategoryTreeNode `json:"children,omitempty"` // 子分类列表（无子节点时隐藏）
}

// GetShopCategories 获取商家分类列表（支持分页、按父级/状态/名称筛选）
func GetShopCategories(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	parentIDStr := c.DefaultQuery("parent_id", "-1") // 默认 -1 查全部分类
	parentID, _ := strconv.ParseInt(parentIDStr, 10, 64)
	name := c.Query("name")
	code := c.Query("code")
	statusStr := c.DefaultQuery("status", "1") // C端API默认仅查启用状态 (1)
	status, _ := strconv.Atoi(statusStr)
	order := c.DefaultQuery("order", "sort")

	search := &models.ShopCategory{
		ParentID: parentID,
		Name:     name,
		Code:     code,
		Status:   int8(status),
	}

	list, err := models.GetShopCategoryList(limit, page, search, order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取商家分类列表失败: " + err.Error(),
		})
		return
	}

	total := models.GetShopCategoryTotal(search)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "获取商家分类列表成功",
		"data": gin.H{
			"list":  list,
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// GetShopCategoryTree 获取商家分类树状结构
func GetShopCategoryTree(c *gin.Context) {
	categories, err := models.GetAllActiveShopCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取商家分类树失败: " + err.Error(),
		})
		return
	}

	// 1. 将模型数据映射为树节点 Node Map
	nodeMap := make(map[uint64]*ShopCategoryTreeNode)
	for _, item := range categories {
		nodeMap[item.ID] = &ShopCategoryTreeNode{
			ID:          item.ID,
			ParentID:    item.ParentID,
			Name:        item.Name,
			Code:        item.Code,
			Icon:        item.Icon,
			Sort:        item.Sort,
			Status:      item.Status,
			Description: item.Description,
		}
	}

	// 2. 挂载节点形成树状层级
	var tree []*ShopCategoryTreeNode
	for _, item := range categories {
		currentNode := nodeMap[item.ID]
		// ParentID == 0 代表顶级分类
		if item.ParentID == 0 {
			tree = append(tree, currentNode)
		} else {
			if parentNode, exists := nodeMap[uint64(item.ParentID)]; exists {
				parentNode.Children = append(parentNode.Children, currentNode)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "获取商家分类树成功",
		"data": tree,
	})
}

// GetShopCategoryByPid 根据 ParentID 获取下一级分类列表
func GetShopCategoryByPid(c *gin.Context) {
	parentIDStr := c.DefaultQuery("parent_id", "0") // 默认查顶级 (parent_id=0)
	var parentID int64
	if _, err := fmt.Sscanf(parentIDStr, "%d", &parentID); err != nil {
		parentID = 0
	}

	search := &models.ShopCategory{
		ParentID: parentID,
		Status:   1, // 默认仅查启用
	}

	// 默认返回前100条
	list, err := models.GetShopCategoryList(100, 1, search, "sort")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取子分类失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "获取子分类成功",
		"data": list,
	})
}
