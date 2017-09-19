package cmd

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

var cli *client.Client

var (
	stress = "deltaskelta/alpine-stress"
)

func TestMain(m *testing.M) {
	var err error
	cli, err = client.NewEnvClient()
	if err != nil {
		log.Println(err)
	}

	code := m.Run()

	_ = cli.ContainerRemove(context.TODO(), "test", types.ContainerRemoveOptions{
		Force: true,
	})

	os.Exit(code)
}

// TestContainer is used for starting test containers
type TestContainer struct {
	Image string
	Name  string
	CMD   []string
}

func (c *TestContainer) StartContainer() error {
	_, err := cli.ContainerCreate(context.TODO(), &container.Config{
		Image: c.Image,
		Cmd:   c.CMD,
	}, nil, nil, c.Name)
	if err != nil {
		return err
	}

	err = cli.ContainerStart(context.TODO(), c.Name, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	return nil
}

// Exec runs a command in a container
func (c *TestContainer) Exec(cmd ...string) error {
	resp, err := cli.ContainerExecCreate(context.TODO(), c.Name, types.ExecConfig{
		Cmd: cmd,
	})
	if err != nil {
		return err
	}

	err = cli.ContainerExecStart(context.TODO(), resp.ID, types.ExecStartCheck{})
	if err != nil {
		return err
	}

	return nil
}

func (c *TestContainer) StopContainer() error {
	err := cli.ContainerRemove(context.TODO(), c.Name, types.ContainerRemoveOptions{
		Force: true,
	})
	if err != nil {
		return err
	}

	return nil
}

func Setup(t *testing.T, name string, cnt []TestContainer) {
	log.Println(name)
	for _, c := range cnt {
		err := c.StartContainer()
		if err != nil {
			t.Error(err)
		}
	}
}

func Teardown(t *testing.T, cnt []TestContainer) {
	for _, c := range cnt {
		err := c.StopContainer()
		if err != nil {
			t.Error(err)
		}
	}
}

func CheckHasErr(e []error, s error) bool {
	for _, err := range e {
		if ErrContainsErr(err, s) {
			return true
		}
	}
	return false
}

func TestCheckExists(t *testing.T) {
	tests := []struct {
		Name               string
		Config             *Conf
		ExpectedAlertLen   int
		ExpectedAlert      error
		ExpectedShouldSend bool
		AlertActive        bool
		Containers         []TestContainer
	}{
		{
			Name: "test fails existence check",
			Config: &Conf{
				Duration:   100,
				Iterations: 2,
				Containers: []Container{
					{
						Name: "test",
					},
				},
			},
			ExpectedAlertLen:   1,
			ExpectedAlert:      ErrExistCheckFail,
			ExpectedShouldSend: true,
		},
		{
			Name: "test passes existence check",
			Config: &Conf{
				Duration:   100,
				Iterations: 2,
				Containers: []Container{
					{
						Name: "test",
					},
				},
			},
			ExpectedAlertLen:   0,
			ExpectedShouldSend: false,
			Containers: []TestContainer{
				TestContainer{
					Name:  "test",
					Image: stress,
					CMD:   []string{"echo", "hello world"},
				},
			},
		},
		{
			Name: "test recovers existence check",
			Config: &Conf{
				Duration:   100,
				Iterations: 2,
				Containers: []Container{
					{
						Name: "test",
					},
				},
			},
			ExpectedAlertLen:   1,
			ExpectedAlert:      ErrExistCheckRecovered,
			ExpectedShouldSend: true,
			AlertActive:        true,
			Containers: []TestContainer{
				TestContainer{
					Name:  "test",
					Image: stress,
					CMD:   []string{"echo", "hello world"},
				},
			},
		},
	}

	for _, test := range tests {
		Setup(t, test.Name, test.Containers)

		a := &Alert{Messages: []error{}}
		cnt := InitCheckers(test.Config)

		if test.AlertActive {
			cnt[0].ExistenceCheck.AlertActive = true
		}

		for i := int64(0); i < test.Config.Iterations; i++ {
			time.Sleep(1 * time.Second) // small delay to allow containers to exit
			CheckContainers(cnt, cli, a)

			if a.Len() != test.ExpectedAlertLen {
				t.Errorf("alert len %d does not match expected: %d\n", a.Len(), test.ExpectedAlertLen)
				t.Error(a.Messages)
			}

			if a.ShouldSend() != test.ExpectedShouldSend {
				t.Errorf("alert should send: %t does not match expected: %t", a.ShouldSend(), test.ExpectedShouldSend)
				t.Error(a.Messages)
			}

			if test.ExpectedAlert != nil {
				gotErr := CheckHasErr(a.Messages, test.ExpectedAlert)
				if !gotErr {
					t.Errorf("expected error message: %s not found in error messages", test.ExpectedAlert.Error())
					t.Error(a.Messages)
				}
			}

			time.Sleep(time.Duration(test.Config.Duration) * time.Millisecond)
		}

		Teardown(t, test.Containers)
	}
}

func TestCheckRunning(t *testing.T) {
	tests := []struct {
		Name               string
		Config             *Conf
		ExpectedAlertLen   int
		ExpectedAlert      error
		ExpectedShouldSend bool
		AlertActive        bool
		Containers         []TestContainer
	}{
		{
			Name: "test passes running check",
			Config: &Conf{
				Duration:   100,
				Iterations: 2,
				Containers: []Container{
					{
						Name:            "test",
						ExpectedRunning: true,
					},
				},
			},
			ExpectedAlertLen:   0,
			ExpectedShouldSend: false,
			Containers: []TestContainer{
				TestContainer{
					Name:  "test",
					Image: stress,
					CMD:   []string{"sleep", "10"},
				},
			},
		},
		{
			Name: "test fails running check",
			Config: &Conf{
				Duration:   100,
				Iterations: 2,
				Containers: []Container{
					{
						Name:            "test",
						ExpectedRunning: true,
					},
				},
			},
			ExpectedAlertLen:   1,
			ExpectedAlert:      ErrRunningCheckFail,
			ExpectedShouldSend: true,
			Containers: []TestContainer{
				TestContainer{
					Name:  "test",
					Image: stress,
					CMD:   []string{},
				},
			},
		},
		{
			Name: "test recovers running check",
			Config: &Conf{
				Duration:   100,
				Iterations: 2,
				Containers: []Container{
					{
						Name:            "test",
						ExpectedRunning: true,
					},
				},
			},
			ExpectedAlertLen:   1,
			ExpectedAlert:      ErrRunningCheckRecovered,
			ExpectedShouldSend: true,
			AlertActive:        true,
			Containers: []TestContainer{
				TestContainer{
					Name:  "test",
					Image: stress,
					CMD:   []string{"sleep", "10"},
				},
			},
		},
	}

	for _, test := range tests {
		Setup(t, test.Name, test.Containers)

		a := &Alert{Messages: []error{}}
		cnt := InitCheckers(test.Config)

		if test.AlertActive {
			cnt[0].RunningCheck.AlertActive = true
		}

		for i := int64(0); i < test.Config.Iterations; i++ {
			time.Sleep(1 * time.Second) // small delay to allow containers to exit
			CheckContainers(cnt, cli, a)

			if a.Len() != test.ExpectedAlertLen {
				t.Errorf("alert len %d does not match expected: %d\n", a.Len(), test.ExpectedAlertLen)
				t.Error(a.Messages)
			}

			if a.ShouldSend() != test.ExpectedShouldSend {
				t.Errorf("alert should send: %t does not match expected: %t", a.ShouldSend(), test.ExpectedShouldSend)
				t.Error(a.Messages)
			}

			if test.ExpectedAlert != nil {
				gotErr := CheckHasErr(a.Messages, test.ExpectedAlert)
				if !gotErr {
					t.Errorf("expected error message: %s not found in error messages", test.ExpectedAlert.Error())
					t.Error(a.Messages)
					t.Error(a.Len())
				}
			}

			time.Sleep(time.Duration(test.Config.Duration) * time.Millisecond)
		}

		Teardown(t, test.Containers)
	}
}

func TestCheckCPUUsage(t *testing.T) {
	tests := []struct {
		Name               string
		Config             *Conf
		ExpectedAlertLen   int
		ExpectedAlert      error
		ExpectedShouldSend bool
		AlertActive        bool
		Containers         []TestContainer
	}{
		{
			Name: "test fails cpu check",
			Config: &Conf{
				Duration:   100,
				Iterations: 2,
				Containers: []Container{
					{
						MaxCPU:          20,
						ExpectedRunning: true,
						Name:            "test",
					},
				},
			},
			ExpectedAlertLen:   1,
			ExpectedAlert:      ErrCPUCheckFail,
			ExpectedShouldSend: true,
			Containers: []TestContainer{
				TestContainer{
					Name:  "test",
					Image: stress,
					CMD:   []string{"stress", "--cpu", "1"},
				},
			},
		},
		{
			Name: "test passes cpu check",
			Config: &Conf{
				Duration:   100,
				Iterations: 2,
				Containers: []Container{
					{
						MaxCPU:          20,
						ExpectedRunning: true,
						Name:            "test",
					},
				},
			},
			ExpectedAlertLen:   0,
			ExpectedShouldSend: false,
			Containers: []TestContainer{
				TestContainer{
					Name:  "test",
					Image: stress,
					CMD:   []string{"sleep", "5"},
				},
			},
		},
		{
			Name: "test recover cpu check",
			Config: &Conf{
				Duration:   100,
				Iterations: 2,
				Containers: []Container{
					{
						MaxCPU:          20,
						ExpectedRunning: true,
						Name:            "test",
					},
				},
			},
			ExpectedAlertLen:   1,
			ExpectedAlert:      ErrCPUCheckRecovered,
			ExpectedShouldSend: true,
			AlertActive:        true,
			Containers: []TestContainer{
				TestContainer{
					Name:  "test",
					Image: stress,
					CMD:   []string{"sleep", "5"},
				},
			},
		},
	}

	for _, test := range tests {
		Setup(t, test.Name, test.Containers)

		a := &Alert{Messages: []error{}}
		cnt := InitCheckers(test.Config)

		if test.AlertActive {
			cnt[0].CPUCheck.AlertActive = true
		}

		for i := int64(0); i < test.Config.Iterations; i++ {
			CheckContainers(cnt, cli, a)

			if a.Len() != test.ExpectedAlertLen {
				t.Errorf("alert len %d does not match expected: %d\n", a.Len(), test.ExpectedAlertLen)
				t.Error(a.Messages)
			}

			if a.ShouldSend() != test.ExpectedShouldSend {
				t.Errorf("alert should send: %t does not match expected: %t", a.ShouldSend(), test.ExpectedShouldSend)
				t.Error(a.Messages)
			}

			if test.ExpectedAlert != nil {
				gotErr := CheckHasErr(a.Messages, test.ExpectedAlert)
				if !gotErr {
					t.Errorf("expected error message: %s not found in error messages", test.ExpectedAlert.Error())
					t.Error(a.Messages)
				}
			}

			time.Sleep(time.Duration(test.Config.Duration) * time.Millisecond)
		}

		Teardown(t, test.Containers)
	}

}

func TestCheckMemory(t *testing.T) {
	tests := []struct {
		Name               string
		Config             *Conf
		ExpectedAlertLen   int
		ExpectedAlert      error
		ExpectedShouldSend bool
		AlertActive        bool
		Containers         []TestContainer
	}{
		{
			Name: "test fails mem check",
			Config: &Conf{
				Duration:   100,
				Iterations: 2,
				Containers: []Container{
					{
						MaxMem:          10,
						ExpectedRunning: true,
						Name:            "test",
					},
				},
			},
			ExpectedAlertLen:   1,
			ExpectedAlert:      ErrMemCheckFail,
			ExpectedShouldSend: true,
			Containers: []TestContainer{
				TestContainer{
					Name:  "test",
					Image: stress,
					CMD:   []string{"stress", "--vm", "1", "--vm-bytes", "2G"},
				},
			},
		},
		{
			Name: "test passes memcheck",
			Config: &Conf{
				Duration:   100,
				Iterations: 2,
				Containers: []Container{
					{
						MaxMem:          10,
						ExpectedRunning: true,
						Name:            "test",
					},
				},
			},
			ExpectedAlertLen:   0,
			ExpectedShouldSend: false,
			Containers: []TestContainer{
				TestContainer{
					Name:  "test",
					Image: stress,
					CMD:   []string{"sleep", "5"},
				},
			},
		},
		{
			Name: "test recovered memcheck",
			Config: &Conf{
				Duration:   100,
				Iterations: 2,
				Containers: []Container{
					{
						MaxMem:          10,
						ExpectedRunning: true,
						Name:            "test",
					},
				},
			},
			ExpectedAlertLen:   1,
			ExpectedAlert:      ErrMemCheckRecovered,
			ExpectedShouldSend: true,
			AlertActive:        true,
			Containers: []TestContainer{
				TestContainer{
					Name:  "test",
					Image: stress,
					CMD:   []string{"sleep", "5"},
				},
			},
		},
	}

	for _, test := range tests {
		Setup(t, test.Name, test.Containers)

		a := &Alert{Messages: []error{}}
		cnt := InitCheckers(test.Config)

		if test.AlertActive {
			cnt[0].MemCheck.AlertActive = true
		}

		for i := int64(0); i < test.Config.Iterations; i++ {
			CheckContainers(cnt, cli, a)

			if a.Len() != test.ExpectedAlertLen {
				t.Errorf("alert len %d does not match expected: %d\n", a.Len(), test.ExpectedAlertLen)
				t.Error(a.Messages)
			}

			if a.ShouldSend() != test.ExpectedShouldSend {
				t.Errorf("alert should send: %t does not match expected: %t", a.ShouldSend(), test.ExpectedShouldSend)
				t.Error(a.Messages)
			}

			if test.ExpectedAlert != nil {
				gotErr := CheckHasErr(a.Messages, test.ExpectedAlert)
				if !gotErr {
					t.Errorf("expected error message: %s not found in error messages", test.ExpectedAlert.Error())
					t.Error(a.Messages)
				}
			}

			time.Sleep(time.Duration(test.Config.Duration) * time.Millisecond)
		}

		Teardown(t, test.Containers)
	}
}

func TestCheckPID(t *testing.T) {
	tests := []struct {
		Name               string
		Config             *Conf
		ExpectedAlertLen   int
		ExpectedAlert      error
		ExpectedShouldSend bool
		AlertActive        bool
		Cmd                []string
		Containers         []TestContainer
	}{
		{
			Name: "test passes PID check",
			Config: &Conf{
				Duration:   100,
				Iterations: 2,
				Containers: []Container{
					{
						MinProcs:        2,
						ExpectedRunning: true,
						Name:            "test",
					},
				},
			},
			ExpectedAlertLen:   0,
			ExpectedShouldSend: false,
			Cmd:                []string{"sleep", "100"},
			Containers: []TestContainer{
				TestContainer{
					Name:  "test",
					Image: stress,
					CMD:   []string{"sleep", "100"},
				},
			},
		},
		{
			Name: "test fails PID check",
			Config: &Conf{
				Duration:   100,
				Iterations: 2,
				Containers: []Container{
					{
						MinProcs:        3,
						ExpectedRunning: true,
						Name:            "test",
					},
				},
			},
			ExpectedAlertLen:   1,
			ExpectedAlert:      ErrMinPIDCheckFail,
			ExpectedShouldSend: true,
			Cmd:                []string{"sleep", "100"},
			Containers: []TestContainer{
				TestContainer{
					Name:  "test",
					Image: stress,
					CMD:   []string{"sleep", "100"},
				},
			},
		},
		{
			Name: "test recovers PID check",
			Config: &Conf{
				Duration:   100,
				Iterations: 2,
				Containers: []Container{
					{
						MinProcs:        2,
						ExpectedRunning: true,
						Name:            "test",
					},
				},
			},
			ExpectedAlertLen:   1,
			ExpectedAlert:      ErrMinPIDCheckRecovered,
			ExpectedShouldSend: true,
			AlertActive:        true,
			Cmd:                []string{"sleep", "100"},
			Containers: []TestContainer{
				TestContainer{
					Name:  "test",
					Image: stress,
					CMD:   []string{"sleep", "100"},
				},
			},
		},
	}

	for _, test := range tests {
		Setup(t, test.Name, test.Containers)

		err := test.Containers[0].Exec("sleep", "100")
		if err != nil {
			t.Error(err)
		}

		a := &Alert{Messages: []error{}}
		cnt := InitCheckers(test.Config)

		if test.AlertActive {
			cnt[0].PIDCheck.AlertActive = true
		}

		for i := int64(0); i < test.Config.Iterations; i++ {
			CheckContainers(cnt, cli, a)

			if a.Len() != test.ExpectedAlertLen {
				t.Errorf("alert len %d does not match expected: %d\n", a.Len(), test.ExpectedAlertLen)
				t.Error(a.Messages)
			}

			if a.ShouldSend() != test.ExpectedShouldSend {
				t.Errorf("alert should send: %t does not match expected: %t", a.ShouldSend(), test.ExpectedShouldSend)
				t.Error(a.Messages)
			}

			if test.ExpectedAlert != nil {
				gotErr := CheckHasErr(a.Messages, test.ExpectedAlert)
				if !gotErr {
					t.Errorf("expected error message: %s not found in error messages", test.ExpectedAlert.Error())
					t.Error(a.Messages)
				}
			}

			time.Sleep(time.Duration(test.Config.Duration) * time.Millisecond)
		}

		Teardown(t, test.Containers)
	}
}
