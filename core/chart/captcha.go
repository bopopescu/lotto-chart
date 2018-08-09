package chart

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"github.com/satori/go.uuid"
)

const (
	//宽
	dx = 240
	//高
	dy = 80
)

var (
	capStore = func() base64Captcha.Store {
		st := base64Captcha.NewMemoryStore(base64Captcha.GCLimitNumber, base64Captcha.Expiration)
		base64Captcha.SetCustomStore(st)
		return st
	}()
)

type Captcha struct {
	Content []byte
	Text    string
	Key     string
}

var CaptchaHandler gin.HandlerFunc = func(c *gin.Context) {
	key := c.Query("key")
	var capN *Captcha
	if key != "" {
		capN = NewCapFresh(key)
	} else {
		capN = NewCap()
	}
	c.Header("Access-Control-Expose-Headers", "Set-Cap-Key")
	c.Header("Set-Cap-Key", capN.Key)
	c.Data(http.StatusOK, "image/png", capN.Content)
}

func NewCap() *Captcha {
	return NewCapFresh("")
}

func NewCapFresh(key string) *Captcha {
	uuidData, _ := uuid.NewV1()
	if key == "" {
		key = "captcha_" + uuidData.String()
	}
	cp := &Captcha{Key: key}
	config := base64Captcha.ConfigCharacter{
		Height:           dy,
		Width:            dx,
		CaptchaLen:       4,
		Mode:             base64Captcha.CaptchaModeNumber,
		IsUseSimpleFont:  false,
		IsShowHollowLine: false,
		IsShowNoiseDot:   false,
		IsShowNoiseText:  false,
		IsShowSlimeLine:  false,
		IsShowSineLine:   false,
	}
	char := base64Captcha.EngineCharCreate(config)
	cp.Text = char.VerifyValue
	cp.Content = char.BinaryEncodeing()
	capStore.Set(cp.Key, cp.Text)
	return cp
}

func CapVerify(key string, text string) bool {
	return base64Captcha.VerifyCaptcha(key, text)
}
