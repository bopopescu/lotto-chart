package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sync"

	"github.com/atcharles/gof/goform"
	"github.com/atcharles/gof/gofutils"
	"github.com/atcharles/lotto-chart/core/chart"
	"github.com/atcharles/lotto-chart/core/ini"
	"github.com/atcharles/lotto-chart/core/orm"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type BaseModel struct {
	ID      int64           `json:"id,omitempty" xorm:"autoincr pk"`
	Created goform.JSONTime `json:"created,omitempty" xorm:"created index"`
	Updated goform.JSONTime `json:"updated,omitempty" xorm:"updated index"`
	Deleted goform.JSONTime `json:"deleted,omitempty" xorm:"deleted"`
	Version int64           `json:"version,omitempty" xorm:"version index"`
}

type Users struct {
	BaseModel `xorm:"extends"`
	Name      string `json:"name" xorm:"varchar(20) unique" binding:"required"`
	Phone     string `json:"phone" xorm:"varchar(20) unique"`
	Password  string `json:"password,omitempty"`
	RoleID    int64  `json:"role_id" xorm:"index"`
	RoleName  string `json:"role_name"`
	SmsCode   string `json:"sms_code,omitempty" xorm:"-"`
	CapObj    `xorm:"-"`
}

func (m *Users) InitData() (err error) {
	var (
		a int64
	)
	if a, err = orm.Engine.Count(m); err != nil {
		return
	}
	if a == 0 {
		bean := &Users{
			Name:     "super_manager",
			Phone:    "139xxxxxxxx",
			Password: "123456",
		}
		bean.SetRole(3)
		err = bean.CreateOne()
	}
	return
}

func (m *Users) SetRole(rid int64) error {
	var (
		roleName string
		has      bool
	)
	roles := map[int64]string{
		1: "会员",
		2: "超级会员",
		3: "管理员",
	}
	if roleName, has = roles[rid]; !has {
		return errors.New("不存在的用户角色")
	}

	m.RoleID = rid
	m.RoleName = roleName
	return nil
}

var authMap = &sync.Map{}

func (m *Users) Login(c *gin.Context) {
	var (
		passwordOk, has bool
		err             error
		roleList        []int64
		tokenString     string
	)
	bean := &Users{}
	if err = c.Bind(bean); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "bind user bean:" + err.Error()})
		return
	}
	if !chart.CapVerify(bean.CapKey, bean.Captcha) {
		CheckErrFunc(c, errors.New("验证码不正确"))
		return
	}
	if bean.RoleID == 1 {
		roleList = []int64{1, 2}
	} else if bean.RoleID == 3 {
		roleList = []int64{bean.RoleID}
	} else {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "wrong role"})
		return
	}
	rawUserBean := &Users{}
	has, err = orm.Engine.Where("name = ? or phone = ?", bean.Name, bean.Name).
		In("role_id", roleList).NoAutoCondition().NoCache().Get(rawUserBean)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}
	if !has {
		CheckErrFunc(c, errors.New("用户名/手机号码不存在"))
		return
	}
	if passwordOk, err = bean.VerifyPassword(bean.Password, rawUserBean.Password); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}
	if !passwordOk {
		CheckErrFunc(c, errors.New("密码错误"))
		return
	}
	//对外不显示密码
	rawUserBean.Password = ""

	if bean.RoleID == 3 {
		acValue, ok := authMap.Load(rawUserBean.ID)
		if !ok {
			if tokenString, err = rawUserBean.GenerateToken(); err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
				return
			}
			authMap.Store(rawUserBean.ID, tokenString)
		} else {
			tokenString = acValue.(string)
		}
	} else if bean.RoleID == 1 {
		if tokenString, err = rawUserBean.GenerateToken(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
			return
		}
		authMap.Store(rawUserBean.ID, tokenString)
	}

	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": gin.H{"token": tokenString, "user": rawUserBean}})
}

//GenerateToken 生成令牌
func (m *Users) GenerateToken() (tokenString string, err error) {
	rawBytes, _ := json.Marshal(m)
	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  m.ID,
		"BTS": string(rawBytes),
	})
	if tokenString, err = token.SignedString([]byte(ini.GetSystemKey())); err != nil {
		return
	}
	return
}

func (m *Users) AuthValidator(c *gin.Context) {
	var (
		err    error
		token  *jwt.Token
		claims jwt.MapClaims
		ok     bool
	)
	headerToken := c.GetHeader("Authorization")
	token, err = jwt.Parse(headerToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v ", token.Header["alg"])
		}
		return []byte(ini.GetSystemKey()), nil
	})
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	if claims, ok = token.Claims.(jwt.MapClaims); !(ok && token.Valid) {
		c.AbortWithError(http.StatusUnauthorized, errors.New("invalid token"))
		return
	}

	uid := claims["id"].(float64)
	if mapToken, has := authMap.Load(int64(uid)); !(has && mapToken == headerToken) {
		c.AbortWithError(http.StatusUnauthorized, errors.New("cannot found token in storage or token invalid"))
		return
	}
	bean := &Users{}
	if err = json.Unmarshal([]byte(claims["BTS"].(string)), bean); err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}
	c.Set("visitor", bean)
}

func IsUsableUsername(name string) bool {
	rg := regexp.MustCompile(`^\w{3,10}$`)
	return rg.MatchString(name)
}

func (m *Users) Register(context *gin.Context) {
	var (
		repeat bool
		err    error
	)
	bean := &Users{}
	bean.SetRole(1)
	if err = context.Bind(bean); err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}
	//name \w{3,10},phone \d{11}
	if !IsMobile(bean.Phone) {
		CheckErrFunc(context, errors.New("手机号码不正确"))
		return
	}
	if bean.Name != "" && !IsUsableUsername(bean.Name) {
		CheckErrFunc(context, errors.New("不合法的用户名,由3-10个英文或字母组成"))
		return
	}
	rg := regexp.MustCompile(`^\S{3,20}$`)
	if !rg.MatchString(bean.Password) {
		CheckErrFunc(context, errors.New("密码长度错误,由3-20个字符组成"))
		return
	}
	//短信验证码
	if !VerifySmsCode(bean.Phone, bean.SmsCode) {
		CheckErrFunc(context, errors.New("短信验证码不正确"))
		return
	}

	repeat, err = orm.Engine.Where("name=? or phone=?", bean.Name, bean.Phone).NoCache().NoAutoCondition().Exist(bean)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}
	if repeat {
		CheckErrFunc(context, errors.New("用户名/手机号码重复"))
		return
	}

	if err = bean.CreateOne(); err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"code": 1, "data": bean})
}

func (m *Users) CreateOne() (err error) {
	var (
		a          int64
		newPwdByte []byte
	)
	if m.Name == "" {
		//u, _ := uuid.NewV1()
		//bn := binary.BigEndian.Uint16(u.Bytes())
		m.Name = fmt.Sprintf("user%s", m.Phone)
	}
	newPwdByte = gofutils.AESEncrypt([]byte(ini.GetSystemKey()), []byte(m.Password))
	m.Password = gofutils.BytesToString(newPwdByte)
	if a, err = orm.Engine.InsertOne(m); err != nil {
		return
	}
	if a == 0 {
		return errors.New("创建用户失败")
	}
	m.Password = ""
	return
}

//rawPwd:输入密码->123,EncPwd:加密后的密码->*****
func (m *Users) VerifyPassword(rawPwd, EncPwd string) (ok bool, err error) {
	var (
		newPwdByte []byte
	)
	if newPwdByte, err = gofutils.AESDecrypt([]byte(ini.GetSystemKey()), []byte(EncPwd)); err != nil {
		return
	}
	ok = gofutils.BytesToString(newPwdByte) == rawPwd
	return
}
