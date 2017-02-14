package monitor

import (
	"fmt"
	"reflect"
	"testing"
)

var json_data = `
{
	"containers": [
		{
			"name": "container1",
			"max-cpu": 0,
			"max-mem": 20,
			"min-procs": 3
		}
	],
	"email_settings": {
		"from": "auto@freshpowpow.com",
		"to": "jeff@gnarfresh.com",
		"smtp": "smtp.coolserver.com",
		"password": "gnarlesbarkely",
		"port": 587
	}
}
`

func TestGetContainersJSON(t *testing.T) {
	a := []byte(json_data)

	// get the configuration and if there is an error, fail the test
	configuration, err := GetConfJSON(&a)
	if err != nil {
		fmt.Println("Error parsing JSON")
		t.Fail()
	}

	// creating a test struct to see if it matches the one made from JSON
	testContainer := Container{"container1", 0, 20, 3}

	testEmailSettings := EmailSettings{
		"auto@freshpowpow.com", "jeff@gnarfresh.com", "smtp.coolserver.com",
		"gnarlesbarkely", 587}

	testConf := Conf{
		[]Container{testContainer}, testEmailSettings}

	// If they are not equal, it needs to fail
	if !reflect.DeepEqual(testConf, *configuration) {
		fmt.Println("Parsed config does not match test config")
		t.Fail()
	}
}

var bad_data = `
{
	"containers": [],
	"email_settings": {}
}
`

func TestGetContainersJSONBadData(t *testing.T) {
	a := []byte(bad_data)
	_, err := GetConfJSON(&a)

	// The configuration is bad so there should be an error
	if err == nil {
		t.Fail()
	}
}

var multiple_containers = `
{
	"containers": [
		{
			"name": "container1",
			"max-cpu": 0,
			"max-mem": 20,
			"min-procs": 3
		},
		{
			"name": "container2",
			"max-cpu": 20,
			"max-mem": 20,
			"min-procs": 4
		}
	],
	"email_settings": {
		"from": "auto@freshpowpow.com",
		"to": "jeff@gnarfresh.com",
		"smtp": "smtp.coolserver.com",
		"password": "gnarlesbarkely",
		"port": 587
	}
}
`

func TestGetContainersJSONMultiple(t *testing.T) {
	a := []byte(multiple_containers)

	configuration, err := GetConfJSON(&a)
	if err != nil {
		fmt.Println("Error parsing JSON")
		t.Fail()
	}

	// creating a test struct to see if it matches the one made from JSON
	testContainer := Container{"container1", 0, 20, 3}
	testContainer2 := Container{"container2", 20, 20, 4}

	testEmailSettings := EmailSettings{
		"auto@freshpowpow.com", "jeff@gnarfresh.com", "smtp.coolserver.com",
		"gnarlesbarkely", 587}

	testConf := Conf{
		[]Container{testContainer, testContainer2},
		testEmailSettings}

	if !reflect.DeepEqual(testConf, *configuration) {
		fmt.Println("Configuration not equal to testConfig")
		t.Fail()
	}
}
