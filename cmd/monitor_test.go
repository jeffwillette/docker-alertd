package cmd

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/docker/docker/api/types"
)

// TestCloser is to satisfy a closer so I can build a ReadCloser
type TestCloser struct{}

// Close is the function that satisfies the Closer Interface
func (tc TestCloser) Close() error {
	return nil
}

// TestReadCloser satisfied the ReadCloser interface in io package. and is used in
// making my fake ReadCloser in the ContainerStats
type TestReadCloser struct {
	io.Reader
	io.Closer
}

func TestUnmarshalStats(t *testing.T) {

	// First there needs to be a ContainerStats object to put into the test
	cs := types.ContainerStats{
		Body: TestReadCloser{
			Reader: bytes.NewReader(testStatsJSON),
			Closer: TestCloser{},
		},
		OSType: "Lindows",
	}

	type args struct {
		c types.ContainerStats
	}
	tests := []struct {
		name string
		args args
		want *AlertdStats
	}{
		{
			name: "1: testing that the JSON unmarshals correctly into *AlertdStats",
			args: args{
				cs,
			},
			want: &AlertdStats{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is making sure that the AlertdStats that was received is not
			// equal to the empty one in the test
			if got := UnmarshalStats(tt.args.c); reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnmarshalStats() = %v, want %v", got, tt.want)
			}
		})
	}
}
