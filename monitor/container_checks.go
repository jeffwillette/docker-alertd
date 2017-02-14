package monitor

import (
)

// When the CPU usage is checked, it should return a bool indicating whether or not
// an alert should be sent, as well as an error (if it exists)
func checkCpuUsage(cont *Container, alertdStats *AlertdStats) (bool, float64) {
	totalUsage := alertdStats.CPUStats.CPUUsage.TotalUsage
	preTotalUsage := alertdStats.PreCPUStats.CPUUsage.TotalUsage
	systemCPUUsage := alertdStats.CPUStats.SystemUsage
	preSystemCPUUsage := alertdStats.PreCPUStats.SystemUsage

	realUsage := (totalUsage - preTotalUsage) / (systemCPUUsage - preSystemCPUUsage) * 100

	return realUsage > float64(cont.MaxCpu), realUsage
}

// Checks the min pids setting and returns true if alerts should be sent
func checkMinPids(cont *Container, alertdStats *AlertdStats) (bool, uint64) {
	pids := alertdStats.PidsStats.Current
	return pids < uint64(cont.MinProcs), pids
}


// Checks the memory used by the container in MB
func checkMemory(cont *Container, alertdStats *AlertdStats) (bool, float64) {
	// Memory level in MB
	memUsage := alertdStats.MemoryStats.Usage / 1000000
	return memUsage > float64(cont.MaxMem), float64(memUsage)
}
