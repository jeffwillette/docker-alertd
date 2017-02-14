package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/deltaskelta/docker-alertd/monitor"
)

func main() {
	// Defining the nexessary file input flag, returns
	fileArg := flag.String(
		"f", "nil", "Usage: required configuration .json file")

	flag.Parse()

	switch {
	case *fileArg == "nil":
		fmt.Printf("  Define a JSON configuration file\n\n")
		flag.PrintDefaults()
	default:
		// Read the JSON configuration file
		fileData, err := ioutil.ReadFile(*fileArg)
		if err != nil {
			log.Fatalf("Error opening the configuration file: %s", err)
		}

		// Parse the configuration, returning a Conf object pointer
		conf, err := monitor.GetConfJSON(&fileData)
		if err != nil {
			log.Fatal(err)
		}

		monitor.Start(conf)
	}
}
