package model

import (
	"log"
	"net/http"

	"github.com/atcharles/lotto-chart/core/orm"
	"github.com/gin-gonic/gin"
)

func GinHttpMsg(c *gin.Context, code int) {
	c.AbortWithStatusJSON(code, gin.H{"msg": http.StatusText(code)})
}

func GinHttpWithError(c *gin.Context, code int, err error) {
	c.AbortWithStatusJSON(code, gin.H{"msg": http.StatusText(code), "err": err.Error()})
}

func CheckErrFunc(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 2, "msg": err.Error()})
}

func GinReturnOk(c *gin.Context, data interface{}) {
	c.AbortWithStatusJSON(http.StatusOK, gin.H{"code": 1, "msg": data})
}

//安装,创建数据库
func Initialize() {
	var (
		err error
	)

	log.Println("初始化基本数据...")

	//创建数据引擎
	orm.Initialize()
	userBean := &Users{}
	ltBean := &GameLts{}
	kjBean := &GameKjData{}
	smsBean := &SmsRecords{}
	cardTypesBean := &CardTypes{}
	uoc := &UserOwnCard{}
	beans := []interface{}{
		userBean,
		ltBean,
		kjBean,
		smsBean,
		cardTypesBean,
		uoc,
		new(UserByList),
	}
	if err = orm.Engine.Sync2(beans...); err != nil {
		log.Fatalln("初始化数据表失败:" + err.Error())
	}

	var initDB = func() (err error) {
		if err = userBean.InitData(); err != nil {
			return
		}

		if err = new(SmsConfig).Parse(); err != nil {
			return
		}

		ltBean.InitData()

		//card
		cards := []*CardTypes{{Name: "天卡", Days: 1, Price: 15.05}}
		InitData(cards)

		return
	}

	if err = initDB(); err != nil {
		log.Fatalln("初始化基础数据失败:" + err.Error())
	}
	log.Println("初始化基本数据完成!")
}
