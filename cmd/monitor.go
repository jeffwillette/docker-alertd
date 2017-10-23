package cmd

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func uint64P(u uint64) *uint64 {
	p := u
	return &p
}

// intP returns a pointer to an int
func int64P(i int64) *int64 {
	p := i
	return &p
}

// boolP returns a pointer to a bool
func boolP(b bool) *bool {
	p := b
	return &p
}

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
	// Taking the values from the conf and adding them into the AlertdContainers
	var containers []AlertdContainer
	for _, v := range c.Containers {
		containers = append(containers, AlertdContainer{
			Name: v.Name,
			Alert: &Alert{
				Messages: []error{},
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
				Expected:    boolP(true),
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

// CheckContainers goes through and checks all the containers in a loop
func CheckContainers(cnt []AlertdContainer, cli *client.Client, a *Alert) {
	for _, c := range cnt {
		// make sure we have a clean alert for this loop
		c.Alert.Clear()

		// handling whether the container exists, if these checks fail, the checking
		// process should stop
		j, err := ContainerInspect(&c, cli)
		c.CheckStatics(j, err)

		// if an alert should be sent that means it either failed existence or running
		// checks which means that nothing more can be checked
		if c.ChecksShouldStop() {
			a.Concat(c.Alert) // add the alert in the container to the main alert
			continue
		}

		s, err := GetStats(&c, cli)
		c.CheckMetrics(s, err)

		if c.Alert.ShouldSend() {
			a.Concat(c.Alert)
		}
	}
}

// Monitor contains all the calls for the main loop of the monitor
func Monitor(c *Conf, a *Alert) {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}

	cnt := InitCheckers(c)

	if c.Duration == nil {
		c.Duration = int64P(100)
	}

	switch c.Iterations {
	case nil:
		for {
			a.Clear()
			CheckContainers(cnt, cli, a)
			a.Evaluate()
			time.Sleep(time.Duration(*c.Duration) * time.Millisecond)
		}
	default:
		for i := int64(0); i < *c.Iterations; i++ {
			a.Clear()
			CheckContainers(cnt, cli, a)
			a.Evaluate()
			time.Sleep(time.Duration(*c.Duration) * time.Millisecond)
		}
	}
}

// Start the main monitor loop for a set amount of iterations
func Start(c *Conf) {
	log.Printf("starting docker-alertd\n------------------------------")
	a := &Alert{Messages: []error{}}
	Monitor(c, a)
}
