package cmd

import "testing"

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		Name        string
		Config      *Conf
		ExpectedErr error
	}{
		{
			Name:        "testing blank config fails",
			Config:      &Conf{},
			ExpectedErr: ErrEmptyConfig,
		},
		{
			Name: "config with one container (pass)",
			Config: &Conf{
				Containers: []Container{
					Container{
						Name: "some_container",
					},
				},
			},
			ExpectedErr: nil,
		},
	}

	for _, test := range tests {
		err := test.Config.Validate()
		if err != test.ExpectedErr {
			t.Errorf("%s:\nexpected err: %s\ngot err: %s\n", test.Name,
				test.ExpectedErr.Error(), err.Error())
		}
	}
}
