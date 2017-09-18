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
	"log"
	"os"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	// Config is the configuration object defined in conf-types file
	Config   Conf
	confName = ".docker-alertd"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "docker-alertd",
	Short: "docker-alertd: alert daemon for docker engine",
	Long: `docker-alerts parses a configuration file and then monitors containers through
the docker api.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {

		log.Println(Config)
		err := Config.Validate()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		Start(&Config)
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "",
		fmt.Sprintf("config file (default is ./%[1]s.yaml or $HOME/%[1]s.yaml)", confName))
	RootCmd.PersistentFlags().Int64P("iterations", "i", 0,
		"the number of iterations that the monitor will run. (default 0 is infinite)")
	RootCmd.PersistentFlags().Int64P("duration", "t", 1000,
		"the duration between monitor calls to the docker API in milliseconds (default 1000)")

	// Cobra also supports local flags, which will only run
	// Bind all the flags to viper for handling
	viper.BindPFlag("iterations", RootCmd.PersistentFlags().Lookup("iterations"))
	viper.BindPFlag("duration", RootCmd.PersistentFlags().Lookup("duration"))

	// when this action is called directly.
	//RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(confName) // name of config file (without extension)
	// add working dir path above the home dir for the config file
	path, _ := os.Getwd()
	viper.AddConfigPath(path)
	viper.AddConfigPath(os.Getenv("HOME")) // adding home directory as first search path
	viper.AutomaticEnv()                   // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	switch {
	case err != nil:
		log.Println(err)
	default:
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	err = viper.Unmarshal(&Config)
	if err != nil {
		log.Println(errors.Wrap(err, "unmarshal config file"))
	}
}
