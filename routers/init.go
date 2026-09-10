package routers

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
	"tiyu/api"
	"tiyu/controllers"
	_ "tiyu/models"
	"tiyu/utils"

	"github.com/gin-gonic/gin"
)

// IPVisitor 用于记录每个 IP 的请求频次
type IPVisitor struct {
	Count      int
	LastActive time.Time
}

var (
	visitors = make(map[string]*IPVisitor)
	mu       sync.Mutex
)

// 自动清理过期 IP 记录（防止内存泄漏）
func init() {
	go func() {
		for {
			time.Sleep(3 * time.Minute)
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.LastActive) > 3*time.Minute {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()
}

// RateLimitMiddleware 防刷限流中间件 (单 IP 在 window 时间内最多允许 maxRequests 次请求)
func RateLimitMiddleware(maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		v, exists := visitors[ip]
		if !exists {
			visitors[ip] = &IPVisitor{
				Count:      1,
				LastActive: time.Now(),
			}
			mu.Unlock()
			c.Next()
			return
		}

		// 如果超出了时间窗口，重置计数
		if time.Since(v.LastActive) > window {
			v.Count = 1
			v.LastActive = time.Now()
			mu.Unlock()
			c.Next()
			return
		}

		// 在时间窗口内，递增请求次数
		v.Count++
		v.LastActive = time.Now()

		if v.Count > maxRequests {
			mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code": 429,
				"msg":  "请求过于频繁，疑似机器人操作，请稍后再试",
			})
			c.Abort()
			return
		}

		mu.Unlock()
		c.Next()
	}
}

func init() {
	router := gin.Default()
	router.Use(Cors())

	// ==========================================
	// 1. 公开接口：无需 Token (增加防刷限流限制)
	// ==========================================
	adminPublic := router.Group("/admin")
	// 限制同一 IP 在 1 分钟内最多请求 10 次登录，防止密码爆破与脚本刷接口
	adminPublic.Use(RateLimitMiddleware(10, 1*time.Minute))
	{
		adminPublic.POST("/login", controllers.LoginController) // 管理员登录
		adminPublic.POST("/userlogin", controllers.UserLogin)   // 前台会员登录
	}

	// ==========================================
	// 2. 跨系统 API 路由组 (/api)
	// ==========================================
	// 将变量名 api 改为 apiGroup
	apiGroup := router.Group("/api")
	{
		// ================= 新增 Auth / 短信相关路由 =================
		// 1. 发送短信验证码接口 (推荐挂载防刷中间件，如 1 分钟最多 3 次，防止被刷短信)
		apiGroup.POST("/send-sms", RateLimitMiddleware(3, 1*time.Minute), api.SendLoginSms)

		// 2. 手机号 + 验证码 快捷登录/自动注册接口
		apiGroup.POST("/quick-login", api.QuickLoginByPhone)
		apiGroup.GET("/test-qwen-ocr", api.TestQwenOCR)
		apiGroup.POST("/public/upload-images", api.PublicUploadImagesHandler)
		apiGroup.POST("/public/recognize-ticket", api.RecognizeTicketByURLHandler)
		apiGroup.POST("/upload-qwen-ocr", api.UploadQwenOCR)
		apiGroup.POST("/adjustuserlevel", controllers.AdjustUserLevel)

		// 跨系统单点登录（SSO 免登录）：挂载防刷中间件 (1分钟最多5次)
		apiGroup.POST("/sso-login", RateLimitMiddleware(5, 1*time.Minute), controllers.SSOLogin)

		// 模拟返回用户信息的接口 (支持 GET 和 POST)
		apiGroup.GET("/mock-user-info", api.MockUserInfo)
		apiGroup.POST("/mock-user-info", api.MockUserInfo)

		// 城市列表（热门 + A-Z分组）
		apiGroup.GET("/cities", api.GetCityList)

		// ------------------------------------------
		// 新增：省市区三级联动接口
		// ------------------------------------------
		// 1. 一次性获取全量树状结构 (推荐前端 Cascader 级联选择器使用)
		apiGroup.GET("/city-tree", api.GetCityTree)

		// 2. 按父级 ID (pid) 动态懒加载下级列表
		apiGroup.GET("/cities-by-pid", api.GetCitiesByPid)

		// 商家分类接口
		// ------------------------------------------
		// 1. 获取商家分类列表（支持分页/搜索）
		apiGroup.GET("/shop-categories", api.GetShopCategories)

		// 2. 获取分类树结构（顶级分类 + 子分类层级）
		apiGroup.GET("/shop-categories/tree", api.GetShopCategoryTree)

		// 3. 根据父级 ID (parent_id) 懒加载获取子分类列表
		apiGroup.GET("/shop-categories/by-pid", api.GetShopCategoryByPid)

		// 前台商家列表：支持 category_id、city_id、user_lng、user_lat、radius_km、order
		apiGroup.GET("/shops", api.GetShopList)
		apiGroup.POST("/shops", api.GetShopList)
		// 前台指定店铺的有效优惠券列表
		apiGroup.POST("/shop/coupons", api.GetCouponListByShopIDHandler)
		// 新增：受 pgtoken 保护的前台用户中心路由组 (/api/user)
		// 👇 新增：AI票根核验与权益发放接口
		apiGroup.POST("/ticket/verify", utils.UserJWTAuth(), api.VerifyTicketHandler)
		// ------------------------------------------
		userGroup := apiGroup.Group("/user")
		userGroup.Use(utils.UserJWTAuth()) // 👈 使用你 utils 里写的中间件
		{
			// 示例接口：获取个人信息
			userGroup.GET("/profile", api.GetUserProfile)
			// 会员“我的票根”列表，数据来自 MySQL
			userGroup.POST("/ticket/list", api.GetMyMemberTicketListHandler)
			// 使用会员票根兑换优惠券
			userGroup.POST("/ticket/redeem-coupon", api.RedeemMemberTicketHandler)
			// 会员进入商户详情后提交订单
			userGroup.POST("/order/create", api.CreateOrderHandler)
			// 会员“我的订单”列表
			userGroup.POST("/order/list", api.GetMyOrderListHandler)

			// 示例接口：修改个人资料
			// userGroup.POST("/profile/update", api.UpdateUserProfile)
			// 👇 商家申请相关 API (统一使用 POST)
			userGroup.POST("/shop-apply", api.SubmitShopApplication)           // 提交/重新提交商家入驻申请
			userGroup.POST("/shop-apply/status", api.GetShopApplicationStatus) // 查询当前用户商家申请/店铺状态
			// 👇 新增：获取与编辑补充店铺资料接口
			userGroup.POST("/shop/info", api.GetShopInfo) // 根据店铺ID获取商家详细信息（用于回显）
			userGroup.POST("/shop/edit", api.EditShop)    //追加商家补充资料信息
			// 👇 新增：图片上传路由（支持单图/多图批量上传）
			userGroup.POST("/common/upload", api.UploadImagesHandler)
			// 👇 商家优惠券
			userGroup.POST("/shop/coupon/list", api.GetCouponListHandler)
			userGroup.POST("/shop/coupon/add", api.AddCouponHandler)
			userGroup.POST("/shop/coupon/edit", api.EditCouponHandler) // 新增编辑优惠券接口
			// 👇 新增：AI票根核验与权益发放接口

		}
	}

	// ==========================================
	// 3. 管理后台受保护接口：统一走 JWT 鉴权中间件
	// ==========================================
	admin := router.Group("/admin")
	admin.Use(utils.JWTAuth())
	{
		// ================= 原有路由 =================
		admin.POST("/logout", controllers.Loginout)           // 退出登录
		admin.POST("/add", controllers.AddAdmin)              // 添加用户 (包含组别关系)
		admin.POST("/getinfo", controllers.GetAdminInfo)      // 获取当前登录管理员信息
		admin.POST("/getrule", controllers.GetAdminRule)      // 获取权限菜单列表
		admin.POST("/getadminlist", controllers.GetAdminlist) // 获取管理员列表
		admin.POST("/edit", controllers.UpdateAdmin)          // 编辑用户 (同时修改账号名称与所属组别)
		admin.POST("/delete", controllers.DeleteAdmin)        // 删除用户 (带事务处理，连带清除组别关联)
		admin.POST("/resetpwd", controllers.ResetPassword)    // 一键还原密码为默认值 (pg123456)
		// ================= 新增/补全路由 =================

		admin.POST("/changepwd", controllers.ChangePassword)    // 用户修改个人密码
		admin.POST("/updategroup", controllers.UpdateUserGroup) // 仅修改用户组别 (可选)
		admin.POST("/getgrouplist", controllers.Getgrouplist)   // 获取角色组列表
		admin.POST("/getgrouptree", controllers.Getgrouptree)   // 获取角色组树状选择列表 (供下拉框/选择器使用)
		admin.POST("/addgroup", controllers.Addgroup)           // 新增角色组
		admin.POST("/editgroup", controllers.Editgroup)         // 修改/编辑角色组
		admin.POST("/delgroup", controllers.Delgroup)           // 删除角色组

		// 权限规则模块
		admin.POST("/getruleslist", controllers.Getruleslist)     // 规则表格分页列表
		admin.POST("/rulestree", controllers.GetRulesTree)        // 下拉/角色分配时的全量规则树
		admin.POST("/getmenubyrules", controllers.GetMenuByRules) // 👈 客户端登录后获取的可视菜单树
		admin.POST("/addrules", controllers.AddRules)             // 新增规则
		admin.POST("/editrules", controllers.EditRules)           // 编辑规则
		admin.POST("/delrules", controllers.DelRules)             // 删除规则
		// 短信路由配置
		admin.POST("sms/list", controllers.GetSmsConfigList)
		admin.POST("sms/add", controllers.AddSmsConfig)
		admin.POST("sms/edit", controllers.EditSmsConfig)
		admin.POST("sms/del", controllers.DelSmsConfig)

		admin.POST("/getcitylist", controllers.Getcitylist)
		admin.POST("/addcity", controllers.Addcity)
		admin.POST("/editcity", controllers.Editcity)
		admin.POST("/delcity", controllers.Delcity)
		admin.POST("/getsportscategorylist", controllers.GetSportsCategorylist)
		admin.POST("/addsportscategory", controllers.AddSportsCategory)
		admin.POST("/editsportscategory", controllers.EditSportsCategory)
		admin.POST("/delsportscategory", controllers.DelSportsCategory)
		// 商家分类 (ShopCategory) 路由
		admin.POST("/getshopcategorylist", controllers.GetShopCategoryList)
		admin.POST("/addshopcategory", controllers.AddShopCategory)
		admin.POST("/editshopcategory", controllers.EditShopCategory)
		admin.POST("/delshopcategory", controllers.DelShopCategory)
		// 商家 (Shop) 路由
		admin.POST("/getshoplist", controllers.GetShopList)
		admin.POST("/getshopdetail", controllers.GetShopDetail)
		admin.POST("/addshop", controllers.AddShop)
		admin.POST("/editshop", controllers.EditShop)
		admin.POST("/delshop", controllers.DelShop)

		admin.POST("/geteventlist", controllers.GetEventlist)
		admin.POST("/addevent", controllers.AddEvent)
		admin.POST("/editevent", controllers.EditEvent)
		admin.POST("/delevent", controllers.DelEvent)
		admin.POST("/gettickettemplatelist", controllers.GetTicketTemplatelist)
		admin.POST("/addtickettemplate", controllers.AddTicketTemplate)
		admin.POST("/edittickettemplate", controllers.EditTicketTemplate)
		admin.POST("/deltickettemplate", controllers.DelTicketTemplate)
		admin.POST("/getticketclaimlist", controllers.GetTicketClaimlist)
		admin.POST("/recognizeticket", controllers.RecognizeTicket)
		admin.POST("/getticketrecognitionlist", controllers.GetTicketRecognitionlist)
		admin.POST("/getticketfraudlist", controllers.GetTicketFraudlist)
		admin.POST("/getuserlist", controllers.GetUserlist)
		admin.POST("/adduser", controllers.AddUser)
		admin.POST("/edituser", controllers.EditUser)
		admin.POST("/deluser", controllers.DelUser)
		admin.POST("/getuserflowlist", controllers.GetUserFlowlist)
		admin.POST("/adjustuserpoint", controllers.AdjustUserPoint)
		admin.POST("/adjustuserlevel", controllers.AdjustUserLevel)

		// 存储配置模块接口
		admin.POST("/getstorageconfiglist", controllers.GetStorageConfiglist)
		admin.POST("/getstorageconfigdetail", controllers.GetStorageConfigDetail)
		admin.POST("/addstorageconfig", controllers.AddStorageConfig)
		admin.POST("/updatestorageconfig", controllers.UpdateStorageConfig)
		admin.POST("/deletestorageconfig", controllers.DeleteStorageConfig)
		admin.POST("/setdefaultstorageconfig", controllers.SetDefaultStorageConfig)
		// 新增：上传图片到默认云存储接口（自动带 JWT 鉴权保护）
		admin.POST("/upload", controllers.UploadImageHandler)
	}

	router.Static("/uploads", "./uploads")
	router.Run(":8081")
}

// Cors 跨域中间件
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		var headerKeys []string
		for k := range c.Request.Header {
			headerKeys = append(headerKeys, k)
		}
		headerStr := strings.Join(headerKeys, ", ")
		if headerStr != "" {
			headerStr = fmt.Sprintf("access-control-allow-origin, access-control-allow-headers, %s", headerStr)
		} else {
			headerStr = "access-control-allow-origin, access-control-allow-headers"
		}
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token, Pgtoken, session, X_Requested_With, Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language, DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type, Expires, Last-Modified, Pragma, FooBar")
			c.Header("Access-Control-Max-Age", "172800")
			c.Header("Access-Control-Allow-Credentials", "false")
			c.Set("content-type", "application/json,multipart/form-data,application/x-www-form-urlencoded")
		}

		// 放行所有 OPTIONS 预检请求
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
