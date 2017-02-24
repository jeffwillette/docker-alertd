package monitor

import (
	"encoding/json"
	"fmt"
	"log"

	"io/ioutil"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"golang.org/x/net/context"
)

/*
	TODO:
		- add a subject line in the conf JSON for the email
*/

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
func getStats(c *MondContainer, cli *client.Client) (*AlertdStats, error) {
	rawStats, err := cli.ContainerStats(context.Background(), c.Name, false)
	if err != nil {
		return &AlertdStats{}, err
	}

	defer rawStats.Body.Close()
	return UnmarshalStats(rawStats), nil
}

// NewMondContainer takes in the simple configuration file and returns a slice of
// MondContainers, this is necessary in order to keep the configuration file simple
// and short and also have the Alerts store their states
func NewMondContainer(c *Conf) *[]MondContainer {
	// Taking the Conf and changing it into a more appropriate format for the monitor
	var mondCont []MondContainer
	for _, v := range c.Containers {
		mondCont = append(mondCont, MondContainer{
			Name: v.Name,
			Alerts: []ContainerAlert{
				ContainerAlert{
					Name:     "CPU Usage Alert",
					Function: CheckCPUUsage,
					Limit:    v.MaxCPU,
					Active:   false,
				},
				ContainerAlert{
					Name:     "Memory Usage Alert",
					Function: CheckMemory,
					Limit:    v.MaxMem,
					Active:   false,
				},
				ContainerAlert{
					Name:     "Minimum Processes Alert",
					Function: CheckMinPids,
					Limit:    v.MinProcs,
					Active:   false,
				},
			},
		})
	}
	return &mondCont
}

// Start the main loop should continously run forever
func Start(c *Conf) {
	log.Printf("started docker-alertd process\n------------------------------")

	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}

	mondConts := NewMondContainer(c)

	for {
		var emailMessage string
		for _, container := range *mondConts {
			// checkContainer checks against all of the alerts.
			alertdStats, err := getStats(&container, cli)
			if err != nil {
				email := &Email{c.Email.From, c.Email.To, c.Email.Subject,
					[]byte("Docker-Alertd Crashed trying to get container stats " +
						"check the status of your docker install"),
				}
				c.Emailer.Send(email)
				log.Fatal("Error getting stats from docker: ", err)
			}
			a := container.CheckContainer(alertdStats)
			// Concatenating all alert messages into one string
			for _, alert := range a {
				emailMessage += fmt.Sprintf("%s", alert)
			}
		}
		if len(emailMessage) > 0 {
			fMsg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n",
				c.Email.To, c.Email.Subject, emailMessage))
			email := &Email{c.Email.From, c.Email.To, c.Email.Subject, fMsg}

			go func() {
				err := c.Emailer.Send(email)
				if err != nil {
					log.Fatal(err)
				}
				log.Println("alert email sent")
			}()
		}
	}
}
