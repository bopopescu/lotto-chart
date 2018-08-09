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

	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		b, err := ioutil.ReadFile(pidFile)
		os.Remove(pidFile)
		if err != nil {
			log.Fatalln(err)
		}
		killCmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("kill -9 %s", string(b)))
		killCmd.Stderr = os.Stderr
		killCmd.Stdout = os.Stdout
		if err = killCmd.Run(); err != nil {
			log.Fatalln(err)
		}
		log.Println("服务停止成功")
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
