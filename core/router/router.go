package router

import (
	"net/http"

	"github.com/atcharles/gof/gofconf"
	"github.com/atcharles/lotto-chart/core/chart"
	"github.com/atcharles/lotto-chart/core/midd"
	"github.com/atcharles/lotto-chart/core/model"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func Router(eg *gin.Engine) *gin.Engine {
	eg.NoRoute(func(c *gin.Context) {
		model.GinHttpMsg(c, http.StatusNotFound)
	})

	eg.NoMethod(func(c *gin.Context) {
		model.GinHttpMsg(c, http.StatusMethodNotAllowed)
	})

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowCredentials = true
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{
		"Origin",
		"Content-Length",
		"Content-Type",
		"Authorization",
	}
	eg.Use(
		gzip.Gzip(gzip.BestCompression),
		cors.New(corsConfig),
		func(c *gin.Context) {
			c.Header(gofconf.HeaderCacheControl, "no-cache,no-store")
			c.Header(gofconf.HeaderPragma, "no-cache,no-store")
			c.Header("X-Server", chart.ServerName+"/"+chart.Version)
		},
	)

	//models
	userBean := &model.Users{}
	gameLtsBean := &model.GameLts{}
	smsCfg := &model.SmsConfig{}
	kjData := &model.GameKjData{}

	//static
	{
		//eg.Static("/apidoc", chart.RootDir+"/apidoc")
		eg.Static("/static", chart.RootDir+"/static")
		eg.Static("/admin", chart.RootDir+"/admin")
		eg.Static("/web", chart.RootDir+"/web")

		eg.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "web")
		})
	}
	//static end

	//root
	{
		eg.GET("/captcha", chart.CaptchaHandler) //验证码
		eg.POST("/SmsCode", smsCfg.SmsSend)      //短信验证码
		eg.POST("/register", userBean.Register)
		eg.POST("/login", userBean.Login)
		eg.POST("/push_file", midd.PushFile)
		eg.POST("/reboot", midd.Reboot)

		//支付回调接口
		eg.POST("/pay", new(model.PayFormData).Pay)
	}

	//所有用户
	api := eg.Group("/Api", userBean.AuthValidator)
	{
		api.GET("/games", gameLtsBean.Request)
	}

	//管理员
	manager := api.Group("/Manager", midd.IsManager)
	{
		//sms账户信息
		manager.Any("/sms", smsCfg.SmsPut)
		//支付宝收款商户设置
		manager.Any("/AliPaySet", new(model.AliPaySet).AliPayPut)
		//彩种设置
		manager.Any("/games", gameLtsBean.Request)
		//点卡设置
		manager.Any("/Cards", new(model.CardTypes).Request)
		//用户列表
		manager.GET("/users", userBean.Request)
		//重置登录密码
		manager.PATCH("/ResetPassword", userBean.ResetPassword)
		//订单列表
		manager.GET("/BuyList", new(model.VBuyList).Request)
		//手动处理订单,通过/拒绝
		manager.PUT("/BuyList", new(model.UserByList).Put)
	}

	//会员
	vip := api.Group("/Vip", midd.IsVip)
	{
		//获取二维码收款信息
		vip.GET("/AliPay", new(model.AliPaySet).AliPayPut)
		//点卡列表
		vip.GET("/Cards", new(model.CardTypes).Request)
		//修改密码
		vip.PATCH("/ChangePassword", userBean.ChangePassword)
		//use
		vip.GET("/History/:gid", new(model.UserOwnCard).Check, kjData.History)
		//查询彩种点卡有效期
		vip.GET("/CardExpire/:gid", new(model.UserOwnCard).CardExpire)
		//购买点卡
		vip.POST("/BuyCard", new(model.UserOwnCard).BuyCard)
		//获取用户订单
		vip.GET("/BuyList", new(model.VBuyList).GetList)
	}

	return eg
}
