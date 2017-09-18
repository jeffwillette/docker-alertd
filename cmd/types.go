package cmd

import (
	"fmt"
	"log"
	"net/smtp"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

// Container gets data from the Unmarshaling of the configuration file JSON and stores
// the data throughout the course of the monitor.
type Container struct {
	Name            string
	MaxCPU          uint64
	MaxMem          uint64
	MinProcs        uint64
	ExpectedRunning bool
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
	Iterations    int64
	Duration      int64
}

// these errors are for the purpose of being able to compare them later
var (
	ExistFailMsg             = "Existence check failure"
	ExistRecoverMsg          = "Existence check recovered"
	RunningCheckFailMsg      = "Running check failure"
	RunningCheckRecoveredMsg = "Running check recovered"
	CPUCheckFailMsg          = "CPU check failure"
	CPUCheckRecoveredMsg     = "CPu check recovered"
	MemCheckFailMsg          = "Memory check failure"
	MemCheckRecoverMsg       = "Memory check recovered"
	MinPIDCheckFailMsg       = "Min PID check Failure"
	MinPIDCheckRecoverMsg    = "Min PID check recovered"
	MaxPIDCheckFailMsg       = "Max PID check Failure"
	MaxPIDCheckRecoveredMsg  = "Max PID check recovered"
	UnknownErrMsg            = "Received an unknown error"

	ErrEmptyConfig           = errors.New("The configuration cannot be empty, do you have a config file?")
	ErrNoContainers          = errors.New("There were no containers found in the configuration file")
	ErrExistCheckFail        = errors.New(ExistFailMsg)
	ErrExistCheckRecovered   = errors.New(ExistRecoverMsg)
	ErrRunningCheckFail      = errors.New(RunningCheckFailMsg)
	ErrRunningCheckRecovered = errors.New(RunningCheckRecoveredMsg)
	ErrCPUCheckFail          = errors.New(CPUCheckFailMsg)
	ErrCPUCheckRecovered     = errors.New(CPUCheckRecoveredMsg)
	ErrMemCheckFail          = errors.New(MemCheckFailMsg)
	ErrMemCheckRecovered     = errors.New(MemCheckRecoverMsg)
	ErrMinPIDCheckFail       = errors.New(MinPIDCheckFailMsg)
	ErrMinPIDCheckRecovered  = errors.New(MinPIDCheckRecoverMsg)
	ErrMaxPIDCheckFail       = errors.New(MaxPIDCheckFailMsg)
	ErrMaxPIDCheckRecovered  = errors.New(MaxPIDCheckRecoveredMsg)
	ErrUnknown               = errors.New(UnknownErrMsg)
)

// ErrIsErr returns true if the error string contains the message
func ErrIsErr(e error, baseErr string) bool {
	if strings.Contains(e.Error(), baseErr) {
		return true
	}
	return false
}

// Validate validates the configuration that was passed in
func (c *Conf) Validate() (err error) {
	switch {
	case reflect.DeepEqual(&Conf{}, c):
		return ErrEmptyConfig
	case len(c.Containers) < 1:
		return ErrNoContainers
	}

	return nil
}

// Alerter is something that can send an alert either via email, or slack, etc.
type Alerter interface {
	Alert() error
	ShouldSend() bool
	Evaluate()
	Email(e *EmailSettings) error
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
		err := a.Alert()
		if err != nil {
			log.Println(err)
		}
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

// Alert is for sending out alerts to syslog and to alerts that are active in conf
func (a *Alert) Alert() error {
	a.Log()
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
