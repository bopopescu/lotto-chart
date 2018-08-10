package model

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/atcharles/gof/goflogger"
	"github.com/atcharles/gof/goform"
	"github.com/atcharles/lotto-chart/core/chart"
	"github.com/atcharles/lotto-chart/core/orm"
	"github.com/atcharles/lotto-chart/core/records"
	"github.com/gin-gonic/gin"
	"github.com/levigross/grequests"
)

const (
	CollectServerURL = "http://120.78.59.194:8881/open"
)

var (
	ErrKjDataExist = errors.New("数据已存在") //采集获取的开奖数据已经存在
	collectLogger  = goflogger.GetFile(chart.RootDir + "logs/collect/log.log")
	collectMap     = &sync.Map{} //采集彩种对象
	CollectReboot  = make(chan bool, 1)
)

type GameKjData struct {
	BaseModel         `xorm:"extends"`
	Gid               int64           `json:"gid" xorm:"notnull index"`
	Issue             string          `json:"issue" xorm:"varchar(20) notnull index"`
	NextIssue         string          `json:"next_issue" xorm:"varchar(20) notnull index"`
	OpenNumber        string          `json:"open_number" xorm:"notnull"`
	OpenTime          goform.JSONTime `json:"open_time" xorm:"notnull"`
	OpenTimestamp     int64           `json:"open_timestamp" xorm:"notnull index"`
	NextOpenTime      goform.JSONTime `json:"next_open_time"`
	NextOpenTimestamp int64           `json:"next_open_timestamp" xorm:"notnull index"`
	GetTime           goform.JSONTime `json:"get_time"`
	GetTimestamp      int64           `json:"get_timestamp"`
}

//获取采集到的开奖数据
/**
[
    {
        "gid": 8,
        "name": "北京快乐8",
        "issue": "903095",
        "open_number": "01,05,16,26,33,36,38,39,43,48,51,52,56,58,59,63,66,68,73,74+00",
        "open_time": "2018-08-07 19:25:00",
        "open_timestamp": 1533641100,
        "next_issue": "903096",
        "next_open_time": "2018-08-07 19:30:00",
        "next_open_timestamp": 1533641400
    }
]
*/
func (m *GameKjData) collectAction(lt *GameLts) (err error) {
	var (
		rp  *grequests.Response
		has bool
		a   int64
	)
	getUrl := fmt.Sprintf("%s?name=%s&row=1", CollectServerURL, lt.Name)
	rp, err = grequests.Get(getUrl, &grequests.RequestOptions{
		RequestTimeout: time.Second * 10,
	})
	if err != nil {
		return
	}
	defer rp.Close()
	gbs := []*GameKjData{m}
	if err = rp.JSON(&gbs); err != nil {
		return
	}
	m.Gid = lt.Gid
	m.GetTime = goform.JSONTime(time.Now())
	m.GetTimestamp = time.Now().Unix()

	if m.NextOpenTimestamp-time.Now().Unix() <= 0 {
		return errors.New("开奖中... ")
	}

	m1 := &GameKjData{}
	has, err = orm.Engine.Where("gid=? and issue=?", lt.Gid, m.Issue).NoAutoCondition().NoCache().Get(m1)
	if err != nil {
		return
	}
	if has {
		return ErrKjDataExist
	}
	a, err = orm.Engine.InsertOne(m)
	if err != nil {
		return
	}
	if a == 0 {
		return errors.New("保存开奖数据失败")
	}
	return
}

type Collection struct {
	wait time.Duration
	lt   *GameLts
	stop chan bool
}

func (c *Collection) cron() {
	tk := time.NewTicker(c.wait)
	defer tk.Stop()
	for {
		select {
		case <-tk.C:
			var (
				waitSec int64
			)
			kjData := &GameKjData{}
			if err := kjData.collectAction(c.lt); err != nil {
				if err == ErrKjDataExist {
					waitSec = kjData.NextOpenTimestamp - time.Now().Unix()
					tk = time.NewTicker(time.Second * time.Duration(waitSec))
				} else {
					waitSec = int64(c.wait.Seconds())
					tk = time.NewTicker(c.wait)
				}
				collectLogger.GetLogger().Warningf("彩种[%s]第[%s]期:%s. 线程休眠%d秒",
					c.lt.NameCN, kjData.NextIssue, err.Error(), waitSec)
				continue
			}
			waitSec = kjData.NextOpenTimestamp - time.Now().Unix()
			tk = time.NewTicker(time.Second * time.Duration(waitSec))
			collectLogger.GetLogger().Infof("彩种[%s]第[%s]期,号码[%s],数据保存成功.线程休眠%d秒",
				c.lt.NameCN, kjData.Issue, kjData.OpenNumber, waitSec)
		case <-c.stop:
			return
		}
	}
}

func collectStart() {
	var (
		err error
	)
	beans := make([]*GameLts, 0)
	if err = orm.Engine.Where("enable=1").Find(&beans); err != nil {
		collectLogger.GetLogger().Errorf("get lts error:%s", err.Error())
		return
	}
	for _, lt := range beans {
		cBean := &Collection{lt: lt, wait: time.Second * 3, stop: make(chan bool, 1)}
		//存储采集对象
		collectMap = &sync.Map{}
		collectMap.Store(lt.Name, cBean)
		go cBean.cron()
	}
}

func collectDump() {
	for {
		select {
		case <-CollectReboot:
			collectMap.Range(func(key, value interface{}) bool {
				//key = lt.Name
				switch bean := value.(type) {
				case *Collection:
					bean.stop <- true
				default:
					return false
				}
				return false
			})
			collectStart()
		}
	}
}

func CollectKjData() {
	go collectDump()
	collectStart()
}

func (m *GameKjData) History(c *gin.Context) {
	var (
		err error
	)
	qb := &records.QueryBean{}
	if err = c.ShouldBindQuery(qb); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}
	data, _ := c.Get("OWN")
	ob := data.(*UserOwnCard)
	wa := fmt.Sprintf("gid=%d", ob.Gid)
	if qb.WhereParam != "" {
		qb.WhereParam = qb.WhereParam + "," + wa
	} else {
		qb.WhereParam = wa
	}
	res := records.NewBeanRecords([]*GameKjData{}, qb)
	if err = res.List(); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}
	GinReturnOk(c, res)
}
