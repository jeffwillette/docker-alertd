package cmd

import (
	"log"
	"strings"
)

// ContainerCheck stores the name of the alert, a function, and a active boolean
type ContainerCheck struct {
	Name        string
	Function    func(c *ContainerCheck, a *AlertdStats) (bool, float64)
	Limit       int64
	AlertActive bool
}

// AlertdContainer is made from the configuration file and stores the name of the container
// and all of the checks which will run on the container
type AlertdContainer struct {
	Name   string `json:"name"`
	Checks []ContainerCheck
}

// IsUnknown is a check that takes the error when the docker API is polled, it
// mathes part of the error string that is returned.
func (c *AlertdContainer) IsUnknown(err error) bool {
	return strings.Contains(err.Error(), "No such container:")
}

// HasErrored returns true when the error is something other than `isUnknownContainer`
// which means that docker-alertd probably crashed.
func (c *AlertdContainer) HasErrored(e error) bool {
	return e != nil && !c.IsUnknown(e)
}

// HasBecomeKnown returns true if there is an active alert and error is nil, which means
// that the container check was successful and the 0 index check (existence check) cannot
// be active
func (c *AlertdContainer) HasBecomeKnown(e error) bool {
	return c.Checks[0].AlertActive && e == nil
}

// CheckContainer loops through all of the Alerts in the struct and calls the function
// which checks them
func (c *AlertdContainer) CheckContainer(a *AlertdStats) *Alert {
	alert := Alert{Message: ""}
	for i, v := range c.Checks {
		shouldAlert, metric := v.Function(&v, a)

		switch {
		case v.Limit == 0:
			continue // if the limit is 0 then the check can be considered inactive
		case shouldAlert && !v.AlertActive:
			// If an alert comes back and the Alert is not currently active...
			alert.Add("%s: %s exceeded alert threshold of %d, it is currently using %f.\n",
				v.Name, c.Name, v.Limit, metric)

			c.Checks[i].AlertActive = true
			log.Print(alert.Message)
		case !shouldAlert && v.AlertActive:
			// If no alert comes back and the alert is active then it needs a recovered alert
			alert.Add("%s: %s recovered. threshold: %d, current: %f.\n", v.Name, c.Name,
				v.Limit, metric)

			c.Checks[i].AlertActive = false
			log.Print(alert.Message)
		}
	}
	return &alert
}

// NullCheck is for putting a "blank" check in the container checks so I can check things
// like existence which need to happen outside of the normal loop of checks on metrics
func NullCheck(c *ContainerCheck, a *AlertdStats) (bool, float64) {
	return false, 0
}

// CheckCPUUsage When the CPU usage is checked, it should return a bool indicating
// whether or not an alert should be sent, as well as an error (if it exists)
func CheckCPUUsage(c *ContainerCheck, a *AlertdStats) (bool, float64) {
	totalUsage := a.CPUStats.CPUUsage.TotalUsage
	preTotalUsage := a.PreCPUStats.CPUUsage.TotalUsage
	systemCPUUsage := a.CPUStats.SystemUsage
	preSystemCPUUsage := a.PreCPUStats.SystemUsage

	realUsage := (totalUsage - preTotalUsage) / (systemCPUUsage - preSystemCPUUsage) * 100

	return realUsage > float64(c.Limit), realUsage
}

// CheckMinPids uses the min pids setting and check the number of PIDS in the container
// returns true if alerts should be sent
func CheckMinPids(c *ContainerCheck, a *AlertdStats) (bool, float64) {
	pids := float64(a.PidsStats.Current)
	return pids < float64(c.Limit), pids
}

// CheckMemory checks the memory used by the container in MB, returns true if an
// error should be sent
func CheckMemory(c *ContainerCheck, a *AlertdStats) (bool, float64) {
	// Memory level in MB
	memUsage := a.MemoryStats.Usage / 1000000
	return memUsage > float64(c.Limit), float64(memUsage)
}
