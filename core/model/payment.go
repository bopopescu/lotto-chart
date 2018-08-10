package model

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/atcharles/gof/goform"
	"github.com/atcharles/gof/gofutils"
	"github.com/atcharles/lotto-chart/core/orm"
	"github.com/gin-gonic/gin"
)

var (
	ErrIncorrectOrder = errors.New("IncorrectOrder")
	ErrFail           = fmt.Errorf("Fail ")
)

type PayFormData struct {
	Gateway       string `form:"Gateway"`
	Money         string `form:"Money"`
	AliPayAccount string `form:"alipay_account"`
	Memo          string `form:"memo"`
	Title         string `form:"title"`
	TradeNo       string `form:"tradeNo"`
	Sign          string `form:"Sign"`
	PayTime       string `form:"Paytime"`
}

func (h *PayFormData) Pay(c *gin.Context) {
	err := h.post(c)
	if err != nil {
		var code int
		if err == ErrIncorrectOrder {
			code = http.StatusOK
		} else if err == ErrFail {
			code = http.StatusOK
		} else {
			code = http.StatusInternalServerError
		}
		c.String(code, strings.TrimSpace(err.Error()))
		return
	}
	c.String(http.StatusOK, "Success")
}

func (h *PayFormData) post(c *gin.Context) (err error) {
	res := &PayFormData{}
	if err := c.ShouldBind(res); err != nil {
		return err
	}

	if res.Memo != "lotto-chart" {
		return ErrIncorrectOrder
	}

	buyListBean := &UserByList{OrderID: res.Title}
	has, err := orm.Engine.Get(buyListBean)
	if err != nil {
		return err
	}
	if !has {
		return ErrIncorrectOrder
	}
	if buyListBean.Status != -1 {
		return ErrIncorrectOrder
	}

	aliBean := &AliPaySet{}
	if err = aliBean.Parse(); err != nil {
		return ErrIncorrectOrder
	}

	//check sign
	str := fmt.Sprintf("%s%s%s%s%s%s", aliBean.Sid, aliBean.Secret, res.TradeNo, res.Money, res.Title, res.Memo)
	localSign := strings.ToUpper(gofutils.Md5([]byte(str)))
	if res.Sign != localSign {
		return ErrFail
	}

	//更新订单,添加账户点卡

	buyListBean.Status = 1
	buyListBean.ManualUpdate = false
	buyListBean.OnlineOrderID = res.TradeNo
	buyListBean.PayTime, _ = goform.Todatetime(res.PayTime)
	buyListBean.Gateway = res.Gateway
	buyListBean.PaidMoney, _ = strconv.ParseFloat(res.Money, 64)

	if buyListBean.PaidMoney < buyListBean.Money {
		return ErrFail
	}

	sn := orm.Engine.NewSession()
	buyListBean.session = sn
	sn.Begin()
	defer sn.Close()
	var (
		a int64
	)
	a, err = sn.ID(buyListBean.ID).UseBool().Where("status=-1").Update(buyListBean)
	if err != nil {
		sn.Rollback()
		return
	}
	if a == 0 {
		return ErrIncorrectOrder
	}
	ownBean := &UserOwnCard{
		Uid:     buyListBean.Uid,
		Gid:     buyListBean.Gid,
		session: sn,
	}

	if err = ownBean.Add(buyListBean.CardID); err != nil {
		sn.Rollback()
		return
	}

	return sn.Commit()
}
