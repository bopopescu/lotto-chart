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
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
			return
		}
		slice := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(bean)), 0, 0)
		beans := reflect.New(slice.Type())
		beans.Elem().Set(slice)
		res := records.NewBeanRecords(beans.Interface(), qb)
		if err = res.List(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
			return
		}
		GinReturnOk(c, res)
		return
	case "POST":
		if err = c.Bind(bean); err != nil {
			return
		}
		if _, err = orm.Engine.InsertOne(bean); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
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
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
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
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
			return
		}
		if a == 0 {
			CheckErrFunc(c, errors.New("删除失败"))
			return
		}
		GinReturnOk(c, "删除成功")
		return
	default:
		c.AbortWithStatusJSON(http.StatusMethodNotAllowed, gin.H{"msg": "MethodNotAllowed"})
		return
	}
	GinReturnOk(c, bean)
}
