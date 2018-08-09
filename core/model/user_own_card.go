package model

import "github.com/atcharles/gof/goform"

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
	Gid             int64           `json:"gid" xorm:"notnull index"`
	CardTypeID      int64           `json:"card_type_id" xorm:"notnull index"`
	Expire          goform.JSONTime `json:"expire" xorm:"notnull index"`
	ExpireTimestamp int64           `json:"expire_timestamp" xorm:"notnull index"`
}
