package router

import (
	"net/http"

	"github.com/atcharles/lotto-chart/core/chart"
	"github.com/atcharles/lotto-chart/core/midd"
	"github.com/atcharles/lotto-chart/core/model"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func Router(eg *gin.Engine) *gin.Engine {
	eg.Use(
		gzip.Gzip(gzip.BestCompression),
		cors.Default(),
		func(c *gin.Context) {
			c.Header("Cache-Control", "no-cache, no-store")

			c.Next()

			c.Header("X-Server", chart.ServerName+"/"+chart.Version)
			if c.Writer.Status() == http.StatusNotFound {
				c.JSON(http.StatusNotFound, gin.H{"msg": http.StatusText(http.StatusNotFound)})
			}
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

		eg.POST("/pay", midd.Reboot)
	}

	//所有用户
	api := eg.Group("/Api", userBean.AuthValidator)
	{
		api.GET("/games", gameLtsBean.Request)
	}

	//管理员
	manager := api.Group("/Manager", midd.IsManager)
	{
		manager.Any("/games", gameLtsBean.Request)
		//sms账户信息
		manager.Any("/sms", smsCfg.SmsPut)
		manager.Any("/Cards", new(model.CardTypes).Request)
		//用户列表
		manager.GET("/users", userBean.Request)
		//重置登录密码
		manager.PATCH("/ResetPassword", userBean.ResetPassword)
		//订单列表
		manager.GET("/BuyList", new(model.UserByList).Request)
		//手动处理订单,通过/拒绝
		manager.PUT("/BuyList", new(model.UserByList).Put)
	}

	//会员
	vip := api.Group("/Vip", midd.IsVip)
	{
		//修改密码
		vip.PATCH("/ChangePassword", userBean.ChangePassword)
		//use
		vip.GET("/History/:gid", new(model.UserOwnCard).Check, kjData.History)
		//查询彩种点卡有效期
		vip.GET("/CardExpire/:gid", new(model.UserOwnCard).CardExpire)
		//购买点卡
		vip.POST("/BuyCard", new(model.UserOwnCard).BuyCard)
	}

	return eg
}
