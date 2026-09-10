package controllers

import (
	"time"
	"tiyu/lib"
	"tiyu/models"

	"github.com/gin-gonic/gin"
)

type Usersearch struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Mobile   string `json:"mobile"`
	Status   string `json:"status"`
	Limit    int    `json:"limit"`
	Page     int    `json:"page"`
	Order    string `json:"sort"`
}

// UserInput 新增/修改会员的入参 DTO（绕过 User.Password 的 json:"-" 限制）
type UserInput struct {
	Id              int64  `json:"id"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	Nickname        string `json:"nickname"`
	Avatar          string `json:"avatar"`
	Realname        string `json:"realname"`
	Mobile          string `json:"mobile"`
	Email           string `json:"email"`
	Gender          int    `json:"gender"`
	Birthday        string `json:"birthday"`
	Idcard          string `json:"idcard"`
	Address         string `json:"address"`
	FavoriteTeams   string `json:"favorite_teams"`
	FavoritePlayers string `json:"favorite_players"`
	Point           int    `json:"point"`
	Level           int    `json:"level"`
	Status          string `json:"status"`
}

// 会员列表
func GetUserlist(c *gin.Context) {
	var s Usersearch
	c.BindJSON(&s)

	search := &models.User{
		Id:       s.Id,
		Username: s.Username,
		Nickname: s.Nickname,
		Mobile:   s.Mobile,
		Status:   s.Status,
	}

	listdata := models.GetMemberList(s.Limit, s.Page, search, s.Order)
	listnum := models.GetMemberTotal(search)

	result := make(map[string]interface{})
	result["page"] = s.Page
	result["totalnum"] = listnum
	result["limit"] = s.Limit

	if listdata == nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "获取会员列表失败",
			"data":    "",
		})
		return
	}

	result["listdata"] = listdata
	c.JSON(200, gin.H{
		"code":    200,
		"message": "数据获取成功",
		"data":    result,
	})
}

// 新增会员
func AddUser(c *gin.Context) {
	var in UserInput
	if err := c.BindJSON(&in); err != nil {
		c.JSON(200, gin.H{"code": 201, "message": "参数错误", "data": err.Error()})
		return
	}

	// 用户名重名校验
	if existed, _ := models.SelectMemberByUsername(in.Username); existed != nil {
		c.JSON(200, gin.H{"code": 201, "message": "用户名已存在", "data": ""})
		return
	}

	// 生成密码盐并加密
	salt := lib.GetRandomString(6)
	pwd := lib.Md5([]byte(in.Password + "988cj.com" + salt))

	u := models.User{
		Username: in.Username,
		Password: pwd,
		Salt:     salt,
		Nickname: in.Nickname,
		Avatar:   in.Avatar,
		Realname: in.Realname,
		Mobile:   in.Mobile,
		Email:    in.Email,
		Gender:   in.Gender,
		Birthday: in.Birthday,

		Point:  in.Point,
		Level:  in.Level,
		Status: in.Status,
	}
	if u.Status == "" {
		u.Status = "normal"
	}
	if u.Level == 0 {
		u.Level = 1
	}

	err := models.AddMember(&u)
	if err != nil {
		c.JSON(200, gin.H{"code": 201, "message": "新增失败", "data": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 200, "message": "新增成功", "data": ""})
}

// 修改会员
func EditUser(c *gin.Context) {
	var in UserInput
	if err := c.BindJSON(&in); err != nil {
		c.JSON(200, gin.H{"code": 201, "message": "参数错误", "data": err.Error()})
		return
	}
	if in.Id == 0 {
		c.JSON(200, gin.H{"code": 201, "message": "缺少 ID", "data": ""})
		return
	}

	u := models.User{
		Id:       in.Id,
		Nickname: in.Nickname,
		Avatar:   in.Avatar,
		Realname: in.Realname,
		Mobile:   in.Mobile,
		Email:    in.Email,
		Gender:   in.Gender,
		Birthday: in.Birthday,
		Point:    in.Point,
		Level:    in.Level,
		Status:   in.Status,
	}

	// 若提交了明文密码则重新加密（长度 < 64 视为明文）
	if in.Password != "" && len(in.Password) < 64 {
		salt := lib.GetRandomString(6)
		u.Salt = salt
		u.Password = lib.Md5([]byte(in.Password + "988cj.com" + salt))
	}

	// xorm 默认不更新零值字段，password 为空时自动跳过（不会清空原密码）
	err := models.EditMember(&u)
	if err != nil {
		c.JSON(200, gin.H{"code": 201, "message": "修改失败", "data": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 200, "message": "修改成功", "data": ""})
}

// 删除会员
func DelUser(c *gin.Context) {
	var req struct {
		Id int64 `json:"id"`
	}
	c.BindJSON(&req)
	if req.Id == 0 {
		c.JSON(200, gin.H{"code": 201, "message": "缺少 ID", "data": ""})
		return
	}
	outnum := models.DelMember(req.Id)
	c.JSON(200, gin.H{"code": 200, "message": "删除成功", "data": outnum})
}

// 会员登录（前台 API）
func UserLogin(c *gin.Context) {
	var req struct {
		Username string `form:"username" json:"username" binding:"required"`
		Password string `form:"password" json:"password" binding:"required"`
	}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(200, gin.H{"code": 201, "message": "参数不完整"})
		return
	}

	u, _ := models.SelectMemberByUsername(req.Username)
	if u == nil {
		c.JSON(200, gin.H{"code": 201, "message": "用户不存在"})
		return
	}

	pwd := lib.Md5([]byte(req.Password + "988cj.com" + u.Salt))
	if pwd != u.Password {
		c.JSON(200, gin.H{"code": 201, "message": "密码不正确"})
		return
	}

	if u.Status == "banned" {
		c.JSON(200, gin.H{"code": 201, "message": "账号已被禁用，请联系管理员"})
		return
	}

	// 更新最后登录信息
	u.LastLoginIp = c.ClientIP()
	u.LastLoginAt = int64(time.Now().Unix())
	_ = models.EditMember(u)

	c.JSON(200, gin.H{
		"code":    200,
		"message": "登录成功",
		"data": gin.H{
			"id":       u.Id,
			"username": u.Username,
			"nickname": u.Nickname,
			"avatar":   u.Avatar,
			"point":    u.Point,
			"level":    u.Level,
		},
	})
}
