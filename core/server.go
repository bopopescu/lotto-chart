package core

import (
	"log"

	"github.com/atcharles/gof/goflogger"
	"github.com/atcharles/lotto-chart/core/chart"
	"github.com/atcharles/lotto-chart/core/model"
	"github.com/atcharles/lotto-chart/core/orm"
	"github.com/atcharles/lotto-chart/core/router"
	"github.com/gin-gonic/gin"
)

func GinServer() {
	orm.Initialize()

	gin.DisableConsoleColor()
	gin.DefaultWriter = goflogger.GetFile(chart.RootDir + "logs/web/log.log").GetFile()
	gin.DefaultErrorWriter = goflogger.GetFile(chart.RootDir + "logs/web/err.log").GetFile()
	gin.SetMode("release")

	//
	go model.CollectKjData()

	log.Fatalln(router.Router(gin.Default()).Run(":39100"))
}
