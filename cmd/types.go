package cmd

import (
	"errors"
	"fmt"
	"log"
	"net/smtp"
	"reflect"
)

// Container gets data from the Unmarshaling of the configuration file JSON and stores
// the data throughout the course of the monitor.
type Container struct {
	Name     string
	MaxCPU   int64
	MaxMem   int64
	MinProcs int64
}

// EmailSettings implements the Alerter interface and sends emails
type EmailSettings struct {
	Active   bool
	SMTP     string
	Password string
	Port     string
	From     string
	To       []string
	Subject  string
}

// Conf struct that combines containers and email settings structs
type Conf struct {
	Containers    []Container
	EmailSettings EmailSettings
}

// Validate validates the configuration that was passed in
func (c *Conf) Validate() (err error) {
	switch {
	case reflect.DeepEqual(&Conf{}, c):
		return errors.New("The configuration cannot be empty, do you have a config file?")
	case len(c.Containers) < 1:
		return errors.New("There were no containers found in the configuration file")
	case len(c.EmailSettings.To) < 1:
		return errors.New("There was no \"To:\" email in the configuration")
	}

	importantStrings := []string{
		c.EmailSettings.From,
		c.EmailSettings.SMTP,
		c.EmailSettings.Password,
		c.EmailSettings.Port,
	}

	for _, v := range importantStrings {
		if len(v) == 0 {
			errorString := fmt.Sprintf("There is a missing field in the "+
				"configuration: %v", c)
			return errors.New(errorString)
		}
	}

	return nil
}

// Alerter is something that can send an alert either via email, or slack, etc.
type Alerter interface {
	Trigger() error
	ShouldSend() bool
	Evaluate()
	Email(e *EmailSettings) error
}

// Alert is the struct that stores information about alerts and its methods satisfy the
// Alerter interface
type Alert struct {
	Message string
}

// ShouldSend returns true if there is an alert message to be sent
func (a *Alert) ShouldSend() bool {
	return len(a.Message) > 0
}

// Evaluate will check if error should be sent and then trigger it if necessary
func (a *Alert) Evaluate() {
	if a.ShouldSend() {
		err := a.Trigger()
		if err != nil {
			log.Println(err)
		}
	}
}

// Add is for adding a call to Sprintf without making the actualt Sprintf call
func (a *Alert) Add(fmtString string, args ...interface{}) {
	a.Message += fmt.Sprintf(fmtString, args...)
}

// Trigger is for sending out alerts to syslog and to alerts that are active in conf
func (a *Alert) Trigger() error {
	log.Println(a.Message)
	//go func() {
	//	err := alert.Email(&c.EmailSettings)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	log.Println("alert email sent")
	//}()
	return nil
}

// Email sends an email alert
func (a *Alert) Email(e *EmailSettings) error {
	// The email message formatted properly
	formattedMsg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%v\r\n",
		e.To, e.Subject, a))

	// Set up authentication/address information
	auth := smtp.PlainAuth("", e.From, e.Password, e.SMTP)
	addr := fmt.Sprintf("%s:%s", e.SMTP, e.Port)

	err := smtp.SendMail(addr, auth, e.From, e.To, formattedMsg)
	if err != nil {
		return err
	}
	return nil
}
