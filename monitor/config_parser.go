package monitor

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/smtp"
)

// Container gets data from the Unmarshaling of the configuration file JSON and stores
// the data throughout the course of the monitor.
type Container struct {
	Name     string `json:"name"`
	MaxCPU   int64  `json:"max-cpu"`
	MaxMem   int64  `json:"max-mem"`
	MinProcs int64  `json:"min-procs"`
}

// Email is partly taken from the email_settings in the JSON configuration file
// and then the subject and message part of it is made when alert is triggerred
type Email struct {
	From    string   `json:"from"`
	To      []string `json:"to"` //TODO: Make the to into slice of strings
	Subject string   `json:"subject"`
	Message []byte
}

// Mailer is the same as smtp.SendMail
type Mailer interface {
	Send(mail *Email) error
}

// Emailer implements the Mailer interface and sends emails
type Emailer struct {
	SMTP     string `json:"smtp"`
	Password string `json:"password"`
	Port     string `json:"port"`
}

// Send Is the method that satisfies the interface Mailer, which is passed
// to send alert
func (mlr Emailer) Send(ml *Email) error {
	// Set up authentication information.
	auth := smtp.PlainAuth("", ml.From, mlr.Password, mlr.SMTP)
	addr := mlr.SMTP + ":" + string(mlr.Port)
	err := smtp.SendMail(addr, auth, ml.From, ml.To, ml.Message)
	if err != nil {
		return err
	}
	return nil
}

// Conf struct that combines containers and email settings structs
type Conf struct {
	Containers []Container `json:"containers"`
	Email      Email       `json:"email_addresses"`
	Emailer    Emailer     `json:"email_settings"`
}

// GetConfJSON Parse the configuration file and log fatal if there are any errors, or there are
// 0 containers
func GetConfJSON(j *[]byte) (c *Conf, err error) {
	var conf Conf
	err = json.Unmarshal(*j, &conf)
	if err != nil {
		// returning and empty pointer to a Conf when there is an error
		return &Conf{}, err
	}

	if len(conf.Containers) < 1 {
		err := errors.New(
			"There were no containers found in the configuration file")
		// return an empty pointer to a Conf when there are no containers
		return &Conf{}, err
	} else if len(conf.Email.To) < 1 {
		err := errors.New("There was no \"To:\" email in the configuration")
		return &Conf{}, err
	}

	importantStrings := []string{
		conf.Email.From,
		conf.Emailer.SMTP,
		conf.Emailer.Password,
		conf.Emailer.Port,
	}

	for _, v := range importantStrings {
		if len(v) == 0 {
			errorString := fmt.Sprintf("There is a missing field in the "+
				"congiguration: %v", conf)
			err := errors.New(errorString)
			return &Conf{}, err
		}
	}

	return &conf, err
}
