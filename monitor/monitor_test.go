package monitor

import (
	"encoding/json"
	"fmt"
	"testing"
)

var test_json = `
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
`

// Unmarhsaling function that will be used in tests
func UnmarshalTestJson(b *[]byte) (*AlertdStats, error) {
	var alertdStats AlertdStats
	j := []byte(test_json)
	err := json.Unmarshal(j, &alertdStats)
	return &alertdStats, err
}

// Testing that the unmarshaling of JSON from the docker API actually works
func TestJsonUnmarshal(t *testing.T) {
	j := []byte(test_json)
	alertdStats, err := UnmarshalTestJson(&j)
	if err != nil {
		fmt.Printf("Error unmarshaling the JSON: %s", err)
		t.Fail()
	}

	if alertdStats.CPUStats.SystemUsage != float64(5.5157401e+14) ||
		alertdStats.PreCPUStats.SystemUsage != float64(5.5157202e+14) ||
		alertdStats.MemoryStats.Usage != float64(6.08256e+07) ||
		alertdStats.PidsStats.Current != uint64(3) {
		fmt.Println("Some of the unmarhsaled values do not match test" +
			"JSON blob")
		t.Fail()
	}
}
