package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/pkg/errors"
)

// Evaluator evaluates a set of alerts and decides if they need to be sent
type Evaluator interface {
	Send(a []Alerter)
	Evaluate()
}

// Alert is the struct that stores information about alerts and its methods satisfy the
// Alerter interface
type Alert struct {
	Messages         []error
	SubjectAddendums []string
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
func (a *Alert) Add(e1, e2 error, s, subAddendum string) {

	a.SubjectAddendums = append(a.SubjectAddendums, subAddendum)

	e := e1
	if e2 != nil {
		e = errors.Wrap(e1, e2.Error())
	}

	err := errors.Wrap(e, s)
	a.Messages = append(a.Messages, err)
}

// Concat will concat different alerts from containers together into one
func (a *Alert) Concat(b ...*Alert) {
	for _, v := range b {
		for _, msg := range v.Messages {
			a.Messages = append(a.Messages, msg)
		}

		for _, addendum := range v.SubjectAddendums {
			a.SubjectAddendums = append(a.SubjectAddendums, addendum)
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
	a.SubjectAddendums = []string{}
}

// Dump takes the slice of alerts and dumps them to a single string
func (a *Alert) Dump() string {
	s := ""
	for _, v := range a.Messages {
		s += fmt.Sprintf("%s\n\n", v.Error())
	}
	return s
}

// DumpEmail behaves like dump, but formats them for email by splitting on ":" and adding
// \n\t (newline and tab) for the first two segments and joining the last segment. This
// should result in an email that is formatted as follows...
// [containerName]:
// 		[alertName]:
// 		Error: [errString]
func (a *Alert) DumpEmail() (s string) {
	for _, e := range a.Messages {
		errString := e.Error()
		splitErr := strings.SplitN(errString, ":", 3)

		for _, v := range splitErr {
			s += fmt.Sprintf("%s\n\t", v)
		}
		s += fmt.Sprintf("\n\n")

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
