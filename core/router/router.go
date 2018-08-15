package router

import (
	"net/http"

	"github.com/atcharles/lotto-chart/core/chart"
	"github.com/atcharles/lotto-chart/core/midd"
	"github.com/atcharles/lotto-chart/core/model"
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
	eg.Use(gzip.Gzip(gzip.BestCompression), midd.Cors(), midd.CustomMiddleware)

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
	//root
	{
		eg.GET("/captcha", chart.CaptchaHandler)          //验证码
		eg.POST("/SmsCode", new(model.SmsConfig).SmsSend) //短信验证码
		eg.POST("/register", new(model.Users).Register)
		eg.POST("/login", new(model.Users).Login)
		eg.POST("/push_file", midd.PushFile)
		eg.POST("/reboot", midd.Reboot)

		//支付回调接口
		eg.POST("/pay", new(model.PayFormData).Pay)
	}

	//所有用户
	api := eg.Group("/Api", new(model.Users).AuthValidator)
	{
		api.GET("/games", new(model.GameLts).Request)
	}

	//管理员
	manager := api.Group("/Manager", midd.IsManager)
	{
		//sms账户信息
		manager.Any("/sms", new(model.SmsConfig).SmsPut)
		//支付宝收款商户设置
		manager.Any("/AliPaySet", new(model.AliPaySet).AliPayPut)
		//彩种设置
		manager.Any("/games", new(model.GameLts).Request)
		//点卡设置
		manager.Any("/Cards", new(model.CardTypes).Request)
		//用户列表
		manager.GET("/users", new(model.Users).Request)
		//重置登录密码
		manager.PATCH("/ResetPassword", new(model.Users).ResetPassword)
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
		vip.PATCH("/ChangePassword", new(model.Users).ChangePassword)
		//use
		vip.GET("/History/:gid", new(model.UserOwnCard).Check, new(model.GameKjData).History)
		//查询彩种点卡有效期
		vip.GET("/CardExpire/:gid", new(model.UserOwnCard).CardExpire)
		//购买点卡
		vip.POST("/BuyCard", new(model.UserOwnCard).BuyCard)
		//获取用户订单
		vip.GET("/BuyList", new(model.VBuyList).GetList)
	}

	return eg
}
