package monitor

import (
	"errors"
	//"net/smtp"
	"log"
	"github.com/antonholmquist/jason"
)


func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}


type Container struct {
	name string
	maxCpu int64
	maxMem int64
	minProcs int64
}


func GetContainersJSON(d *[]byte) (c *[]Container, err error){

	conf, err := jason.NewObjectFromBytes(*d)
	check(err)

    containers := []Container{}
    contArray, err := conf.GetObjectArray("containers")
    check(err)
    for _, c := range contArray {
    	name, err := c.GetString("name")
    	check(err)
    	maxCpu, err := c.GetInt64("max-cpu")
    	check(err)
    	maxMem, err := c.GetInt64("max-mem")
    	check(err)
    	minProcs, err := c.GetInt64("min-procs")
    	check(err)
    	containers = append(containers,
    		Container{name, maxCpu, maxMem, minProcs})
    }

    if len(containers) < 1 {
    	err = errors.New("There are no containers to monitor")
    }

	return &containers, err
}


// If params are added here, they also need to be added in the ParseYAML functions
type Email struct {
	from string
	to string
	smtp string
	port int64
	password string
}

func GetEmailConfJSON(d *[]byte) (email *Email) {
	obj, err := jason.NewObjectFromBytes(*d)
	check(err)

	emailObj, err := obj.GetObject("email")
	check(err)

	from, err := emailObj.GetString("from")
	check(err)
	to, err := emailObj.GetString("to")
	check(err)
	smtp, err := emailObj.GetString("smtp")
	check(err)
	port, err := emailObj.GetInt64("port")
	check(err)
	password, err := emailObj.GetString("password")
	check(err)

	return &Email{from, to, smtp, port, password}
}