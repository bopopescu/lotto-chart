package model

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/atcharles/gof/goform"
	"github.com/atcharles/lotto-chart/core/orm"
	"github.com/gin-gonic/gin"
	"github.com/go-xorm/xorm"
)

/**
用户拥有的点卡列表
查询的时候,取 expire 时间大于当前时间的列表

/如果存在,则说明,用户有权限
取expire时间最久的一个,就是到期时间

#####
购买保存
取 expire 时间大于当前时间的列表
/如果存在,取expire最大的时间,将新买的点卡有效期加到上面,保存数据
//不存在直接保存
*/
type UserOwnCard struct {
	BaseModel       `xorm:"extends"`
	Uid             int64           `json:"uid" xorm:"notnull index"`
	Gid             int64           `json:"gid" xorm:"notnull index"`
	Expire          goform.JSONTime `json:"expire" xorm:"notnull index"`
	ExpireTimestamp int64           `json:"expire_timestamp" xorm:"notnull index"`
	session         *xorm.Session   `json:"-" xorm:"-"`
}

func (m *UserOwnCard) BeforeInsert() {
	m.ExpireTimestamp = time.Time(m.Expire).Unix()
}

func (m *UserOwnCard) BeforeUpdate() {
	m.ExpireTimestamp = time.Time(m.Expire).Unix()
}

var (
	ErrorNoCardHad = errors.New("没有购买点卡")
	ErrorNoTime    = errors.New("点卡已过期")
)

type BuyCardObj struct {
	Gid    int64 `json:"gid" binding:"required"`
	CardID int64 `json:"card_id" binding:"required"`
}

func (m *BuyCardObj) GetCard() (card *CardTypes, err error) {
	var (
		has bool
	)
	card = &CardTypes{}
	card.ID = m.CardID
	has, err = orm.Engine.Get(card)
	if err != nil {
		return
	}
	if !has {
		err = errors.New("没有这个点卡")
		return
	}
	return
}

func (m *UserOwnCard) BuyCard(c *gin.Context) {
	//购买点卡,1:gid,2:card_id
	var (
		err         error
		requestUser *Users
		card        *CardTypes
	)
	v, _ := c.Get("visitor")
	requestUser = v.(*Users)
	buyBean := &BuyCardObj{}
	if err = c.ShouldBindJSON(buyBean); err != nil {
		GinHttpWithError(c, http.StatusBadRequest, err)
		return
	}
	card, err = buyBean.GetCard()
	if err != nil {
		GinHttpWithError(c, http.StatusInternalServerError, err)
		return
	}
	listBean := &UserByList{
		Uid:    requestUser.ID,
		Gid:    buyBean.Gid,
		CardID: buyBean.CardID,
		Status: -1,
		Money:  card.Price,
	}

	var has bool
	has, err = orm.Engine.Get(listBean)
	if err != nil {
		GinHttpWithError(c, http.StatusInternalServerError, err)
		return
	}
	if has {
		GinReturnOk(c, listBean)
		return
	}

	if _, err = orm.Engine.InsertOne(listBean); err != nil {
		GinHttpWithError(c, http.StatusInternalServerError, err)
		return
	}

	GinReturnOk(c, listBean)
}

func (m *UserOwnCard) CardExpire(c *gin.Context) {
	v, _ := c.Get("visitor")
	var (
		err         error
		requestUser = v.(*Users)
	)
	pGid := c.Param("gid")
	gid, _ := strconv.Atoi(pGid)
	bean := &UserOwnCard{Gid: int64(gid), Uid: requestUser.ID}
	if err = bean.Get(); err != nil {
		if err == ErrorNoCardHad {
			CheckErrFunc(c, err)
		} else {
			GinHttpWithError(c, http.StatusInternalServerError, err)
		}
		return
	}
	GinReturnOk(c, bean)
}

func (m *UserOwnCard) Check(c *gin.Context) {
	var (
		err error
	)
	pGid := c.Param("gid")
	gid, _ := strconv.Atoi(pGid)
	bean := &UserOwnCard{Gid: int64(gid)}
	visitor, _ := c.Get("visitor")
	requestUser := visitor.(*Users)
	bean.Uid = requestUser.ID
	if err = bean.Get(); err != nil {
		if err == ErrorNoCardHad {
			CheckErrFunc(c, err)
		} else {
			GinHttpWithError(c, http.StatusInternalServerError, err)
		}
		return
	}
	if time.Now().After(time.Time(bean.Expire)) {
		CheckErrFunc(c, ErrorNoTime)
		return
	}
	c.Set("OWN", bean)
}

func (m *UserOwnCard) Get() (err error) {
	var has bool
	has, err = orm.Engine.Get(m)
	if err != nil {
		return
	}
	if !has {
		return ErrorNoCardHad
	}
	return
}

//Add must be in a session,添加点卡时间
func (m *UserOwnCard) Add(cardID int64) (err error) {
	var (
		card *CardTypes
		a    int64
	)

	byCard := &BuyCardObj{CardID: cardID}
	card, err = byCard.GetCard()
	if err != nil {
		return
	}

	normalTime := goform.JSONTime(time.Now().Add(time.Hour * time.Duration(24*card.Days)))
	if err = m.Get(); err != nil {
		if err == ErrorNoCardHad {
			m.Expire = normalTime
			_, err = m.session.InsertOne(m)
		}
		return
	}

	//过期的点卡,从当前时间开始计时
	if time.Time(m.Expire).Before(time.Now()) {
		m.Expire = normalTime
	} else {
		m.Expire = goform.JSONTime(time.Time(m.Expire).Add(time.Hour * time.Duration(24*card.Days)))
	}

	a, err = m.session.ID(m.ID).Update(m)
	if err != nil {
		return err
	}
	if a == 0 {
		return errors.New("更新数据失败")
	}

	return
}
