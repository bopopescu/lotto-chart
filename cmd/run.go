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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/atcharles/gof/goflogger"
	"github.com/atcharles/gof/gofutils"
	"github.com/atcharles/lotto-chart/core"
	"github.com/atcharles/lotto-chart/core/chart"
	"github.com/spf13/cobra"
)

var (
	pidFile    = chart.RootDir + "pid"
	daemon     = false
	outPutFile = chart.RootDir + "output.log"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if !daemon {
			if gofutils.FileExists(pidFile) {
				log.Fatalln("程序已经启动,请查看pid文件是否存在,若服务未运行,请删除pid文件重新运行")
			}
			os.RemoveAll(chart.RootDir + "logs")
			os.Remove(outPutFile)
			outFile := goflogger.GetFile(outPutFile).GetFile()
			cmdRun := exec.Command("/bin/bash", "-c", fmt.Sprintf(`%s run -d`, os.Args[0]))
			cmdRun.Stdout = outFile
			cmdRun.Stderr = outFile
			if err := cmdRun.Start(); err != nil {
				log.Fatalln(err)
			}
			log.Println("服务启动成功")
			return
		}

		pid := os.Getpid()
		log.Printf("The current process PID is %d", pid)
		ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), os.ModePerm)
		core.GinServer()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVarP(&daemon, "daemon", "d", false, "守护进程")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
