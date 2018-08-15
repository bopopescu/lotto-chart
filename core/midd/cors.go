package midd

import (
	"github.com/atcharles/gof/gofconf"
	"github.com/atcharles/lotto-chart/core/chart"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowCredentials = true
	corsConfig.AllowAllOrigins = false
	corsConfig.AllowOriginFunc = func(origin string) bool {
		return true
	}
	corsConfig.AddAllowMethods("OPTIONS")
	corsConfig.AddAllowHeaders("Authorization")
	return cors.New(corsConfig)
}

func CustomMiddleware(c *gin.Context) {
	c.Header(gofconf.HeaderCacheControl, "no-cache,no-store")
	c.Header(gofconf.HeaderPragma, "no-cache,no-store")
	c.Header("X-Server", chart.ServerName+"/"+chart.Version)
}
