package model

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/atcharles/lotto-chart/core/orm"
	"github.com/atcharles/lotto-chart/core/records"
	"github.com/gin-gonic/gin"
)

func NormalRequests(c *gin.Context, bean interface{}) {
	var (
		err error
	)

	switch c.Request.Method {
	case "GET":
		qb := &records.QueryBean{}
		if err = c.ShouldBindQuery(qb); err != nil {
			GinHttpWithError(c, http.StatusBadRequest, err)
			return
		}
		slice := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(bean)), 0, 0)
		beans := reflect.New(slice.Type())
		beans.Elem().Set(slice)
		res := records.NewBeanRecords(beans.Interface(), qb)
		if err = res.List(); err != nil {
			GinHttpWithError(c, http.StatusInternalServerError, err)
			return
		}
		GinReturnOk(c, res)
		return
	case "POST":
		if err = c.Bind(bean); err != nil {
			return
		}
		if _, err = orm.Engine.InsertOne(bean); err != nil {
			GinHttpWithError(c, http.StatusInternalServerError, err)
			return
		}
	case "PUT":
		if err = c.Bind(bean); err != nil {
			return
		}
		beanValue := reflect.ValueOf(bean).Elem()
		id := beanValue.FieldByName("ID").Int()
		var a int64
		a, err = orm.Engine.ID(id).Update(bean)
		if err != nil {
			GinHttpWithError(c, http.StatusInternalServerError, err)
			return
		}
		if a == 0 {
			CheckErrFunc(c, errors.New("更新数据失败"))
			return
		}
		//case "PATCH":
	case "DELETE":
		if err = c.Bind(bean); err != nil {
			return
		}
		var a int64
		if a, err = orm.Engine.Delete(bean); err != nil {
			GinHttpWithError(c, http.StatusInternalServerError, err)
			return
		}
		if a == 0 {
			CheckErrFunc(c, errors.New("删除失败"))
			return
		}
		GinReturnOk(c, "删除成功")
		return
	default:
		GinHttpMsg(c, http.StatusMethodNotAllowed)
		return
	}
	GinReturnOk(c, bean)
}

func InitData(beans interface{}) (err error) {
	var (
		a int64
	)
	m := reflect.New(reflect.Indirect(reflect.ValueOf(beans)).Type().Elem().Elem()).Interface()
	if a, err = orm.Engine.Count(m); err != nil {
		return
	}
	if a == 0 {
		_, err = orm.Engine.Insert(beans)
	}
	return
}
