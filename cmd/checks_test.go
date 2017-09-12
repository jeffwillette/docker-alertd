package cmd

import (
	"encoding/json"
	"reflect"
	"testing"
)

var testStatsJSON = []byte(`
{
    "blkio_stats": {
        "io_merged_recursive": [],
        "io_queue_recursive": [],
        "io_service_bytes_recursive": [
            {
                "major": 8,
                "minor": 0,
                "op": "Read",
                "value": 446464
            },
            {
                "major": 8,
                "minor": 0,
                "op": "Write",
                "value": 0
            },
            {
                "major": 8,
                "minor": 0,
                "op": "Sync",
                "value": 0
            },
            {
                "major": 8,
                "minor": 0,
                "op": "Async",
                "value": 446464
            },
            {
                "major": 8,
                "minor": 0,
                "op": "Total",
                "value": 446464
            }
        ],
        "io_service_time_recursive": [],
        "io_serviced_recursive": [
            {
                "major": 8,
                "minor": 0,
                "op": "Read",
                "value": 77
            },
            {
                "major": 8,
                "minor": 0,
                "op": "Write",
                "value": 0
            },
            {
                "major": 8,
                "minor": 0,
                "op": "Sync",
                "value": 0
            },
            {
                "major": 8,
                "minor": 0,
                "op": "Async",
                "value": 77
            },
            {
                "major": 8,
                "minor": 0,
                "op": "Total",
                "value": 77
            }
        ],
        "io_time_recursive": [],
        "io_wait_time_recursive": [],
        "sectors_recursive": []
    },
    "cpu_stats": {
        "cpu_usage": {
            "percpu_usage": [
                3.0802479604e+10,
                3.1926624623e+10
            ],
            "total_usage": 6.2729104227e+10,
            "usage_in_kernelmode": 1.783e+10,
            "usage_in_usermode": 3.129e+10
        },
        "system_cpu_usage": 5.5157401e+14,
        "throttling_data": {
            "periods": 0,
            "throttled_periods": 0,
            "throttled_time": 0
        }
    },
    "id": "2ac67fb9001683144411b0212f97fb473a88cef96e240a16604dca8c5fa0e113",
    "memory_stats": {
        "limit": 2.096275456e+09,
        "max_usage": 7.184384e+07,
        "stats": {
            "active_anon": 5.5250944e+07,
            "active_file": 1.036288e+06,
            "cache": 1.51552e+06,
            "dirty": 0,
            "hierarchical_memory_limit": 9.223372036854772e+18,
            "hierarchical_memsw_limit": 9.223372036854772e+18,
            "inactive_anon": 0,
            "inactive_file": 479232,
            "mapped_file": 0,
            "pgfault": 324526,
            "pgmajfault": 0,
            "pgpgin": 167475,
            "pgpgout": 153616,
            "rss": 5.5250944e+07,
            "rss_huge": 0,
            "swap": 0,
            "total_active_anon": 5.5250944e+07,
            "total_active_file": 1.036288e+06,
            "total_cache": 1.51552e+06,
            "total_dirty": 0,
            "total_inactive_anon": 0,
            "total_inactive_file": 479232,
            "total_mapped_file": 0,
            "total_pgfault": 324526,
            "total_pgmajfault": 0,
            "total_pgpgin": 167475,
            "total_pgpgout": 153616,
            "total_rss": 5.5250944e+07,
            "total_rss_huge": 0,
            "total_swap": 0,
            "total_unevictable": 0,
            "total_writeback": 0,
            "unevictable": 0,
            "writeback": 0
        },
        "usage": 6.08256e+07
    },
    "name": "/test_container",
    "networks": {
        "eth0": {
            "rx_bytes": 90754,
            "rx_dropped": 0,
            "rx_errors": 0,
            "rx_packets": 2727,
            "tx_bytes": 2608,
            "tx_dropped": 0,
            "tx_errors": 0,
            "tx_packets": 36
        },
        "eth1": {
            "rx_bytes": 95165,
            "rx_dropped": 0,
            "rx_errors": 0,
            "rx_packets": 2781,
            "tx_bytes": 6961,
            "tx_dropped": 0,
            "tx_errors": 0,
            "tx_packets": 50
        }
    },
    "num_procs": 0,
    "pids_stats": {
        "current": 3
    },
    "precpu_stats": {
        "cpu_usage": {
            "percpu_usage": [
                3.0802223526e+10,
                3.1926350445e+10
            ],
            "total_usage": 6.2728573971e+10,
            "usage_in_kernelmode": 1.783e+10,
            "usage_in_usermode": 3.129e+10
        },
        "system_cpu_usage": 5.5157202e+14,
        "throttling_data": {
            "periods": 0,
            "throttled_periods": 0,
            "throttled_time": 0
        }
    },
    "preread": "2017-02-10T05:57:29.844015652Z",
    "read": "2017-02-10T05:57:30.844460171Z",
    "storage_stats": {}
}
`)

var cont1 = MondContainer{
	Name: "test_container",
	Alerts: []ContainerAlert{
		ContainerAlert{
			Name:     "CPU Usage Alert",
			Function: CheckCPUUsage,
			Limit:    20,
			Active:   false,
		},
	},
}

var cont2 = MondContainer{
	Name: "test_container",
	Alerts: []ContainerAlert{
		ContainerAlert{
			Name:     "CPU Usage Alert",
			Function: CheckCPUUsage,
			Limit:    0,
			Active:   false,
		},
	},
}

var alertdStats AlertdStats

func TestCheckCPUUsage(t *testing.T) {

	err := json.Unmarshal(testStatsJSON, &alertdStats)

	if err != nil {
		t.Error("Error unmarshaling the JSON")
	}

	type args struct {
		ca          *ContainerAlert
		alertdStats *AlertdStats
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 float64
	}{
		{
			name: "1: Testing a container that passes the check",
			args: args{
				ca:          &cont1.Alerts[0],
				alertdStats: &alertdStats,
			},
			want:  false,
			want1: 1, // REAL CPU usage is .02, passing if less than 1
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := CheckCPUUsage(tt.args.ca, tt.args.alertdStats)
			if got != tt.want {
				t.Errorf("CheckCPUUsage() got = %v, want %v", got, tt.want)
			}
			if got1 > tt.want1 {
				t.Errorf("CheckCPUUsage() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

var cMinPidPass = &ContainerAlert{
	"Check Minimum Processes pass",
	CheckMinPids,
	3,
	false,
}

var cMinPidFail = &ContainerAlert{
	"Check Minimum Processes fail",
	CheckMinPids,
	4,
	false,
}

func TestCheckMinPids(t *testing.T) {
	json.Unmarshal(testStatsJSON, &alertdStats)

	type args struct {
		ca          *ContainerAlert
		alertdStats *AlertdStats
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 float64
	}{
		{
			name: "1: Testing a container within the limits",
			args: args{
				ca:          cMinPidPass,
				alertdStats: &alertdStats,
			},
			want:  false,
			want1: 3,
		},
		{
			name: "2: Testing a container within the limits",
			args: args{
				ca:          cMinPidFail,
				alertdStats: &alertdStats,
			},
			want:  true,
			want1: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := CheckMinPids(tt.args.ca, tt.args.alertdStats)
			if got != tt.want {
				t.Errorf("CheckMinPids() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("CheckMinPids() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

var cMaxMemPass = &ContainerAlert{
	"Check Maximum Memory Pass",
	CheckMemory,
	100,
	false,
}

var cMaxMemFail = &ContainerAlert{
	"Check Maximum Memory Fail",
	CheckMemory,
	50,
	true,
}

func TestCheckMemory(t *testing.T) {
	json.Unmarshal(testStatsJSON, &alertdStats)

	type args struct {
		ca          *ContainerAlert
		alertdStats *AlertdStats
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 float64
	}{
		{
			name: "1: Test max memory pass",
			args: args{
				ca:          cMaxMemPass,
				alertdStats: &alertdStats,
			},
			want:  false,
			want1: 70,
		},
		{
			name: "1: Test max memory fail",
			args: args{
				ca:          cMaxMemFail,
				alertdStats: &alertdStats,
			},
			want:  true,
			want1: 70,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := CheckMemory(tt.args.ca, tt.args.alertdStats)
			if got != tt.want {
				t.Errorf("CheckMemory() got = %v, want %v", got, tt.want)
			}
			if got1 > tt.want1 {
				t.Errorf("CheckMemory() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestMondContainer_CheckContainer(t *testing.T) {
	json.Unmarshal(testStatsJSON, &alertdStats)

	type fields struct {
		Name   string
		Alerts []ContainerAlert
	}
	type args struct {
		a *AlertdStats
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			name: "1: testing containers that generate no alerts",
			fields: fields{
				Name: "test_container",
				Alerts: []ContainerAlert{
					ContainerAlert{
						Name:     "Check CPU",
						Function: CheckCPUUsage,
						Limit:    20,
						Active:   false,
					},
					ContainerAlert{
						Name:     "Check Mem",
						Function: CheckMemory,
						Limit:    90,
						Active:   false,
					},
					ContainerAlert{
						Name:     "check min PIDS",
						Function: CheckMinPids,
						Limit:    3,
						Active:   false,
					},
				},
			},
			args: args{
				&alertdStats,
			},
			want: []string{},
		},
		{
			name: "1: testing containers that generate 3 alerts",
			fields: fields{
				Name: "test_container",
				Alerts: []ContainerAlert{
					ContainerAlert{
						Name:     "Check CPU",
						Function: CheckCPUUsage,
						Limit:    0,
						Active:   false,
					},
					ContainerAlert{
						Name:     "Check Mem",
						Function: CheckMemory,
						Limit:    50,
						Active:   false,
					},
					ContainerAlert{
						Name:     "check min PIDS",
						Function: CheckMinPids,
						Limit:    4,
						Active:   false,
					},
				},
			},
			args: args{
				&alertdStats,
			},
			want: []string{
				"Check CPU: test_container exceeded alert threshold of 0, it is " +
					"currently using 0.026646.\n",
				"Check Mem: test_container exceeded alert threshold of 50, it is " +
					"currently using 60.825600.\n",
				"check min PIDS: test_container exceeded alert threshold of 4, it " +
					"is currently using 3.000000.\n",
			},
		},
		{
			name: "3: containers with active alerts (would make alerts, but don't)",
			fields: fields{
				Name: "test_container",
				Alerts: []ContainerAlert{
					ContainerAlert{
						Name:     "Check CPU",
						Function: CheckCPUUsage,
						Limit:    0,
						Active:   true,
					},
					ContainerAlert{
						Name:     "Check Mem",
						Function: CheckMemory,
						Limit:    50,
						Active:   true,
					},
					ContainerAlert{
						Name:     "check min PIDS",
						Function: CheckMinPids,
						Limit:    4,
						Active:   true,
					},
				},
			},
			args: args{
				&alertdStats,
			},
			want: []string{},
		},
		{
			name: "4: containers below threshold, with active alerts " +
				"(generates recovery alert)",
			fields: fields{
				Name: "t",
				Alerts: []ContainerAlert{
					ContainerAlert{
						Name:     "Check CPU",
						Function: CheckCPUUsage,
						Limit:    20,
						Active:   true,
					},
					ContainerAlert{
						Name:     "Check Mem",
						Function: CheckMemory,
						Limit:    80,
						Active:   true,
					},
					ContainerAlert{
						Name:     "check min PIDS",
						Function: CheckMinPids,
						Limit:    3,
						Active:   true,
					},
				},
			},
			args: args{
				&alertdStats,
			},
			want: []string{
				"Check CPU: t recovered. threshold: 20, current: 0.026646.\n",
				"Check Mem: t recovered. threshold: 80, current: 60.825600.\n",
				"check min PIDS: t recovered. threshold: 3, current: 3.000000.\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := &MondContainer{
				Name:   tt.fields.Name,
				Alerts: tt.fields.Alerts,
			}
			got := md.CheckContainer(tt.args.a)
			if !(len(got) == 0 && len(tt.want) == 0) && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MondContainer.CheckContainer() = %v, want %v", got, tt.want)
			}
		})
	}
}
