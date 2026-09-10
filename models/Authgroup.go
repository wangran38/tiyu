package models

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// Authgroup 角色组表模型
type Authgroup struct {
	Id         int64  `json:"id" xorm:"pk autoincr 'id'"`
	Pid        int64  `json:"pid" xorm:"'pid'"`
	Name       string `json:"name" xorm:"'name'"`
	Rules      string `json:"rules" xorm:"'rules'"`
	Created    int64  `json:"created" xorm:"'created'"`
	Updated    int64  `json:"updated" xorm:"'updated'"`
	Status     int    `json:"status" xorm:"'status'"`
	Createtime int64  `json:"createtime" xorm:"'createtime'"`
	Updatetime int64  `json:"updatetime" xorm:"'updatetime'"`
}

func (a *Authgroup) TableName() string {
	return "auth_group"
}

// AuthGroupResponse 用于组别列表返回的结构体（带有解析后的菜单名称数组）
type AuthGroupResponse struct {
	Authgroup          // 嵌套基本字段
	RuleNames []string `json:"rule_names"` // 存放解析后的菜单名称数组
}

// ======================= 组别树形结构定义 =======================

// AuthGroupTreeResponse 用于前端级联/下拉选择框的树形结构体
type AuthGroupTreeResponse struct {
	Authgroup
	Children []*AuthGroupTreeResponse `json:"children,omitempty"` // 递归嵌套子节点
}

// GetGroupTree 专门给菜单/权限树选择时调用的接口（无需分页，返回 Tree 树状数据）
// GetGroupTree 专门给菜单/权限树选择时调用的接口（无需分页，返回 Tree 树状数据）
func GetGroupTree() ([]*AuthGroupTreeResponse, error) {
	var list []*Authgroup

	// 修复1：去掉 status = 1 的强行限制（或者包含 status = 0），保证查出已有组别
	// 如果你后期有软删除或禁用状态，可以去掉 Where 条件，或者根据实际情况查询
	err := Dorm.OrderBy("id ASC").Find(&list)
	if err != nil {
		return nil, err
	}

	// 从顶级节点 (pid = 0) 开始构建树
	return buildTree(list, 0), nil
}

// buildTree 递归构建树状结构的辅助函数
// 修复2：将 pid 参数类型从 int 修改为 int64，与 Authgroup.Pid 统一
func buildTree(list []*Authgroup, pid int64) []*AuthGroupTreeResponse {
	tree := make([]*AuthGroupTreeResponse, 0)

	for _, item := range list {
		if item.Pid == pid {
			node := &AuthGroupTreeResponse{
				Authgroup: *item,
			}
			// 递归寻找当前节点的子节点
			children := buildTree(list, item.Id)
			if len(children) > 0 {
				node.Children = children
			}
			tree = append(tree, node)
		}
	}

	return tree
}

// ======================= CRUD 方法 =======================

// SelectGidRule 根据组别 ID 获取组别和 Rules 权限
func SelectGidRule(id int64) (*Authgroup, error) {
	a := new(Authgroup)
	has, err := Dorm.Where("id = ?", id).Get(a)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("组别数据不存在！")
	}
	return a, nil
}

// Addgroup 添加组别
func Addgroup(a *Authgroup) error {
	a.Created = time.Now().Unix()
	a.Updated = time.Now().Unix()
	_, err := Dorm.Insert(a)
	return err
}

// Editgroup 编辑修改组别信息（新增）
func Editgroup(a *Authgroup) error {
	a.Updated = time.Now().Unix()
	// 显式指定需要更新的列，防止零值被忽略
	_, err := Dorm.ID(a.Id).Cols("pid", "name", "rules", "status", "updatetime").Update(a)
	return err
}

// Getgrouptotal 获取组别总数
func Getgrouptotal(search string) int64 {
	a := new(Authgroup)
	session := Dorm.NewSession()
	defer session.Close()

	if search != "" {
		session.Where("name LIKE ?", "%"+search+"%")
	}

	total, err := session.Count(a)
	if err != nil {
		return 0
	}
	return total
}

// GetgroupList 列表查询方法，带 RuleNames 中文解析
// limit: 每页条数, page: 当前页码(从 1 开始)
func GetgroupList(limit int, page int, search string, order string) ([]*AuthGroupResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	// 1. 查出原始的组别列表
	listdata := make([]*Authgroup, 0)
	byorder := "id ASC"
	if order == "-id" {
		byorder = "id DESC"
	}

	session := Dorm.NewSession()
	defer session.Close()

	if search != "" {
		session.Where("name LIKE ?", "%"+search+"%")
	}

	err := session.OrderBy(byorder).Limit(limit, offset).Find(&listdata)
	if err != nil {
		return nil, err
	}

	// 2. 收集当前页所有组别用到的菜单 ID 去重
	ruleIdSet := make(map[string]struct{})
	for _, group := range listdata {
		if group.Rules != "" {
			ids := strings.Split(group.Rules, ",")
			for _, idStr := range ids {
				idStr = strings.TrimSpace(idStr)
				if idStr != "" {
					ruleIdSet[idStr] = struct{}{}
				}
			}
		}
	}

	// 3. 批量查询关联的菜单数据，并建立 ID -> Title 的映射
	ruleTitleMap := make(map[string]string)
	if len(ruleIdSet) > 0 {
		ids := make([]string, 0, len(ruleIdSet))
		for idStr := range ruleIdSet {
			ids = append(ids, idStr)
		}

		var rules []Authrule
		// 批量查询所有用到的权限菜单 title
		if err := Dorm.In("id", ids).Cols("id", "title").Find(&rules); err == nil {
			for _, r := range rules {
				idStr := strconv.FormatInt(r.Id, 10)
				ruleTitleMap[idStr] = r.Title
			}
		}
	}

	// 4. 组装最终返回的数据结构
	responseList := make([]*AuthGroupResponse, 0, len(listdata))
	for _, group := range listdata {
		item := &AuthGroupResponse{
			Authgroup: *group,
			RuleNames: make([]string, 0),
		}

		if group.Rules != "" {
			ids := strings.Split(group.Rules, ",")
			for _, idStr := range ids {
				idStr = strings.TrimSpace(idStr)
				if title, ok := ruleTitleMap[idStr]; ok {
					item.RuleNames = append(item.RuleNames, title)
				}
			}
		}
		responseList = append(responseList, item)
	}

	return responseList, nil
}

// ======================= 数据库更新与删除函数 =======================

// UpdateAdminGroup 修改或新增用户的组别关联 (对应 auth_group_access 表)
func UpdateAdminGroup(uid int64, gid int64) error {
	access := new(Authaccess)
	has, err := Dorm.Where("uid = ?", uid).Get(access)
	if err != nil {
		return err
	}

	if has {
		// 存在记录，执行更新
		access.Gid = gid
		_, err = Dorm.Where("uid = ?", uid).Cols("gid").Update(access)
	} else {
		// 不存在记录，执行新增
		access.Uid = uid
		access.Gid = gid
		_, err = Dorm.Insert(access)
	}
	return err
}

// UpdateAdminPassword 更新指定用户的密码和盐
func UpdateAdminPassword(uid int64, password string, salt string) error {
	admin := &Admin{
		Password: password,
		Salt:     salt,
		Updated:  time.Now(),
	}
	_, err := Dorm.Where("id = ?", uid).Cols("password", "salt", "updated").Update(admin)
	return err
}

// Delgroup 根据 ID 删除 Authgroup 记录，并返回受影响行数及错误信息
func Delgroup(id int64) (int, error) {
	a := new(Authgroup)
	outnum, err := Dorm.ID(id).Delete(a)
	if err != nil {
		return 0, err
	}
	return int(outnum), nil
}
