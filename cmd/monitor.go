package cmd

import (
	"context"
	"encoding/json"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// GetStats just uses the docker API and an already tested Unmarshal function, no
// testing needed.
func GetStats(a *AlertdContainer, c *client.Client) (*types.Stats, error) {
	cs, err := c.ContainerStats(context.Background(), a.Name, false)
	if err != nil {
		return nil, err
	}
	defer cs.Body.Close()

	d := json.NewDecoder(cs.Body)
	d.UseNumber()

	var stats types.Stats
	if err := d.Decode(&stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

// ContainerInspect returns the information which can decide if the container is current;y running
// or not.
func ContainerInspect(a *AlertdContainer, c *client.Client) (*types.ContainerJSON, error) {
	containerJSON, err := c.ContainerInspect(context.Background(), a.Name)
	if err != nil {
		return nil, err
	}

	return &containerJSON, nil
}

// InitCheckers returns a slice of containers with all the info needed to run a
// check on the container. Active is for whether or not the alert is active, not the check
func InitCheckers(c *Conf) []AlertdContainer {
	// Taking the Conf and changing it into a more appropriate format for the monitor
	var containers []AlertdContainer
	for _, v := range c.Containers {
		containers = append(containers, AlertdContainer{
			Name: v.Name,
			Alert: &Alert{
				Message: "",
			},
			CPUCheck: &MetricCheck{
				Limit:       v.MaxCPU,
				AlertActive: false,
			},
			MemCheck: &MetricCheck{
				Limit:       v.MaxMem,
				AlertActive: false,
			},
			PIDCheck: &MetricCheck{
				Limit:       v.MinProcs,
				AlertActive: false,
			},
			ExistenceCheck: &StaticCheck{
				Expected:    true,
				AlertActive: false,
			},
			RunningCheck: &StaticCheck{
				Expected:    v.ExpectedRunning,
				AlertActive: false,
			},
		})
	}
	return containers
}

// Start the main loop should continously run forever
func Start(c *Conf) {
	log.Printf("started docker-alertd process\n------------------------------")

	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}

	containers := InitCheckers(c)

	for {
		alert := &Alert{Message: ""}
		for _, c := range containers {
			// make sure to reset the alert message on every loop
			c.Alert.Message = ""

			// this check should take care of checking whether or not the conainer exists,
			// so the error handling in the next one should be just to default to sending
			// an alert.
			j, err := ContainerInspect(&c, cli)
			c.CheckStatics(j, err)

			// if an alert should be sent that means it either failed existence or running
			// checks which means that nothing more can be checked
			if c.Alert.ShouldSend() || !c.RunningCheck.Expected || c.RunningCheck.AlertActive || c.ExistenceCheck.AlertActive {
				alert.Concat(c.Alert) // add the alert in the container to the main alert
				continue
			}

			s, err := GetStats(&c, cli)
			c.CheckMetrics(s, err)

			if c.Alert.ShouldSend() {
				alert.Concat(c.Alert)
			}
		}
		alert.Evaluate()
	}
}
