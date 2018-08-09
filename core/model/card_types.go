package model

import (
	"github.com/gin-gonic/gin"
)

type CardTypes struct {
	BaseModel `xorm:"extends"`
	Name      string  `json:"name" xorm:"varchar(20) notnull"`
	Days      int     `json:"days" xorm:"notnull"`
	Price     float64 `json:"price" xorm:"notnull"`
}

func (m *CardTypes) Request(c *gin.Context) {
	NormalRequests(c, &CardTypes{})
}
