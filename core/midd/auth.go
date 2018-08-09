package midd

import (
	"net/http"

	"github.com/atcharles/lotto-chart/core/model"
	"github.com/gin-gonic/gin"
)

var IsVip gin.HandlerFunc = func(c *gin.Context) {
	userBean, has := c.Get("visitor")
	if !has {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ub := userBean.(*model.Users)
	if !(ub.RoleID == 1 || ub.RoleID == 2) {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
}
var IsManager gin.HandlerFunc = func(c *gin.Context) {
	userBean, has := c.Get("visitor")
	if !has {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ub := userBean.(*model.Users)
	if !(ub.RoleID == 3) {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
}
