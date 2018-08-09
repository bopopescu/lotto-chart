/*******************************************************************************
 * Copyright (c) 2018  charles
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NON INFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 * -------------------------------------------------------------------------
 * created at 2018-06-08 18:29:39
 ******************************************************************************/
package orm

import (
	"fmt"
	"log"
	"time"

	"github.com/atcharles/gof/goflogger"
	"github.com/atcharles/gof/gofutils"
	"github.com/atcharles/gof/gofutils/errors"
	"github.com/fsnotify/fsnotify"
	_ "github.com/go-sql-driver/mysql"
	xormcore "github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	"github.com/spf13/viper"
)

var (
	Engine      *xorm.EngineGroup
	configName  = "xorm_database.yaml"
	vp          = viper.New()
	defaultConf = Conf{
		UserCache: true,
		ShowSQL:   true,
		Master: Database{
			Type:     "mysql",
			User:     "root",
			Password: "111",
			DB:       "ty_chart",
			Address:  "127.0.0.1",
			Port:     3306,
		},
		Slave: Database{Address: "127.0.0.1", Port: 3306},
	}
)

type (
	//Conf  All Database configuration Settings
	Conf struct {
		UserCache bool
		ShowSQL   bool
		Master    Database
		Slave     Database `yaml:",flow"`
	}
	//Database  configuration Settings
	Database struct {
		Type     string `yaml:",omitempty"`
		Address  string
		Port     int
		User     string `yaml:",omitempty"`
		Password string `yaml:",omitempty"`
		DB       string `yaml:",omitempty"`
	}
)

func Initialize() {
	if err := settingDatabase(); err != nil {
		log.Fatalln(err.Error())
	}
}

//readConf ... Read configuration information
func readConf() error {
	fileName := gofutils.SelfDir() + "conf/" + configName
	if err := gofutils.TouchFile(fileName); err != nil {
		return err
	}
	vp.SetConfigFile(fileName)
	if err := vp.ReadInConfig(); err != nil {
		return err
	}
	ptr := &defaultConf
	key := gofutils.SnakeString(gofutils.ObjectName(ptr))
	if vp.IsSet(key) {
		return vp.UnmarshalKey(key, ptr)
	}
	vp.Set(key, ptr)
	if err := vp.WriteConfig(); err != nil {
		return err
	}

	vp.OnConfigChange(func(in fsnotify.Event) {
		vp.ReadInConfig()
		vp.UnmarshalKey(key, ptr)
	})

	return nil
}

//createDatabaseEngine  The database connection is created
func createDatabaseEngine() error {
	if err := readConf(); err != nil {
		return err
	}
	dbType := defaultConf.Master.Type
	switch dbType {
	case "mysql":
		sdn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?timeout=30s&charset=utf8mb4&parseTime=true",
			defaultConf.Master.User,
			defaultConf.Master.Password,
			defaultConf.Master.Address,
			defaultConf.Master.Port,
			defaultConf.Master.DB,
		)
		sdn2 := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?timeout=30s&charset=utf8mb4&parseTime=true",
			defaultConf.Master.User,
			defaultConf.Master.Password,
			defaultConf.Slave.Address,
			defaultConf.Slave.Port,
			defaultConf.Master.DB,
		)
		cons := []string{sdn, sdn2}
		db, err := xorm.NewEngineGroup(dbType, cons)
		if err != nil {
			return err
		}
		Engine = db
	default:
		return errors.Errorf("Unsupported database type:%s!", dbType)
	}
	return nil
}

func settingDatabase() error {
	if err := createDatabaseEngine(); err != nil {
		return err
	}

	Engine.SetMaxIdleConns(10)
	Engine.SetMaxOpenConns(100)
	Engine.SetConnMaxLifetime(60 * time.Second)

	Engine.SetMapper(xormcore.GonicMapper{})

	Engine.ShowSQL(defaultConf.ShowSQL)
	fName := gofutils.SelfDir() + "logs/sql/sql.log"
	if defaultConf.ShowSQL {
		Engine.SetLogger(xorm.NewSimpleLogger(goflogger.GetFile(fName).GetFile()))
		Engine.SetLogLevel(xormcore.LOG_INFO)
	} else {
		Engine.SetLogger(xorm.NewSimpleLogger(nil))
		Engine.SetLogLevel(xormcore.LOG_OFF)
	}

	Engine.SetDisableGlobalCache(!defaultConf.UserCache)
	if defaultConf.UserCache {
		cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000)
		Engine.SetDefaultCacher(cacher)
	}
	ping := func() {
		tk := time.NewTicker(30 * time.Second)
		to := time.NewTimer(10 * time.Second)
		for {
			to.Reset(10 * time.Second)
			select {
			case <-to.C:
				return
			case <-tk.C:
				Engine.Ping()
			}
		}
	}

	go ping()

	return nil
}

//创建数据库
func CreateDB() (err error) {
	var (
		eg      *xorm.Engine
		pwd     string
		confPtr = &defaultConf
	)

	if err := readConf(); err != nil {
		return err
	}
	fmt.Println("请填写数据库密码:")
	fmt.Scanln(&pwd)
	confPtr.Master.Password = pwd
	vp.Set("conf", confPtr)
	vp.WriteConfig()

	sdn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?timeout=30s&charset=utf8mb4&parseTime=true",
		confPtr.Master.User,
		confPtr.Master.Password,
		confPtr.Master.Address,
		confPtr.Master.Port,
	)
	eg, err = xorm.NewEngine("mysql", sdn)
	if err != nil {
		return
	}
	createDBSql := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS %s default character set utf8mb4 COLLATE utf8mb4_unicode_ci;",
		confPtr.Master.DB,
	)
	if _, err = eg.Exec(createDBSql); err != nil {
		return
	}
	defer eg.Close()
	return
}
