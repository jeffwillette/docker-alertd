package monitor

import (
	"encoding/json"
	"errors"
	"log"
)

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

// Parse the configuration file and exit if there are any errors or if there
// are no containers in the configuration file.
func GetConfJSON(j *[]byte) (c *Conf, err error) {
	var conf Conf
	err = json.Unmarshal(*j, &conf)
	if err != nil {
		return &conf, err
	}

	if len(conf.Containers) < 1 {
		err := errors.New(
			"There were no containers found in the configuration file")
		return &conf, err
	}

	return &conf, err
}
