package monitor

import (
	"fmt"
	"testing"
)

func TestCheckCpuUsage(t *testing.T) {
	// test_json comes from the json_conf_test.go file
	j := []byte(json_data)
	alertStats, err := UnmarshalTestJson(&j)
	if err != nil {
		fmt.Printf("error unmarshaling JSON: %s", err)
	}

	// creating a container with 10% max cpu should pass the test
	container := Container{"langalang", 10, 0, 0}
	// If true is returns, that means an alert is raised (wrong)
	if test, _ := checkCpuUsage(&container, alertStats); test {
		t.Fail()
	}
}

func TestCheckCpuUsageFail(t *testing.T) {
	j := []byte(test_json)
	alertdStats, err := UnmarshalTestJson(&j)
	if err != nil {
		fmt.Printf("error unmarshaling JSON: %s", err)
	}

	// Creating a container with 0% max CPU usage should cause an alert (true)
	container := Container{"langalang", 0, 0, 0}

	// If false is returned that means alets was not raised (wrong)
	if test, _ := checkCpuUsage(&container, alertdStats); !test {
		t.Fail()
	}
}

func TestCheckMinPids(t *testing.T) {
	j := []byte(test_json)
	alertdStats, err := UnmarshalTestJson(&j)
	if err != nil {
		fmt.Printf("error unmarshaling JSON: %s", err)
	}

	// Creating a container with 3 PIDS should match and
	container := Container{"myContainer", 0, 0, 3}
	// If true is returned, that means an alert is raised (wrong)
	if test, _ := checkMinPids(&container, alertdStats); test {
		t.Fail()
	}
}

func TestCheckMinPidsFail(t *testing.T) {
	j := []byte(test_json)
	alertdStats, err := UnmarshalTestJson(&j)
	if err != nil {
		fmt.Printf("error unmarshaling JSON: %s", err)
	}

	// Creating a container with 0% max CPU usage should cause an alert (true)
	container := Container{"myContainer", 0, 0, 4}
	// If false, that means an alert is raised which (wrong)
	if test, _ := checkMinPids(&container, alertdStats); !test {
		t.Fail()
	}
}

func TestCheckMemory(t *testing.T) {
	j := []byte(test_json)
	alertdStats, err := UnmarshalTestJson(&j)
	if err != nil {
		fmt.Printf("error unmarshaling JSON: %s", err)
	}

	// Creating a container with a high memory limit should pass the test
	container := Container{"myContainer", 0, 1000, 0}
	// If false, that means an alert is raised which (wrong)
	if test, _ := checkMemory(&container, alertdStats); test {
		t.Fail()
	}
}

func TestCheckMemoryFail(t *testing.T) {
	j := []byte(test_json)
	alertdStats, err := UnmarshalTestJson(&j)
	if err != nil {
		fmt.Printf("error unmarshaling JSON: %s", err)
	}

	// Creating a container with a high memory limit should pass the test
	container := Container{"myContainer", 0, 0, 0}
	// If false, that means an alert is raised which (wrong)
	if test, _ := checkMemory(&container, alertdStats); !test {
		t.Fail()
	}
}
