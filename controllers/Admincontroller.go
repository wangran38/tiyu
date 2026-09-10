package controllers

import (
	"net/http"
	"time"
	"tiyu/lib"
	"tiyu/models"

	"github.com/gin-gonic/gin"
)

// 1. 列表搜索参数结构体
type Adminserch struct {
	Username string `json:"title"`
	Limit    int    `json:"limit"`
	Page     int    `json:"page"`
	Order    string `json:"sort"`
}

// 2. 新增用户请求结构体
type Adminform struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password"`                    // 允许为空，默认 pg123456
	GroupId  int64  `form:"group_id" json:"group_id" binding:"required"` // 必填组别ID
}

// 3. 编辑用户请求结构体
type UpdateAdminForm struct {
	Uid      int64  `json:"uid" form:"uid" binding:"required"`
	Username string `json:"username" form:"username"`
	GroupId  int64  `json:"group_id" form:"group_id"`
}

// 4. 单独修改组别请求结构体
type UpdateGroupForm struct {
	Uid     int64 `json:"uid" form:"uid" binding:"required"`
	GroupId int64 `json:"group_id" form:"group_id" binding:"required"`
}

// 5. 删除用户请求结构体
type DeleteAdminForm struct {
	Uid int64 `json:"uid" form:"uid" binding:"required"`
}

// 6. 一键还原密码请求结构体
type ResetPasswordForm struct {
	Uid int64 `json:"uid" form:"uid" binding:"required"`
}

// 7. 用户自主修改密码请求结构体
type ChangePasswordForm struct {
	Uid         int64  `json:"uid" form:"uid" binding:"required"`
	OldPassword string `json:"old_password" form:"old_password" binding:"required"`
	NewPassword string `json:"new_password" form:"new_password" binding:"required"`
}

// ================================= 控制器处理函数 =================================

// AddAdmin 添加用户 (同时关联组别)
func AddAdmin(c *gin.Context) {
	var admindata Adminform
	if err := c.ShouldBind(&admindata); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 201,
			"msg":  "表单未完整，用户名和组别不能为空！",
		})
		return
	}

	// 1. 判断账号是否存在
	Admin := new(models.Admin)
	Admin.Username = admindata.Username
	info, _ := models.SelectUserByUserName(Admin.Username)
	if info != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 201,
			"msg":  "该用户已经存在！",
		})
		return
	}

	// 2. 初始密码逻辑：如果未传密码，默认设置 pg123456
	rawPassword := admindata.Password
	if rawPassword == "" {
		rawPassword = "pg123456"
	}

	// 3. 生成密码和盐
	pwd, salt := lib.Password(4, rawPassword)
	Admin.Password = pwd
	Admin.Salt = salt
	Admin.Created = time.Now()

	// 4. 插入 admin 表
	err := models.AddAdmin(Admin)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 201,
			"msg":  "添加数据出错：" + err.Error(),
		})
		return
	}

	// 5. 绑定用户与组别关系到 auth_group_access 表
	access := &models.Authaccess{
		Uid: Admin.Id,
		Gid: admindata.GroupId,
	}

	err = models.AddAuthAccess(access)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 201,
			"msg":  "账号创建成功，但绑定组别失败：" + err.Error(),
		})
		return
	}

	// 6. 返回结果
	result := make(map[string]interface{})
	result["id"] = Admin.Id
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "数据添加成功！",
		"data": result,
	})
}

// GetAdminlist 获取管理员用户列表
func GetAdminlist(c *gin.Context) {
	var searchdata Adminserch
	c.BindJSON(&searchdata)

	result := make(map[string]interface{})
	limit := searchdata.Limit
	page := searchdata.Page
	username := searchdata.Username
	order := searchdata.Order

	listdata := models.GetUserList(limit, page, username, order)
	listnum := models.GetUsertotal(username)

	result["page"] = page
	result["totalnum"] = listnum
	result["limit"] = limit

	if listdata == nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 201,
			"msg":  "获取管理员列表失败！",
			"data": "",
		})
		return
	}

	result["listdata"] = listdata
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "数据获取成功！",
		"data": result,
	})
}

// UpdateAdmin 编辑用户（同时处理修改用户名和选择组别）
func UpdateAdmin(c *gin.Context) {
	var form UpdateAdminForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 201,
			"msg":  "参数不完整，缺少必要的 UID！",
		})
		return
	}

	// 1. 校验目标用户是否存在
	targetUser, err := models.SelectUserById(form.Uid)
	if err != nil || targetUser == nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 201,
			"msg":  "修改的用户不存在！",
		})
		return
	}

	// 2. 如果修改了用户名，校验与其他账号是否冲突
	if form.Username != "" && form.Username != targetUser.Username {
		existUser, _ := models.SelectUserByUserName(form.Username)
		if existUser != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": 201,
				"msg":  "该用户名已被其他账号占用！",
			})
			return
		}
	}

	// 3. 调用 Model 层的事务更新（同时更新 admin 表和 auth_group_access 表）
	err = models.UpdateAdminWithGroup(form.Uid, form.Username, form.GroupId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 201,
			"msg":  "编辑保存失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "用户信息及组别更新成功！",
	})
}

// UpdateUserGroup 单独修改组别
func UpdateUserGroup(c *gin.Context) {
	var form UpdateGroupForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 201,
			"msg":  "参数不完整！",
		})
		return
	}

	err := models.UpdateAdminGroup(form.Uid, form.GroupId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 201,
			"msg":  "修改组别失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "修改用户组别成功！",
	})
}

// DeleteAdmin 删除用户 (带事务处理：同时清除用户数据和组别关联数据)
func DeleteAdmin(c *gin.Context) {
	var form DeleteAdminForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 201,
			"msg":  "参数不完整，缺少用户 ID！",
		})
		return
	}

	// 防呆校验：保护默认超级管理员（id为1）不被删除
	if form.Uid == 1 {
		c.JSON(http.StatusOK, gin.H{
			"code": 201,
			"msg":  "超级管理员账号不允许删除！",
		})
		return
	}

	err := models.DeleteAdminWithAccess(form.Uid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 201,
			"msg":  "删除用户失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "用户及对应组别关系删除成功！",
	})
}

// ResetPassword 一键还原密码为默认值 (pg123456)
func ResetPassword(c *gin.Context) {
	var form ResetPasswordForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 201,
			"msg":  "参数不完整！",
		})
		return
	}

	pwd, salt := lib.Password(4, "pg123456")
	err := models.UpdateAdminPassword(form.Uid, pwd, salt)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 201,
			"msg":  "密码重置失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "密码已重置为初始密码：pg123456",
	})
}

// ChangePassword 用户修改密码
func ChangePassword(c *gin.Context) {
	var form ChangePasswordForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 201,
			"msg":  "参数不完整！",
		})
		return
	}

	user, err := models.SelectUserById(form.Uid)
	if err != nil || user == nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 201,
			"msg":  "未找到该用户！",
		})
		return
	}

	// 校验旧密码是否正确
	oldPwdCalculated, _ := lib.Password(4, form.OldPassword)
	if user.Password != oldPwdCalculated {
		c.JSON(http.StatusOK, gin.H{
			"code": 201,
			"msg":  "原密码输入错误！",
		})
		return
	}

	// 生成新密码加密串并写入数据库
	newPwd, newSalt := lib.Password(4, form.NewPassword)
	err = models.UpdateAdminPassword(form.Uid, newPwd, newSalt)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 201,
			"msg":  "修改密码失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "密码修改成功！",
	})
}
