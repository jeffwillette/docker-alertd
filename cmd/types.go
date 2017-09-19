package cmd

import (
	"fmt"
	"log"
	"net/smtp"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

// these errors are for the purpose of being able to compare them later
var (
	ErrEmptyConfig           = errors.New("The configuration is completely empty (check config file)")
	ErrEmailNoSMTP           = errors.New("no email SMTP server")
	ErrEmailNoTo             = errors.New("no email to addresses")
	ErrEmailNoFrom           = errors.New("no email from addresses")
	ErrEmailNoPass           = errors.New("no email password")
	ErrEmailNoPort           = errors.New("no email port")
	ErrEmailNoSubject        = errors.New("no email subject")
	ErrNoContainers          = errors.New("There were no containers found in the configuration file")
	ErrExistCheckFail        = errors.New("Existence check failure")
	ErrExistCheckRecovered   = errors.New("Existence check recovered")
	ErrRunningCheckFail      = errors.New("Running check failure")
	ErrRunningCheckRecovered = errors.New("Running check recovered")
	ErrCPUCheckFail          = errors.New("CPU check failure")
	ErrCPUCheckRecovered     = errors.New("CPU check recovered")
	ErrMemCheckFail          = errors.New("Memory check failure")
	ErrMemCheckRecovered     = errors.New("Memory check recovered")
	ErrMinPIDCheckFail       = errors.New("Min PID check Failure")
	ErrMinPIDCheckRecovered  = errors.New("Min PID check recovered")
	ErrMaxPIDCheckFail       = errors.New("Max PID check Failure")
	ErrMaxPIDCheckRecovered  = errors.New("Max PID check recovered")
	ErrUnknown               = errors.New("Received an unknown error")
)

// ErrContainsErr returns true if the error string contains the message
func ErrContainsErr(e, b error) bool {
	switch {
	case e == nil && b == nil:
		return true // they are both nil and essentially equal
	case e == nil || b == nil:
		return false // one of them is nil (previous case took care of both nils)
	case strings.Contains(e.Error(), b.Error()):
		return true // b is within a
	default:
		return false
	}
}

// Container gets data from the Unmarshaling of the configuration file JSON and stores
// the data throughout the course of the monitor.
type Container struct {
	Name            string
	MaxCPU          uint64
	MaxMem          uint64
	MinProcs        uint64
	ExpectedRunning bool
}

// Alerter is the interface which will handle alerting via different methods such as email
// and twitter/slack
type Alerter interface {
	Alert(a *Alert) error
}

// EmailSettings implements the Alerter interface and sends emails
type EmailSettings struct {
	SMTP     string
	Password string
	Port     string
	From     string
	To       []string
	Subject  string
}

// Alert sends an email alert
func (e EmailSettings) Alert(a *Alert) error {
	// alerts in string form
	alerts := a.Dump()

	// The email message formatted properly
	formattedMsg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n",
		e.To, e.Subject, alerts))

	// Set up authentication/address information
	auth := smtp.PlainAuth("", e.From, e.Password, e.SMTP)
	addr := fmt.Sprintf("%s:%s", e.SMTP, e.Port)

	err := smtp.SendMail(addr, auth, e.From, e.To, formattedMsg)
	if err != nil {
		return errors.Wrap(err, "error sending email")
	}

	log.Println("alert email sent")

	return nil
}

// Valid returns true if the email settings are complete
func (e *EmailSettings) Valid() error {
	errString := []string{}

	if e.SMTP == "" {
		errString = append(errString, ErrEmailNoSMTP.Error())
	}

	if len(e.To) < 1 {
		errString = append(errString, ErrEmailNoTo.Error())
	}

	if e.From == "" {
		errString = append(errString, ErrEmailNoFrom.Error())
	}

	if e.Password == "" {
		errString = append(errString, ErrEmailNoPass.Error())
	}

	if e.Port == "" {
		errString = append(errString, ErrEmailNoPort.Error())
	}

	if e.Subject == "" {
		errString = append(errString, ErrEmailNoSubject.Error())
	}

	if len(errString) == 0 {
		return nil
	}

	delimErr := strings.Join(errString, ", ")
	err := errors.New(delimErr)

	return errors.Wrap(err, "email settings validation fail")
}

// Conf struct that combines containers and email settings structs
type Conf struct {
	Containers    []Container
	EmailSettings EmailSettings
	Iterations    int64
	Duration      int64
	Alerters      []Alerter
}

// Validate validates the configuration that was passed in
func (c *Conf) Validate() error {
	// the error to wrap and return at the end
	errString := []string{}

	if reflect.DeepEqual(&Conf{}, c) {
		errString = append(errString, ErrEmptyConfig.Error())
	}

	if len(c.Containers) < 1 {
		errString = append(errString, ErrNoContainers.Error())
	}

	e := c.EmailSettings.Valid()
	switch {
	case reflect.DeepEqual(EmailSettings{}, c.EmailSettings):
		// do nothing because the settings are empty and assumed omitted
	case e != nil:
		errString = append(errString, e.Error())
	default:
		c.Alerters = append(c.Alerters, c.EmailSettings)
	}

	if len(errString) == 0 {
		return nil
	}

	delimErr := strings.Join(errString, ", ")
	err := errors.New(delimErr)

	return errors.Wrap(err, "config validation fail")
}

// Evaluator evaluates a set of alerts and decides if they need to be sent
type Evaluator interface {
	Send(a []Alerter)
	ShouldSend() bool
	Evaluate()
}

// Alert is the struct that stores information about alerts and its methods satisfy the
// Alerter interface
type Alert struct {
	Messages []error
}

// ShouldSend returns true if there is an alert message to be sent
func (a *Alert) ShouldSend() bool {
	return len(a.Messages) > 0
}

// Evaluate will check if error should be sent and then trigger it if necessary
func (a *Alert) Evaluate() {
	if a.ShouldSend() {
		a.Send(Config.Alerters)
	}
}

// Len returns the length of the alert message strings
func (a *Alert) Len() int {
	return len(a.Messages)
}

// Add should take in an error and wrap it
func (a *Alert) Add(e1, e2 error, fs string, args ...interface{}) {

	e := e1
	if e2 != nil {
		e = errors.Wrap(e1, e2.Error())
	}

	err := errors.Wrapf(e, fs, args...)
	a.Messages = append(a.Messages, err)
}

// Concat will concat different alerts from containers together into one
func (a *Alert) Concat(b ...*Alert) {
	for _, v := range b {
		for _, msg := range v.Messages {
			a.Messages = append(a.Messages, msg)
		}
	}
}

// Log prints the alert to the log
func (a *Alert) Log() {
	log.Println("ALERT:")
	for _, msg := range a.Messages {
		log.Println(msg)
	}
}

// Clear will reset the alert to an empty string
func (a *Alert) Clear() {
	a.Messages = []error{}
}

// Dump takes the slice of alerts and dumps them to a single string
func (a *Alert) Dump() string {
	s := ""
	for _, v := range a.Messages {
		s += fmt.Sprintf("%s\n\n", v.Error())
	}
	return s
}

// Send is for sending out alerts to syslog and to alerts that are active in conf
func (a *Alert) Send(b []Alerter) {
	a.Log()
	for i := range b {
		go func(c Alerter) {
			err := c.Alert(a)
			if err != nil {
				log.Println(err)
			}
		}(b[i])
	}
}
