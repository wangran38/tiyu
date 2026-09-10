package controllers

import (
	// "fmt"
	"tiyu/models"
    "tiyu/utils"
	"github.com/gin-gonic/gin"
)

//获取当前用户信息
func GetAdminInfo(c *gin.Context) {
	//从header中获取到token
	token := c.Request.Header.Get("Authorization")
	if token != "" || len(token) != 0 {
    //   fmt.Println(token)
		user := utils.GetLoginUser(token)
		result := make(map[string]interface{})
		// result["user"] = user //返回当前总数

		//角色集合
		// role := models.Selectrules()

		// result["roles"] = role //返回当前总数
		result["id"] = user.Id
		result["username"] = user.Username
		result["nickname"] = user.Nickname

		result["avatar"] = user.Avatar
		// //菜单
		// menu := service.GetMenuPermission(user)
		c.JSON(200, gin.H{
			"code":    200,
			"message": "数据获取成功",
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

//获取当前用户 所有的权限控制
func GetAdminRule(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	if token != "" || len(token) != 0 {
    //   fmt.Println(token)
		user := utils.GetLoginUser(token)
		access, err := models.SelectAdminGid(user.Id)//查找组别
		if err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "失败",
			"data": err,
		})
		return
		}
		group, err := models.SelectGidRule(access.Gid)//查找组别菜单
		if err != nil {
		c.JSON(200, gin.H{
			"code":    201,
			"message": "获取菜单失败",
			"data": err,
		})
		return
		}
		//判断是否是超级用户
		if group.Rules== "*" {
			rule := models.Getruletree()
			if rule== nil {
			c.JSON(200, gin.H{
			"code":    201,
			"message": "获取菜单失败1",
			"data": "",
		})
		return
			} else {

		c.JSON(200, gin.H{
			"code":    200,
			"message": "数据获取成功1",
			"data": rule,
		})
		return
			}
		}
		ruleslist := models.Getruleadmintree(group.Rules)
		// for range
		// result := make([]map[string]interface{},0)
		// // result["user"] = user //返回当前总数
		// //角色集合
		// // role := models.Selectrules()

		// result[rules] = rules //返回当前总数
		// //菜单

		c.JSON(200, gin.H{
			"code":    200,
			"message": "数据获取成功2",
			"data": ruleslist,
		})
	} else {
		c.JSON(201, gin.H{
			"code":    201,
			"message": "你没有权限,或者权限已经过期！",
			"data":    "",
			// "permissions": menu,
			// "roles":       role,
		})
	}
}