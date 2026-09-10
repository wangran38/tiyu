package api

import (
	"fmt"
	"net/http"
	"strings"

	"tiyu/models"

	"github.com/gin-gonic/gin"
)

// CityGroup 字母分组结构体
type CityGroup struct {
	Letter string         `json:"letter"`
	List   []*models.City `json:"list"`
}

// CityTreeNode 三级联动树节点（JSON 字段与 models.City 保持一致）
type CityTreeNode struct {
	Id       int64           `json:"id"`                 // 对应城市ID
	Name     string          `json:"name"`               // 对应城市名称
	Code     string          `json:"code"`               // 行政区划代码
	Children []*CityTreeNode `json:"children,omitempty"` // 下级节点，若为空在 JSON 中自动忽略
}

// GetCityList 获取城市列表（包含热门城市与 A-Z 拼音首字母分组）
func GetCityList(c *gin.Context) {
	// 1. 查询热门城市
	hotCities, err := models.GetHotCities()
	if err != nil {
		hotCities = []*models.City{}
	}

	// 2. 查询地级市列表 (Level=1 代表城市)
	cityList, err := models.GetAllCitiesByLevel(1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取城市列表失败: " + err.Error(),
		})
		return
	}

	// 3. 按拼音首字母 (First 字段) 分组
	groupMap := make(map[string][]*models.City)
	for _, city := range cityList {
		letter := strings.ToUpper(strings.TrimSpace(city.First))
		if letter == "" {
			letter = "#"
		}
		groupMap[letter] = append(groupMap[letter], city)
	}

	// 4. 将 Map 转为按 A-Z 顺序的 Slice
	var groups []CityGroup
	for i := 'A'; i <= 'Z'; i++ {
		letter := string(i)
		if list, ok := groupMap[letter]; ok && len(list) > 0 {
			groups = append(groups, CityGroup{
				Letter: letter,
				List:   list,
			})
		}
	}
	if list, ok := groupMap["#"]; ok && len(list) > 0 {
		groups = append(groups, CityGroup{
			Letter: "#",
			List:   list,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "获取城市列表成功",
		"data": gin.H{
			"hot":    hotCities, // 热门城市
			"groups": groups,    // A-Z 分组城市
		},
	})
}

// GetCityTree 获取省市区三级联动树
func GetCityTree(c *gin.Context) {
	cities, err := models.GetAllAreaList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取地区树失败: " + err.Error(),
		})
		return
	}

	// 1. 将数据转成 Node Map (与 models.City 字段保持一致)
	nodeMap := make(map[int64]*CityTreeNode)
	for _, item := range cities {
		nodeMap[item.Id] = &CityTreeNode{
			Id:   item.Id,
			Name: item.Name,
			Code: item.Code,
		}
	}

	// 2. 构建树形层级结构
	var tree []*CityTreeNode
	for _, item := range cities {
		currentNode := nodeMap[item.Id]
		if item.Pid == 0 {
			tree = append(tree, currentNode)
		} else {
			if parentNode, exists := nodeMap[int64(item.Pid)]; exists {
				parentNode.Children = append(parentNode.Children, currentNode)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "获取省市区联动树成功",
		"data": tree,
	})
}

// GetCitiesByPid 根据父级 ID (pid) 获取下级地区列表
func GetCitiesByPid(c *gin.Context) {
	pidStr := c.DefaultQuery("pid", "0")
	var pid int64
	if _, err := fmt.Sscanf(pidStr, "%d", &pid); err != nil {
		pid = 0
	}

	cities, err := models.GetCitiesByPid(pid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取下级地区失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "获取下级地区成功",
		"data": cities,
	})
}
