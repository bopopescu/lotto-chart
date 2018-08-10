package model

import (
	"net/http"

	"github.com/atcharles/lotto-chart/core/ini"
	"github.com/gin-gonic/gin"
)

type AliPaySet struct {
	Sid    string `json:"sid" ini:"sid" comment:"商户名称"`
	Secret string `json:"secret" ini:"secret" comment:"商户密钥"`
}

func (m *AliPaySet) AliPayPut(c *gin.Context) {
	var (
		bean = &AliPaySet{}
		err  error
	)
	switch c.Request.Method {
	case "GET":
		if err = bean.Parse(); err != nil {
			GinHttpWithError(c, http.StatusInternalServerError, err)
			return
		}
		GinReturnOk(c, bean)
	case "PUT":
		if err := c.ShouldBindJSON(bean); err != nil {
			GinHttpWithError(c, http.StatusBadRequest, err)
			return
		}
		if err = ini.Conf.Section("alipay").ReflectFrom(bean); err != nil {
			GinHttpWithError(c, http.StatusInternalServerError, err)
			return
		}
		if err = ini.Conf.SaveTo(ini.WebFilePath); err != nil {
			GinHttpWithError(c, http.StatusInternalServerError, err)
			return
		}
		GinReturnOk(c, "设置成功")
	default:
		GinHttpMsg(c, http.StatusMethodNotAllowed)
		return
	}
}

func (m *AliPaySet) Parse() (err error) {
	sn := ini.Conf.Section("alipay")
	if err = sn.MapTo(m); err != nil {
		return
	}

	if m.Sid == "" {
		sn.Comment = "支付宝商户设置"
		if err = sn.ReflectFrom(m); err != nil {
			return
		}
		return ini.Conf.SaveTo(ini.WebFilePath)
	}

	return
}
