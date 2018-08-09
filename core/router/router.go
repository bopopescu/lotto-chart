package router

import (
	"github.com/atcharles/lotto-chart/core/chart"
	"github.com/atcharles/lotto-chart/core/midd"
	"github.com/atcharles/lotto-chart/core/model"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func Router(eg *gin.Engine) *gin.Engine {
	eg.Use(gzip.Gzip(gzip.BestCompression))
	eg.Use(cors.Default())
	eg.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store")
	})

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
	}

	//所有用户
	api := eg.Group("/Api")
	api.Use(userBean.AuthValidator)
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
	}

	//会员
	vip := api.Group("/Vip", midd.IsVip)
	{
		//use
		vip.GET("/History", kjData.History)
	}

	return eg
}
