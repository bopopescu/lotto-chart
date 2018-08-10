package model

import (
	"errors"
	"net/http"

	"github.com/atcharles/lotto-chart/core/orm"
	"github.com/gin-gonic/gin"
)

var (
	ErrIncorrectOrder = errors.New("IncorrectOrder")
	ErrInvalidSign    = errors.New("invalid sign")
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
		} else if err == ErrInvalidSign {
			code = http.StatusOK
		} else {
			code = http.StatusInternalServerError
		}
		c.String(code, err.Error())
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

	rl := &UserByList{OrderID: res.Title}
	has, err := orm.Engine.Get(rl)
	if err != nil {
		return err
	}
	if !has {
		return ErrIncorrectOrder
	}
	if rl.Status != -1 {
		return ErrIncorrectOrder
	}

	return
}
