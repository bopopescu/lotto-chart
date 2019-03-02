package model

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/atcharles/gof/goform"
	"github.com/atcharles/lotto-chart/core/orm"
	"github.com/atcharles/lotto-chart/core/records"
	"github.com/gin-gonic/gin"
	"github.com/go-xorm/xorm"
)

type UserByList struct {
	BaseModel     `xorm:"extends"`
	Uid           int64           `json:"uid" xorm:"notnull index"`
	Gid           int64           `json:"gid" xorm:"notnull index"`
	CardID        int64           `json:"card_id" xorm:"notnull index"`
	Status        int             `json:"status" xorm:"notnull index"`
	StatusCode    string          `json:"status_code" xorm:"varchar(10) notnull"`
	Comment       string          `json:"comment" xorm:"varchar(255)"`
	Gateway       string          `json:"gateway" xorm:"notnull varchar(20) index"`          //alipay,bankpay,wxpay
	OnlineOrderID string          `json:"online_order_id" xorm:"notnull varchar(100) index"` //Alipay order_id
	PayTime       goform.JSONTime `json:"pay_time" xorm:"index"`                             //付款时间
	OrderID       string          `json:"order_id" xorm:"varchar(50) notnull unique"`        //订单编号
	Money         float64         `json:"money" xorm:"notnull index"`                        //充值申请金额
	PaidMoney     float64         `json:"paid_money" xorm:"notnull index"`                   //已付金额,实际到账金额
	ManualUpdate  bool            `json:"manual_update" xorm:"notnull index"`
	session       *xorm.Session   `xorm:"-"`
}

func (m *UserByList) StatusParse() {
	ml := map[int]string{
		-1: "已下单",
		1:  "已充值",
		2:  "已拒绝",
	}
	m.StatusCode = ml[m.Status]
}

func (m *UserByList) BeforeInsert() {
	var (
		maxID int64
	)
	m.StatusParse()
	bean := &UserByList{}
	has, _ := orm.Engine.Desc("id").Limit(1).Get(bean)
	if !has {
		maxID = 1
	} else {
		maxID = bean.ID
	}
	m.OrderID = fmt.Sprintf("%03d%03d", m.Uid, maxID)
}

func (m *UserByList) BeforeUpdate() {
	m.StatusParse()
}

func (m *UserByList) Request(c *gin.Context) {
	NormalRequests(c, &UserByList{})
}

func (m *UserByList) GetList(c *gin.Context) {
	var (
		err         error
		requestUser *Users
	)
	v, _ := c.Get("visitor")
	requestUser = v.(*Users)

	qb := &records.QueryBean{}
	if err = c.ShouldBindQuery(qb); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}
	wa := fmt.Sprintf("uid=%d", requestUser.ID)
	if qb.WhereParam != "" {
		qb.WhereParam = qb.WhereParam + "," + wa
	} else {
		qb.WhereParam = wa
	}
	res := records.NewBeanRecords([]*VBuyList{}, qb)
	if err = res.List(); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}
	GinReturnOk(c, res)
}

func (m *UserByList) Put(c *gin.Context) {
	var (
		err error
		a   int64
	)
	bean := &UserByList{}
	if err = c.ShouldBindJSON(bean); err != nil {
		GinHttpWithError(c, http.StatusBadRequest, err)
		return
	}

	bean.session = orm.Engine.NewSession()
	bean.session.Begin()
	defer bean.session.Close()

	bean.ManualUpdate = true
	if bean.Status == 1 {
		bean.PayTime = goform.JSONTime(time.Now())
		bean.PaidMoney = bean.Money
		// update or insert user_own_card

		ownBean := &UserOwnCard{
			Uid:     bean.Uid,
			Gid:     bean.Gid,
			session: bean.session,
		}

		if err = ownBean.Add(bean.CardID); err != nil {
			GinHttpWithError(c, http.StatusInternalServerError, err)
			bean.session.Rollback()
			return
		}

	} else {
		bean.Status = 2
	}
	a, err = bean.session.ID(bean.ID).UseBool().Where("status=-1").Update(bean)
	if err != nil {
		GinHttpWithError(c, http.StatusInternalServerError, err)
		bean.session.Rollback()
		return
	}
	if a == 0 {
		GinHttpWithError(c, http.StatusInternalServerError, errors.New("更新数据失败"))
		bean.session.Rollback()
		return
	}

	bean.session.Commit()
	GinReturnOk(c, bean)
}

type VBuyList struct {
	UserByList   `xorm:"extends"`
	UserName     string `json:"user_name"`
	UserComment  string `json:"user_comment"`
	UserRoleID   int64  `json:"user_role_id"`
	UserRoleName string `json:"user_role_name"`
	LtName       string `json:"lt_name"`
	LtNameCN     string `json:"lt_name_cn"`
	LtEnable     bool   `json:"lt_enable"`
}

func (m *VBuyList) Request(c *gin.Context) {
	NormalRequests(c, &VBuyList{})
}
