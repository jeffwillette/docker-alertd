package main

import (
	"fmt"
	"flag"
	"io/ioutil"
	"log"

	"github.com/deltaskelta/docker-alertd/monitor"
)

func main() {
	// Defining the nexessary file input flag, returns
	fileArg := flag.String(
		"f", "nil", "Usage: required configuration .yaml file")

	flag.Parse()

	switch {
	case *fileArg != "nil":
		// Parse the YAML file, and start monitor is there are no errors
		fileData, err := ioutil.ReadFile(*fileArg)
		if err != nil {
			log.Println("There was a problem reading the configuration file")
			log.Fatal(err)
		}

		email := monitor.GetEmailConfJSON(&fileData)

		conf, err := monitor.GetContainersJSON(&fileData)
		if err != nil {
			log.Fatal(err)
		}

		monitor.Start(conf, email)
	default:
		fmt.Printf("  Define a yaml configuration\n\n")
		flag.PrintDefaults()
	}
}
