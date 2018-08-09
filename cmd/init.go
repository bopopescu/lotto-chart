// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"log"
	"net/http"

	"github.com/atcharles/lotto-chart/core/chart"
	"github.com/atcharles/lotto-chart/core/model"
	"github.com/atcharles/lotto-chart/core/orm"
	"github.com/levigross/grequests"
	"github.com/spf13/cobra"
)

var (
	push    = false
	pushUrl string
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if push {
			log.Println("程序上传...")
			up, err := grequests.FileUploadFromDisk(chart.RootDir + chart.ServerName)
			if err != nil {
				log.Fatalln(err)
			}
			up[0].FieldName = "file"
			up[0].FileName = chart.ServerName
			rp, err := grequests.Post(pushUrl, &grequests.RequestOptions{
				Files: up,
			})
			if err != nil {
				log.Fatalln(err)
			}
			defer rp.Close()
			type ResObj struct {
				Code int    `json:"code"`
				Msg  string `json:"msg"`
			}
			if rp.StatusCode != http.StatusOK {
				log.Fatalf("status: %d ; %s ", rp.StatusCode, rp.String())
			}
			bean := &ResObj{}
			if err = rp.JSON(bean); err != nil {
				log.Fatalln(err)
			}
			log.Println(bean.Msg)
			return
		}
		if err := orm.CreateDB(); err != nil {
			log.Fatalln("创建数据库失败:" + err.Error())
		}
		model.Initialize()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&push, "push", false, "")
	initCmd.Flags().StringVar(&pushUrl, "push_url", "", "")
}
