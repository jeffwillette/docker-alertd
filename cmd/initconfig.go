// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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

	"github.com/spf13/cobra"
)

var dir string

// initconfigCmd represents the initconfig command
var initconfigCmd = &cobra.Command{
	Use:   "initconfig",
	Short: "generate a configuration file in the current directory",
	Long:  `generates a blank configuration file in the current directory`,
	Run: func(cmd *cobra.Command, args []string) {
		filename := fmt.Sprintf("%s/%s.yaml", dir, confName)

		// check to see if the file exists before writing the config file
		file, _ := os.Stat(filename)
		switch {
		case file != nil:
			log.Println("There is already a config file present, please move or delete" +
				" it before regenerating")
		default:
			ioutil.WriteFile(filename, config, 0644)
			log.Println("config successfully created, it has been filled with an example" +
				", which needs to be changed in order for docker-alertd to function.")
		}
	},
}

func init() {
	RootCmd.AddCommand(initconfigCmd)

	// gettng the current dir as the default directory value
	d, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	// Cobra supports local flags which will only run when this command
	initconfigCmd.Flags().StringVarP(&dir, "directory", "d", d, "directory to place config file in")

}

var config = []byte(`---
duration: 100				# duration in ms between docker API calls
iterations: 0				# number of iterations to run (0 = run forever)

# 'containers' is an array of dictionaries that each contain the name of a container to
# monitor, and the metrics which it should be monitored by. If there are no metrics
# present, then it will just be monitored to make sure that is is currently up.

# This will monitor only that the container exists, running or not...
# containers:
#   - name: mycontainer

containers:
  - name: container1
    expectedRunning: true

  - name: container2
    expectedRunning: true
    maxCpu: 20
    maxMem: 20
    minProcs: 4

## ALERTERS...
## If any of the below alerters are present, alerts will be sent through the proper 
## channels. Completely delete the relevant section to disable them. To Test if an alerter
## authenticates properly, run the "testalert" command

email:
  smtp: smtp.nonexistantserver.com
  password: s00p3rS33cret
  port: 587
  from: auto@freshpowpow.com
  subject: "DOCKER_ALERTD"
  to:
    - jeff@gnarfresh.com

# You need to start a slack channel and activate an app to get a webhookURL for your channel
# see https://api.slack.com/apps for more information
slack:
  webhookURL: https://some.url/provided/by/slack/

# You need to create a pushover account to use this
# see https://pushover.net for more information
pushover:
  ApiURL: https://some.url/
  ApiToken: your_api_token
  UserKey: your_user_key
`)
