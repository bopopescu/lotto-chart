package model

type CardTypes struct {
	Name string `json:"name" xorm:"varchar(20) notnull"`
	Days int    `json:"days" xorm:"notnull"`
}
