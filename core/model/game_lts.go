package model

import (
	"errors"
	"net/http"

	"github.com/atcharles/lotto-chart/core/orm"
	"github.com/gin-gonic/gin"
)

type GameLts struct {
	BaseModel `xorm:"extends"`
	Gid       int64  `json:"gid"  xorm:"notnull unique"`
	Name      string `json:"name" xorm:"varchar(10) notnull index"`
	NameCN    string `json:"name_cn" xorm:"varchar(20) notnull"`
	Enable    bool   `json:"enable" xorm:"notnull index"`
}

func (m *GameLts) InitData() (err error) {
	var (
		a int64
	)
	if a, err = orm.Engine.Count(m); err != nil {
		return
	}
	if a == 0 {
		bean := []*GameLts{
			{Gid: 1, Name: "bjkl8", NameCN: "北京快乐8", Enable: true},
		}
		_, err = orm.Engine.Insert(bean)
	}
	return
}

func (m *GameLts) Request(c *gin.Context) {
	var (
		err error
	)
	bean := &GameLts{}
	if err = c.Bind(bean); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	switch c.Request.Method {
	case "GET":
		//获取所有彩种列表
		beans := make([]*GameLts, 0)
		if err = orm.Engine.Find(&beans); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
			return
		}
		GinReturnOk(c, beans)
		return
	case "POST":
		if err = bean.AddLt(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
			return
		}
		CollectReboot <- true
	case "PUT":
		if err = bean.Update(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
			return
		}
		CollectReboot <- true
	default:
		c.AbortWithStatusJSON(http.StatusMethodNotAllowed, gin.H{"msg": "MethodNotAllowed"})
		return
	}

	GinReturnOk(c, bean)
}

//AddLt 新增彩种,receive 传送彩种结构体
func (m *GameLts) AddLt() (err error) {
	var (
		a int64
	)
	if a, err = orm.Engine.InsertOne(m); err != nil {
		return
	}
	if a == 0 {
		return errors.New("新增彩种失败")
	}
	return
}

func (m *GameLts) Update() (err error) {
	var (
		a int64
	)
	if a, err = orm.Engine.ID(m.ID).AllCols().UseBool("enable").NoAutoCondition().Update(m); err != nil {
		return
	}
	if a == 0 {
		return errors.New("更新彩种信息失败")
	}
	return
}
