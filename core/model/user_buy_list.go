package model

import (
	"fmt"

	"github.com/atcharles/gof/goform"
	"github.com/atcharles/lotto-chart/core/orm"
	"github.com/gin-gonic/gin"
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
	}
	m.OrderID = fmt.Sprintf("%2d%2d", m.Uid, maxID)
}

func (m *UserByList) BeforeUpdate() {
	m.StatusParse()
	switch m.Status {
	case -1:
		//已下单
	case 1:
		//已充值
	case 2:
		//已拒绝
	}
}

func (m *UserByList) Request(c *gin.Context) {
	NormalRequests(c, &UserByList{})
}
