package records

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/atcharles/lotto-chart/core/orm"
)

type QueryBean struct {
	Limit          string `json:"limit" form:"limit" binding:"required"`
	OrderParam     string `json:"order_param" form:"order_param"`
	WhereParam     string `json:"where_param" form:"where_param"`
	TimeBetween    string `json:"time_between" form:"time_between"`
	TimeColumnName string `json:"time_column_name" form:"time_column_name"`
}

type BeanRecords struct {
	Count   int64       `json:"count"`
	Records interface{} `json:"records"`
	Qb      *QueryBean  `json:"-"`
}

func NewBeanRecords(rs interface{}, qb *QueryBean) *BeanRecords {
	sv := reflect.ValueOf(rs)
	if sv.Kind() == reflect.Slice {
		rs = reflect.New(sv.Type()).Interface()
	}
	return &BeanRecords{
		Records: rs,
		Qb:      qb,
	}
}

func (op *BeanRecords) List(qus ...string) (err error) {
	var (
		paramString, tum, queryAnd string
		limit                      = make([]int, 0)
		paramSlice                 = make([]string, 0)
		maxRows                    = 100
	)
	strSlice := strings.Split(op.Qb.Limit, ",")
	for _, value := range strSlice {
		a, _ := strconv.Atoi(value)
		limit = append(limit, a)
	}

	if len(limit) < 1 {
		limit = append(limit, 10)
	}

	if op.Qb.OrderParam == "" {
		op.Qb.OrderParam = "id"
	}
	if op.Qb.WhereParam != "" {
		ps := strings.Split(op.Qb.WhereParam, ",")
		paramSlice = append(paramSlice, ps...)
	}
	if op.Qb.TimeColumnName == "" {
		tum = "updated"
	} else {
		tum = op.Qb.TimeColumnName
	}
	if op.Qb.TimeBetween != "" {
		ts := strings.Split(op.Qb.TimeBetween, ",")
		if len(ts) == 2 {
			paramSlice = append(paramSlice, fmt.Sprintf("%s between '%s' and '%s'", tum, ts[0], ts[1]))
		}
	}
	if len(paramSlice) != 0 {
		paramString = strings.Join(paramSlice, " and ")
	}

	limitFun := func() int {
		var (
			rows int
		)
		if len(limit) == 1 {
			rows = limit[0]
		} else {
			rows = limit[1]
		}
		if rows > maxRows {
			rows = maxRows
		}
		return rows
	}

	limitStartFun := func() []int {
		if len(limit) == 1 {
			return nil
		}
		return []int{limit[0]}
	}
	if len(qus) > 0 {
		queryAnd = qus[0]
	}
	err = orm.Engine.NoCache().
		Where(paramString).And(queryAnd).
		Desc(op.Qb.OrderParam).Limit(limitFun(), limitStartFun()...).Find(op.Records)
	if err != nil {
		return
	}

	bean := reflect.New(reflect.ValueOf(op.Records).Elem().Type().Elem().Elem()).Interface()
	op.Count, err = orm.Engine.NoCache().Where(paramString).And(queryAnd).Count(bean)
	if err != nil {
		return
	}
	return
}
