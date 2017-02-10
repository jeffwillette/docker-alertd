package monitor

import (
	"fmt"
	"testing"
)

var json_data = `
{
	"containers": [
		{
			"name": "langalang",
			"max-cpu": 10,
			"max-mem": 20,
			"min-proc": 3
		}
	]
}
`

func TestGetContainersJSON(t *testing.T) {
	a := []byte(json_data)

	_, err := GetContainersJSON(&a)
	if err != nil {fmt.Println("ParseJSON function error")}
}


var bad_data = `
{
	"containers": []
}
`

func TestGetContainersJSONBadData(t *testing.T) {
	a := []byte(bad_data)

	_, err := GetContainersJSON(&a)
	// In this case there should be an error, making sure it exists
	if err == nil {
		fmt.Println("ParseJSON function error")
		t.Fail()
	}
}

var multiple_containers = `
{
	"containers": [
		{
			"name": "langalang",
			"max-cpu": 10,
			"max-mem": 20,
			"min-proc": 3
		},
		{
			"name": "postgres",
			"max-cpu": 10,
			"max-mem": 20,
			"min-proc": 3
		}
	]
}
`

func TestGetContainersJSONMultiple(t *testing.T) {
	a := []byte(multiple_containers)

	_, err := GetContainersJSON(&a)
	// In this case there should be an error, making sure it exists
	if err != nil {
		fmt.Println("ParseJSON function error")
		t.Fail()
	}
}

var email_json = `
{
	"email": {
		"from": "address",
		"to": "addressto",
		"smtp": "server",
		"port": 587,
		"password": "secret"
	}
}
`

func TestGetEmailJSON(t *testing.T) {
	a := []byte(email_json)
	email := GetEmailConfJSON(&a)
	test_email := Email{"address", "addressto", "server", 587, "secret"}
	if test_email != *email {
		t.Fail()
	}
}