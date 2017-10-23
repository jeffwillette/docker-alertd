package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	dir          string
	stdout       bool
	alerterStubs = map[string]*AlerterStub{
		"email": &AlerterStub{
			ShouldPrint: false,
			Bytes:       email,
		},
		"slack": &AlerterStub{
			ShouldPrint: false,
			Bytes:       slack,
		},
		"pushover": &AlerterStub{
			ShouldPrint: false,
			Bytes:       pushover,
		},
	}
)

// AlerterStub stores the bytes needed for the alerter stub in the config file as well as
// a boolean value which is evaluated only if the user has supplied specific alerter stubs
// to include
type AlerterStub struct {
	ShouldPrint bool
	Bytes       []byte
}

// initconfigCmd represents the initconfig command
var initconfigCmd = &cobra.Command{
	Use:   "initconfig",
	Short: "generate a configuration file in the current directory",
	Long:  `generates a blank configuration file in the current directory`,
	Run: func(cmd *cobra.Command, args []string) {
		filename := fmt.Sprintf("%s/%s.yaml", dir, confName)

		buffer := bytes.Buffer{}
		_, err := buffer.Write(config)
		if err != nil {
			log.Println(err)
		}

		switch {
		case !shouldPrintall():
			// this case only prints the alerters requested
			for k, v := range alerterStubs {
				if v.ShouldPrint {
					_, err := buffer.Write(v.Bytes)
					if err != nil {
						log.Println(err)
					}
					log.Printf("added %s alerter stub", k)
				}
			}
		default:
			// default is to print all the alerters stubs to the config file
			for _, v := range alerterStubs {
				_, err := buffer.Write(v.Bytes)
				if err != nil {
					log.Println(err)
				}
			}
		}

		// check to see if the file exists before writing the config file
		file, _ := os.Stat(filename)
		switch {
		case file != nil:
			log.Println("There is already a config file present, please move or delete" +
				" it before regenerating")
		case stdout:
			fmt.Println(buffer.String())
		default:
			ioutil.WriteFile(filename, buffer.Bytes(), 0644)

			log.Println("config successfully created, it has been filled with an example" +
				", which includes all possible alerters. The alerter stubs need to be ch" +
				"anged in order for docker-alertd to function.")
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
	initconfigCmd.Flags().BoolVar(&alerterStubs["email"].ShouldPrint, "email", false, "include email alert stub")
	initconfigCmd.Flags().BoolVar(&alerterStubs["slack"].ShouldPrint, "slack", false, "include slack alert stub")
	initconfigCmd.Flags().BoolVar(&alerterStubs["pushover"].ShouldPrint, "pushover", false,
		"include pushover alert stub")
	initconfigCmd.Flags().BoolVar(&stdout, "stdout", false, "print config to stdout")

}

// shouldPrintall returns true if all of the alerter stubs should be printed to the config
// file on initconfig command.
func shouldPrintall() bool {
	switch {
	case alerterStubs["email"].ShouldPrint:
		return false
	case alerterStubs["slack"].ShouldPrint:
		return false
	case alerterStubs["pushover"].ShouldPrint:
		return false
	default:
		return true
	}
}

var config = []byte(`---
# The duration and interations settings, if omitted, have a default value of 100ms between
# docker API calls and an indefinite number of iterations which will run the monitor forever
#duration: 100				# duration in ms between docker API calls
#iterations: 0				# number of iterations to run

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
`)

var email = []byte(`
email:
  smtp: smtp.nonexistantserver.com
  password: s00p3rS33cret
  port: 587
  from: auto@freshpowpow.com
  subject: "DOCKER_ALERTD"
  to:
    - jeff@gnarfresh.com
`)

var slack = []byte(`
# You need to start a slack channel and activate an app to get a webhookURL for your channel
# see https://api.slack.com/apps for more information
slack:
  webhookURL: https://some.url/provided/by/slack/
`)

var pushover = []byte(`
# You need to create a pushover account to use this
# see https://pushover.net for more information
pushover:
  ApiURL: https://some.url/
  ApiToken: your_api_token
  UserKey: your_user_key
`)
