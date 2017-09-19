package cmd

import (
	"testing"
)

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
		{
			Name: "config with complete email passes",
			Config: &Conf{
				Containers: []Container{
					Container{
						Name: "some_container",
					},
				},
				Email: Email{
					From:     "some@email.com",
					Password: "soopersecret",
					Port:     "587",
					SMTP:     "smtp@someserver.com",
					Subject:  "MY SUBJECT",
					To:       []string{"me@email.com"},
				},
			},
			ExpectedErr: nil,
		},
		{
			Name: "config with partial email fails",
			Config: &Conf{
				Containers: []Container{
					Container{
						Name: "some_container",
					},
				},
				Email: Email{
					Password: "soopersecret",
					Port:     "587",
					SMTP:     "smtp@someserver.com",
					Subject:  "MY SUBJECT",
					To:       []string{"me@email.com"},
				},
			},
			ExpectedErr: ErrEmailNoFrom,
		},
	}

	for _, test := range tests {
		err := test.Config.Validate()
		if !ErrContainsErr(err, test.ExpectedErr) {
			t.Errorf("%s:\nexpected err: %s\ngot err: %s\n", test.Name,
				test.ExpectedErr.Error(), err.Error())
		}
	}
}
