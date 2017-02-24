package monitor

import (
	"reflect"
	"testing"
)

var testJSON = []byte(`
{
	"containers": [
		{
			"name": "container1",
			"max-cpu": 0,
			"max-mem": 20,
			"min-procs": 3
		},
		{
			"name": "container2",
			"max-cpu": 20,
			"max-mem": 20,
			"min-procs": 4
		}
	],
	"email_addresses": {
		"from": "auto@freshpowpow.com",
		"to": [
			"jeff@gnarfresh.com"
		],
		"subject": "DOCKER ALERT"
	},
	"email_settings": {
		"smtp": "smtp.coolserver.com",
		"password": "gnarlesbarkely",
		"port": "587"
	}
}`)

var testJSON2 = []byte(`
{
	"containers": [
		{
			"name": "container1",
			"max-cpu": 0,
			"max-mem": 20,
			"min-procs": 3
		}
		{
			"name": "container2",
			"max-cpu": 20,
			"max-mem": 20,
			"min-procs": 4
		}
	],
	"email_addresses": {
		"from": "auto@freshpowpow.com",
		"to": [
			"jeff@gnarfresh.com"
		],
	},
	"email_settings": {
		"smtp": "smtp.coolserver.com",
		"password": "gnarlesbarkely",
		"port": 587
	}
}`)

func TestGetConfJSON(t *testing.T) {
	type args struct {
		j *[]byte
	}
	tests := []struct {
		name    string
		args    args
		wantC   *Conf
		wantErr bool
	}{
		{
			name: "1: testing all the things",
			args: args{&testJSON},
			wantC: &Conf{
				Containers: []Container{
					Container{
						Name:     "container1",
						MaxCPU:   0,
						MaxMem:   20,
						MinProcs: 3,
					},
					Container{
						Name:     "container2",
						MaxCPU:   20,
						MaxMem:   20,
						MinProcs: 4,
					},
				},
				Email: Email{
					From:    "auto@freshpowpow.com",
					To:      []string{"jeff@gnarfresh.com"},
					Subject: "DOCKER ALERT",
				},
				Emailer: Emailer{
					SMTP:     "smtp.coolserver.com",
					Password: "gnarlesbarkely",
					Port:     "587",
				},
			},
			wantErr: false,
		},
		{
			"2: Testing invalid JSON",
			args{&testJSON2},
			&Conf{},
			true,
		},
		{
			"3: Testing blank JSON",
			args{&[]byte{}},
			&Conf{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotC, err := GetConfJSON(tt.args.j)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotC, tt.wantC) {
				t.Errorf("GetConfJSON() = %v, want %v", gotC, tt.wantC)
			}
		})
	}
}
