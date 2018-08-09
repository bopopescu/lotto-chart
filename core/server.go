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
	webLog := chart.RootDir + "logs/web/log.log"
	gin.DefaultWriter = goflogger.GetFile(webLog).GetFile()
	gin.DefaultErrorWriter = goflogger.GetFile(chart.RootDir + "logs/web/err.log").GetFile()

	gin.SetMode("release")
	eg := gin.Default()
	router.Router(eg)

	//
	go model.CollectKjData()

	log.Fatalln(eg.Run(":39100"))
}
