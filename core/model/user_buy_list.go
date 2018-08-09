package model

type UserByList struct {
	BaseModel  `xorm:"extends"`
	Uid        int64  `json:"uid" xorm:"notnull index"`
	Gid        int64  `json:"gid" xorm:"notnull index"`
	CardID     int64  `json:"card_id" xorm:"notnull index"`
	StatusCode int    `json:"status_code" xorm:"notnull index"`
	Comment    string `json:"comment" xorm:"varchar(255)"`
}
