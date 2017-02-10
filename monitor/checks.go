package monitor

import (
	"github.com/antonholmquist/jason"
)

// When the CPU usage is checked, it should return a bool indicating whether or not
// an alert should be sent, as well as an error (if it exists)
func checkCpuUsage(obj *jason.Object, cont Container) bool {
	total_usage, _ := obj.GetFloat64("cpu_stats", "cpu_usage", "total_usage")
	pre_total_usage, _ := obj.GetFloat64("precpu_stats", "cpu_usage", "total_usage")
	system_cpu_usage, _ := obj.GetFloat64("cpu_stats", "system_cpu_usage")
	pre_system_cpu_usage, _ := obj.GetFloat64("precpu_stats", "system_cpu_usage")

	realUsage := (total_usage - pre_total_usage) / (
		system_cpu_usage - pre_system_cpu_usage) * 100

	return realUsage > float64(cont.maxCpu)
}

// Checks the min pids setting and returns true if alerts should be sent
func checkMinPids(obj *jason.Object, cont Container) bool {
	pids, _ := obj.GetInt64("pids_stats", "current")
	return int64(cont.minProcs) < pids
}