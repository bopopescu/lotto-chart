package model

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/atcharles/gof/goform"
	"github.com/atcharles/lotto-chart/core/chart"
	"github.com/atcharles/lotto-chart/core/ini"
	"github.com/atcharles/lotto-chart/core/orm"
	"github.com/gin-gonic/gin"
	"github.com/labstack/gommon/random"
	"github.com/levigross/grequests"
)

/**
发送验证码1分钟只能点击发送1次；
相同IP手机号码1天最多提交20次；
验证码短信单个手机号码30分钟最多提交10次；
在提交页面加入图形校验码，防止机器人恶意发送；
在发送验证码接口程序中，判断图形校验码输入是否正确；
新用户用接口测试验证码时，请勿输入：测试等无关内容信息，请直接输入：验证码:xxxxxx，发送。
接口发送触发短信时，您可以把短信内容提供给客服绑定短信模板，绑定后24小时即时发送。未绑定模板的短信21点以后提交，隔天才能收到。
*/

const (
	SmsMaxSendCountDay = 10
	SmsExpireMinute    = 30 //验证码有效期30分钟
	SmsUrl             = `http://utf8.api.smschinese.cn/?Uid=%s&Key=%s&smsMob=%s&smsText=%s`
	SmsCodeText        = `验证码:%s,有效时间30分钟,请尽快完成验证.`
)

var (
	smsMobileMap = &sync.Map{}
)

type SmsStorageMap map[string]interface{}

type CapObj struct {
	Captcha string `json:"captcha,omitempty"`
	CapKey  string `json:"cap_key,omitempty"`
}
type SmsConfig struct {
	CapObj `ini:"-"`
	Mobile string `json:"mobile,omitempty" ini:"-"`
	Uid    string `json:"uid" ini:"sms_username"`
	Key    string `json:"key" ini:"sms_key"`
}

func (s *SmsConfig) SmsPut(c *gin.Context) {
	var (
		err error
		cfg = ini.Conf
		sn  = cfg.Section("sms")
	)
	bean := &SmsConfig{}
	switch c.Request.Method {
	case "GET":
		if err = sn.MapTo(bean); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
			return
		}
		GinReturnOk(c, bean)
	case "PUT":
		if err = c.ShouldBindJSON(bean); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
			return
		}

		if err = sn.ReflectFrom(bean); err != nil {
			return
		}
		if err = cfg.SaveTo(ini.WebFilePath); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
			return
		}
		GinReturnOk(c, "设置成功")
	default:
		c.AbortWithStatusJSON(http.StatusMethodNotAllowed, gin.H{"msg": "MethodNotAllowed"})
	}
}

func (s *SmsConfig) SmsSend(c *gin.Context) {
	var err error
	bean := &SmsConfig{}
	if err = c.ShouldBind(bean); err != nil {
		GinHttpWithError(c, http.StatusBadRequest, err)
		return
	}
	//图形验证码
	if !chart.CapVerify(bean.CapKey, bean.Captcha) {
		CheckErrFunc(c, errors.New("验证码不正确"))
		return
	}
	//验证手机号码
	if !IsMobile(bean.Mobile) {
		CheckErrFunc(c, errors.New("手机号码不正确"))
		return
	}
	if err = bean.SendCode(c.ClientIP(), bean.Mobile); err != nil {
		CheckErrFunc(c, err)
		return
	}
	GinReturnOk(c, "验证码发送成功")
}

func IsMobile(mNumber string) bool {
	rg := regexp.MustCompile(`^(13[0-9]|14[579]|15[0-3,5-9]|16[6]|17[0135678]|18[0-9]|19[89])\d{8}$`)
	return rg.MatchString(mNumber)
}

func VerifySmsCode(mobile, smsCode string) bool {
	v, ok := smsMobileMap.Load(mobile)
	if !ok {
		return false
	}
	sv := v.(SmsStorageMap)
	nextEnableSendTime := sv["expire"].(time.Time)
	if nextEnableSendTime.Before(time.Now()) {
		return false
	}
	return sv["code"].(string) == smsCode
}

func (s *SmsConfig) Parse() (err error) {
	cfg := ini.Conf
	if err = cfg.Section("sms").MapTo(s); err != nil {
		return
	}
	if s.Uid == "" {
		//reflect
		sn := cfg.Section("sms")
		sn.Comment = "短信中心设置"
		if err = sn.ReflectFrom(s); err != nil {
			return
		}
		return cfg.SaveTo(ini.WebFilePath)
	}
	return
}

func (s *SmsConfig) SendCode(ip, mobile string) (err error) {
	var (
		smsCode   = random.New().String(4, random.Numeric)
		rp        *grequests.Response
		resString string
		counts    int64
	)
	smsBean := &SmsRecords{IP: ip, Mobile: mobile}
	counts, err = orm.Engine.Where("created > ?", goform.JSONTime(time.Now().Add(-1*time.Hour*24)).String()).Count(smsBean)
	if err != nil {
		return
	}
	if counts > SmsMaxSendCountDay {
		return fmt.Errorf("24小时内信息发送大于%d次", SmsMaxSendCountDay)
	}

	//验证手机号码是否可以发送信息
	if v, ok := smsMobileMap.Load(mobile); ok {
		sv := v.(SmsStorageMap)
		nextEnableSendTime := sv["expire"].(time.Time)
		if nextEnableSendTime.After(time.Now()) {
			return fmt.Errorf("至%s之前,手机号码%s无法发送新的信息", goform.JSONTime(nextEnableSendTime).String(), mobile)
		}
	}

	if err = s.Parse(); err != nil {
		return
	}
	if s.Uid == "" || s.Key == "" {
		return errors.New("短信中心信息未设置,无法发送")
	}
	sendUrl := fmt.Sprintf(SmsUrl, s.Uid, s.Key, mobile, fmt.Sprintf(SmsCodeText, smsCode))
	rp, err = grequests.Get(sendUrl, &grequests.RequestOptions{
		RequestTimeout: time.Second * 10,
	})
	if err != nil {
		return
	}
	defer rp.Close()
	resString = rp.String()

	//不管发送是否成功,都记录到数据库
	smsBean.SmsCode = smsCode
	smsBean.ResultCode = resString
	if _, err = orm.Engine.InsertOne(smsBean); err != nil {
		return
	}

	if resString != "1" {
		return fmt.Errorf("网关接口短信发送失败:返回值:%s", resString)
	}

	//记录手机号码下次发送有效时间
	smsMobileMap.Store(mobile, SmsStorageMap{
		"code":   smsCode,
		"expire": time.Now().Add(time.Minute * SmsExpireMinute),
	})

	return
}

type SmsRecords struct {
	BaseModel  `xorm:"extends"`
	Mobile     string `json:"mobile" xorm:"varchar(11) notnull index"`
	IP         string `json:"ip" xorm:"varchar(20) notnull index"`
	SmsCode    string `json:"sms_code" xorm:"varchar(10) notnull index"`
	ResultCode string `json:"result_code" xorm:"varchar(10) notnull index"`
}
