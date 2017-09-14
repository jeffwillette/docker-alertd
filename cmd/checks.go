package cmd

import (
	"strings"

	"github.com/docker/docker/api/types"
)

// MetricCheck stores the name of the alert, a function, and a active boolean
type MetricCheck struct {
	Limit       uint64
	AlertActive bool
}

// StaticCheck checks the container for some static thing that is not based on usage
// statistics, like its existence, whether it is running or not, etc.
type StaticCheck struct {
	Expected    bool
	AlertActive bool
}

// Checker interface has all of the methods necessary to check a container
type Checker interface {
	CPUCheck(s *types.Stats)
	MemCheck(s *types.Stats)
	PIDCheck(s *types.Stats)
	ExistenceCheck(a *types.ContainerJSON, e error)
	RunningCheck(a *types.ContainerJSON, e error)
}

// AlertdContainer has the name of the container and the StaticChecks, and MetricChecks
// which are to be run on the container.
type AlertdContainer struct {
	Name     string `json:"name"`
	Alert    *Alert
	CPUCheck *MetricCheck
	MemCheck *MetricCheck
	PIDCheck *MetricCheck

	// static checks only below...
	ExistenceCheck *StaticCheck
	RunningCheck   *StaticCheck
}

// CheckMetrics checks everything where the Limit is not 0, there is no return because the
// checks modify the error in AlertdContainer
func (c *AlertdContainer) CheckMetrics(s *types.Stats, e error) {
	switch {
	case e != nil:
		c.Alert.Add("Received an unknown error: %s", e.Error())
	default:
		c.CheckCPUUsage(s)
		c.CheckMinPids(s)
		c.CheckMemory(s)
	}
}

// CheckStatics will run all of the static checks that are listed for a container.
func (c *AlertdContainer) CheckStatics(j *types.ContainerJSON, e error) {
	switch {
	case e != nil:
		c.CheckExists(e)
	default:
		c.CheckRunning(j)
	}
}

// IsUnknown is a check that takes the error when the docker API is polled, it
// mathes part of the error string that is returned.
func (c *AlertdContainer) IsUnknown(err error) bool {
	if err != nil {
		return strings.Contains(err.Error(), "No such container:")
	}
	return false
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
	return c.ExistenceCheck.AlertActive && e == nil
}

// CheckExists checks that the container exists, running or not
func (c *AlertdContainer) CheckExists(e error) {
	switch {
	case c.IsUnknown(e) && !c.ExistenceCheck.AlertActive:
		// if the alert is not active I need to alert and make it active
		c.Alert.Add("%s: failed existence check with error: %s", c.Name, e.Error())
		c.ExistenceCheck.AlertActive = true

	case c.IsUnknown(e) && c.ExistenceCheck.AlertActive:
		// do nothing
	case c.HasErrored(e):
		// if there is some other error besides an existence check error
		c.Alert.Add("%s: unknown error getting stats: %s", c.Name, e.Error())

	case c.HasBecomeKnown(e):
		c.Alert.Add("%s: existence check: recovered (exists)", c.Name)
		c.ExistenceCheck.AlertActive = false
	}
}

// ShouldAlertRunning returns whether the running state is as expected
func (c *AlertdContainer) ShouldAlertRunning(j *types.ContainerJSON) bool {
	// if they are not equal, return true (send alert)
	return c.RunningCheck.Expected != j.State.Running
}

// CheckRunning will check to see if the container is currently running or not
func (c *AlertdContainer) CheckRunning(j *types.ContainerJSON) {
	switch {
	case c.ShouldAlertRunning(j) && !c.RunningCheck.AlertActive:
		c.Alert.Add("%s: failed running state check (expected: %t): current running state: %t",
			c.Name, c.RunningCheck.Expected, j.State.Running)

		c.RunningCheck.AlertActive = true

	case !c.ShouldAlertRunning(j) && c.RunningCheck.AlertActive:
		c.Alert.Add("%s: running state check recovered (expected: %t): current running state: %t",
			c.Name, c.RunningCheck.Expected, j.State.Running)

		c.RunningCheck.AlertActive = false
	}
}

// RealCPUUsage calculates the CPU usage based on the ContainerJSON info
func (c *AlertdContainer) RealCPUUsage(s *types.Stats) uint64 {
	totalUsage := s.CPUStats.CPUUsage.TotalUsage
	preTotalUsage := s.PreCPUStats.CPUUsage.TotalUsage
	systemCPUUsage := s.CPUStats.SystemUsage
	preSystemCPUUsage := s.PreCPUStats.SystemUsage

	return (totalUsage - preTotalUsage) / (systemCPUUsage - preSystemCPUUsage) * 100
}

// ShouldAlertCPU returns true if the limit is breached
func (c *AlertdContainer) ShouldAlertCPU(u uint64) bool {
	return u > c.CPUCheck.Limit
}

// CheckCPUUsage takes care of sending the alerts if they are needed
func (c *AlertdContainer) CheckCPUUsage(s *types.Stats) {

	u := c.RealCPUUsage(s)
	a := c.ShouldAlertCPU(u)

	switch {
	case c.CPUCheck.Limit == 0:
		// do nothing because the check is disabled
	case a && !c.CPUCheck.AlertActive:
		c.Alert.Add("%s: exceed CPU alert (limit: %d): current usage: %d",
			c.Name, c.CPUCheck.Limit, u)

		c.CPUCheck.AlertActive = true

	case !a && c.CPUCheck.AlertActive:
		c.Alert.Add("%s: CPU level recovered (limit: %d): current usage %d",
			c.Name, c.CPUCheck.Limit, u)

		c.CPUCheck.AlertActive = false
	}
}

// ShouldAlertMinPIDS returns true if the minPID check fails
func (c *AlertdContainer) ShouldAlertMinPIDS(s *types.Stats) bool {
	return s.PidsStats.Current < c.PIDCheck.Limit
}

// CheckMinPids uses the min pids setting and check the number of PIDS in the container
// returns true if alerts should be sent, and also returns the amount of running pids.
func (c *AlertdContainer) CheckMinPids(s *types.Stats) {
	a := c.ShouldAlertMinPIDS(s)
	switch {
	case c.PIDCheck.Limit == 0:
		// do nothing because the check is disabled
	case a && !c.PIDCheck.AlertActive:
		c.Alert.Add("%s: failed Min PID check (minimum %d): current PIDs: %d",
			c.Name, c.PIDCheck.Limit, s.PidsStats.Current)

		c.PIDCheck.AlertActive = true

	case !a && c.PIDCheck.AlertActive:
		c.Alert.Add("%s: recovered Min PID check (minimum %d): current PIDs: %d",
			c.Name, c.PIDCheck.Limit, s.PidsStats.Current)

		c.PIDCheck.AlertActive = false
	}
}

// MemUsageMB returns the memory usage in MB
func (c *AlertdContainer) MemUsageMB(s *types.Stats) uint64 {
	return s.MemoryStats.Usage / 1000000
}

// ShouldAlertMemory returns whether the memory limit has been exceeded
func (c *AlertdContainer) ShouldAlertMemory(s *types.Stats) bool {
	// Memory level in MB
	u := c.MemUsageMB(s)
	return u > c.MemCheck.Limit
}

// CheckMemory checks the memory used by the container in MB, returns true if an
// error should be sent as well as the actual memory usage
func (c *AlertdContainer) CheckMemory(s *types.Stats) {

	u := c.MemUsageMB(s)
	a := c.ShouldAlertMemory(s)

	switch {
	case c.MemCheck.Limit == 0:
		// do nothing because the check is disabled
	case a && !c.MemCheck.AlertActive:
		c.Alert.Add("%s: exceeded memory usage limit (limit: %d): currently using: %d",
			c.Name, c.MemCheck.Limit, u)

		c.MemCheck.AlertActive = true

	case a && c.MemCheck.AlertActive:
		c.Alert.Add("%s: recovered memory usage limit (limit: %d): currently using: %d",
			c.Name, c.MemCheck.Limit, u)

		c.MemCheck.AlertActive = false
	}
}
