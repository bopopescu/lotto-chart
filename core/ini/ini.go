package ini

import (
	"github.com/atcharles/gof/gofutils"
	"github.com/atcharles/lotto-chart/core/chart"
	"github.com/go-ini/ini"
)

var (
	WebFilePath = chart.RootDir + "conf/web.ini"
	Conf        = func() *ini.File {
		var (
			f   *ini.File
			err error
		)
		if err = gofutils.TouchFile(WebFilePath); err != nil {
			panic(err.Error())
		}
		if f, err = ini.Load(WebFilePath); err != nil {
			panic("web.ini 配置文件加载错误:" + err.Error())
		}
		return f
	}()
)

func GetSystemKey() (sysKey string) {
	iniKey := Conf.Section("").Key("system_key")
	sysKey = iniKey.Value()
	if sysKey == "" {
		sysKey = gofutils.URLRandomString(32)
		iniKey.SetValue(sysKey)
		iniKey.Comment = "系统加密密钥,用于密码加密等,如果被修改或删除,系统将无法使用"
		if err := Conf.SaveTo(WebFilePath); err != nil {
			panic("写入加密密钥失败:" + err.Error())
		}
	}
	return
}
