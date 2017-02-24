package monitor

// These are taken from the docker API, and uint64's are redefined into
//float64 because they get converted to scientific notation during the
// marshaling process.

// CPUUsage stores All CPU stats aggregated since container inception.
type CPUUsage struct {
	// Total CPU time consumed.
	// Units: nanoseconds (Linux)
	// Units: 100's of nanoseconds (Windows)
	TotalUsage float64 `json:"total_usage"`

	// Total CPU time consumed per core (Linux). Not used on Windows.
	// Units: nanoseconds.
	PercpuUsage []float64 `json:"percpu_usage,omitempty"`

	// Time spent by tasks of the cgroup in kernel mode (Linux).
	// Time spent by all container processes in kernel mode (Windows).
	// Units: nanoseconds (Linux).
	// Units: 100's of nanoseconds (Windows). Not populated for Hyper-V Containers.
	UsageInKernelmode float64 `json:"usage_in_kernelmode"`

	// Time spent by tasks of the cgroup in user mode (Linux).
	// Time spent by all container processes in user mode (Windows).
	// Units: nanoseconds (Linux).
	// Units: 100's of nanoseconds (Windows). Not populated for Hyper-V Containers
	UsageInUsermode float64 `json:"usage_in_usermode"`
}

// ThrottlingData stores CPU throttling stats of one running container.
// Not used on Windows.
type ThrottlingData struct {
	// Number of periods with throttling active
	Periods float64 `json:"periods"`
	// Number of periods when the container hits its throttling limit.
	ThrottledPeriods float64 `json:"throttled_periods"`
	// Aggregate time the container was throttled for in nanoseconds.
	ThrottledTime float64 `json:"throttled_time"`
}

// CPUStats aggregates and wraps all CPU related info of container
type CPUStats struct {
	// CPU Usage. Linux and Windows.
	CPUUsage CPUUsage `json:"cpu_usage"`

	// System Usage. Linux only.
	SystemUsage float64 `json:"system_cpu_usage,omitempty"`

	// Throttling Data. Linux only.
	ThrottlingData ThrottlingData `json:"throttling_data,omitempty"`
}

// MemoryStats aggregates all memory stats since container inception on Linux.
// Windows returns stats for commit and private working set only.
type MemoryStats struct {
	// Linux Memory Stats

	// current res_counter usage for memory
	Usage float64 `json:"usage,omitempty"`
	// maximum usage ever recorded.
	MaxUsage float64 `json:"max_usage,omitempty"`
	// TODO(vishh): Export these as stronger types.
	// all the stats exported via memory.stat.
	Stats map[string]float64 `json:"stats,omitempty"`
	// number of times memory usage hits limits.
	Failcnt float64 `json:"failcnt,omitempty"`
	Limit   float64 `json:"limit,omitempty"`

	// Windows Memory Stats
	// See https://technet.microsoft.com/en-us/magazine/ff382715.aspx

	// committed bytes
	Commit float64 `json:"commitbytes,omitempty"`
	// peak committed bytes
	CommitPeak float64 `json:"commitpeakbytes,omitempty"`
	// private working set
	PrivateWorkingSet float64 `json:"privateworkingset,omitempty"`
}
