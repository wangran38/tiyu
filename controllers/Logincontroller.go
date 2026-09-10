package controllers

import (

	// "hanyun-admin/src/service"
	// "hanyun-admin/src/utils"
	"tiyu/lib"
	"tiyu/models"
	"tiyu/utils"
	"net/http"
	"time"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LoginForm struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

//登录
func LoginController(c *gin.Context) {

	// dtime := time.Date(2020, 8, 25, 0, 0, 0, 0, time.Local).Unix()
	dtime := time.Now().Unix()
	reqIP := c.ClientIP()
	//定义结构体接收数据
	var logindata LoginForm
	if err := c.ShouldBind(&logindata); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    "201",
			"message": "表单未完整！",
			"msgcode": "-1",
			"data":    reqIP,
		})
		return
	}
	// //读取数据库
	admindata, err := models.SelectUserByUserName(logindata.Username) //判断账号是否存在！
	if admindata != nil {                                             //如果账号存在
		pwd := lib.Md5([]byte(logindata.Password + "988cj.com" + admindata.Salt))
		if pwd != admindata.Password {
			c.JSON(http.StatusOK, gin.H{
				"code":    "201",
				"message": "密码不正确！",
				// "pwd":  pwd,
			})
			return

		}
		adminid:=strconv.FormatInt(admindata.Id,10) //转换字符串
		token := utils.CreateJsonWebToken(adminid)
		result := make(map[string]interface{})
		result["token"] = token
		c.JSON(200, gin.H{
			"code":    200,
			"message": "登录成功！",
			// "token":   token,
			"data": result,
		})

	} else { //如果账号不存在

		// pws, _ := lib.Password(4, "")
		c.JSON(http.StatusOK, gin.H{
			"code":    "201",
			"message": "对不起，该用户无效！",
			"data":    err,
			"dtime":   dtime,
			"reqip":   reqIP,
		})

	}

	// fmt.Println(logindata.Username)
	//

}

//获取当前用户信息
func GetLoginAdminInfo(c *gin.Context) {
	//从header中获取到token
	token := c.Request.Header.Get("Authorization")
	if token != "" || len(token) != 0 {
		user := utils.GetLoginUser(token)
		result := make(map[string]interface{})
		// result["user"] = user //返回当前总数

		//角色集合
		// role := models.Selectrules()

		// result["roles"] = role //返回当前总数
		result["name"] = user.Username
		result["avatar"] = user.Avatar
		// //菜单
		// menu := service.GetMenuPermission(user)
		c.JSON(200, gin.H{
			"code":    200,
			"message": "操作成功",

			"data": result,
		})
	} else {
		c.JSON(201, gin.H{
			"code":    201,
			"message": "你没有权限！",
			"data":    "",
			// "permissions": menu,
			// "roles":       role,
		})
	}
}
//退出登录
func Loginout(c *gin.Context) {
		//从header中获取到token
	token := c.Request.Header.Get("Authorization")
	if token != "" || len(token) != 0 {
	utils.DelRedisKey(token)
			c.JSON(201, gin.H{
			"code":    200,
			"message": "退出成功！",
			"data":    "",
			// "permissions": menu,
			// "roles":       role,
		})
	} else {
			c.JSON(201, gin.H{
			"code":    201,
			"message": "你没有权限！",
			"data":    "",
		})
	}

}