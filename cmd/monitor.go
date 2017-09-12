package cmd

import (
	"context"
	"encoding/json"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"io/ioutil"
)

// AlertdStats from the types imported from docker. These include the data that needs to be
// extracted from the docker API
type AlertdStats struct {
	CPUStats    CPUStats    `json:"cpu_stats"`
	PreCPUStats CPUStats    `json:"precpu_stats"`
	MemoryStats MemoryStats `json:"memory_stats"`
	// this one is not redefined because JSON marshal does not convert them
	// into float64's
	PidsStats types.PidsStats `json:"pids_stats"`
}

// UnmarshalStats takes the ContainerStats returned from the docker API and parses the JSON into
// a ContainerStats struct
func UnmarshalStats(c types.ContainerStats) *AlertdStats {
	b, err := ioutil.ReadAll(c.Body)
	if err != nil {
		log.Fatal("Error reading the stats: ", err)
	}

	var alertdStats AlertdStats
	err = json.Unmarshal(b, &alertdStats)
	if err != nil {
		log.Fatalf("Error unmarshaling JSON: %s", err)
	}

	return &alertdStats
}

// getStats just uses the docker API and an already tested Unmarshal function, no
// testing needed.
func getStats(a *AlertdContainer, c *client.Client) (*AlertdStats, error) {
	rawStats, err := c.ContainerStats(context.Background(), a.Name, false)
	if err != nil {
		return nil, err
	}

	defer rawStats.Body.Close()
	return UnmarshalStats(rawStats), nil
}

// NewAlertdContainer returns a slice of containers with all the info needed to run a
// check on the container. Active is for whether or not the alert is active, not the check
func NewAlertdContainer(c *Conf) *[]AlertdContainer {
	// Taking the Conf and changing it into a more appropriate format for the monitor
	var alertdCont []AlertdContainer
	for _, v := range c.Containers {
		alertdCont = append(alertdCont, AlertdContainer{
			Name: v.Name,
			Checks: []ContainerCheck{
				ContainerCheck{
					Name:        "Container Existence Alert",
					Function:    NullCheck,
					Limit:       0,
					AlertActive: false,
				},
				ContainerCheck{
					Name:        "CPU Usage Alert",
					Function:    CheckCPUUsage,
					Limit:       v.MaxCPU,
					AlertActive: false,
				},
				ContainerCheck{
					Name:        "Memory Usage Alert",
					Function:    CheckMemory,
					Limit:       v.MaxMem,
					AlertActive: false,
				},
				ContainerCheck{
					Name:        "Minimum Processes Alert",
					Function:    CheckMinPids,
					Limit:       v.MinProcs,
					AlertActive: false,
				},
			},
		})
	}
	return &alertdCont
}

// Start the main loop should continously run forever
func Start(c *Conf) {
	log.Printf("started docker-alertd process\n------------------------------")

	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}

	alertdConts := NewAlertdContainer(c)

	for {
		alert := &Alert{Message: ""}
		for i, container := range *alertdConts {
			alertdStats, err := getStats(&container, cli)
			switch {
			case container.IsUnknown(err):
				// if the alert is not active I need to alert and make it active
				if !container.Checks[0].AlertActive {
					alert.Add("%s: checking '%s' gave the error: %s\n",
						container.Checks[0].Name, container.Name, err.Error())

					container.Checks[0].AlertActive = true
				}
				continue // If it is unknown and alert is active, nothing left to do
			case container.HasErrored(err):
				alert.Add("%s: Error getting stats for '%s': %s\n", container.Checks[i].Name,
					container.Name, err.Error())

			case container.HasBecomeKnown(err):
				alert.Add("%s: '%s' has recovered", container.Checks[0].Name, container.Name)
				container.Checks[0].AlertActive = false

			default:
				a := container.CheckContainer(alertdStats)
				alert.Add("%s", a.Message)
			}
		}
		alert.Evaluate()
	}
}
