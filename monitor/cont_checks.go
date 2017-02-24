package monitor

import (
	"fmt"
	"log"
)

// ContainerAlert stores the name of the alert, a function, and a active boolean
type ContainerAlert struct {
	Name     string
	Function func(ca *ContainerAlert, a *AlertdStats) (bool, float64)
	Limit    int64
	Active   bool
}

// MondContainer is made from the configuration file and stores the name of the container
// and all of the checks which will run on the container
type MondContainer struct {
	Name   string `json:"name"`
	Alerts []ContainerAlert
}

// CheckContainer loops through all of the Alerts in the struct and calls the function
// which checks them
func (md *MondContainer) CheckContainer(a *AlertdStats) []string {
	var alerts []string
	for i, v := range md.Alerts {
		alert, metric := v.Function(&v, a)
		if alert && !v.Active {
			// If an alert comes back and the Alert is not currently active...
			alertMessage := fmt.Sprintf("%s: %s exceeded alert threshold of %d, it is "+
				"currently using %f.\n", v.Name, md.Name, v.Limit, metric)

			alerts = append(alerts, alertMessage)
			md.Alerts[i].Active = true
			log.Printf(alertMessage)
		} else if !alert && v.Active {
			// If no alert comes back and the alert is active then it needs a recovered
			// alert
			alertMessage := fmt.Sprintf("%s: %s recovered. threshold: %d, current: %f.\n",
				v.Name, md.Name, v.Limit, metric)

			alerts = append(alerts, alertMessage)
			md.Alerts[i].Active = false
			log.Printf(alertMessage)
		}
	}
	return alerts
}

// CheckCPUUsage When the CPU usage is checked, it should return a bool indicating
// whether or not an alert should be sent, as well as an error (if it exists)
func CheckCPUUsage(ca *ContainerAlert, alertdStats *AlertdStats) (bool, float64) {
	totalUsage := alertdStats.CPUStats.CPUUsage.TotalUsage
	preTotalUsage := alertdStats.PreCPUStats.CPUUsage.TotalUsage
	systemCPUUsage := alertdStats.CPUStats.SystemUsage
	preSystemCPUUsage := alertdStats.PreCPUStats.SystemUsage

	realUsage := (totalUsage - preTotalUsage) / (systemCPUUsage - preSystemCPUUsage) * 100

	return realUsage > float64(ca.Limit), realUsage
}

// CheckMinPids uses the min pids setting and check the number of PIDS in the container
// returns true if alerts should be sent
func CheckMinPids(ca *ContainerAlert, alertdStats *AlertdStats) (bool, float64) {
	pids := float64(alertdStats.PidsStats.Current)
	return pids < float64(ca.Limit), pids
}

// CheckMemory checks the memory used by the container in MB, returns true if an
// error should be sent
func CheckMemory(ca *ContainerAlert, alertdStats *AlertdStats) (bool, float64) {
	// Memory level in MB
	memUsage := alertdStats.MemoryStats.Usage / 1000000
	return memUsage > float64(ca.Limit), float64(memUsage)
}
