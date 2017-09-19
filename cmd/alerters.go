package cmd

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

// Alerter is the interface which will handle alerting via different methods such as email
// and twitter/slack
type Alerter interface {
	Valid() error
	Alert(a *Alert) error
}

// Email implements the Alerter interface and sends emails
type Email struct {
	SMTP     string
	Password string
	Port     string
	From     string
	To       []string
	Subject  string
}

// Alert sends an email alert
func (e Email) Alert(a *Alert) error {
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
func (e Email) Valid() error {
	errString := []string{}

	if reflect.DeepEqual(Email{}, e) {
		return nil // assume that email alerts were omitted
	}

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

// Slack contains all the info needed to connect to a slack channel
type Slack struct {
	WebhookURL string
}

// Valid returns an error if slack settings are invalid
func (s Slack) Valid() error {
	errString := []string{}

	if reflect.DeepEqual(Slack{}, s) {
		return nil // assume that slack was omitted
	}

	if s.WebhookURL == "" {
		errString = append(errString, ErrSlackNoWebHookURL.Error())
	}

	if len(errString) == 0 {
		return nil
	}

	delimErr := strings.Join(errString, ", ")
	err := errors.New(delimErr)

	return errors.Wrap(err, "slack settings validation fail")
}

// Alert sends the alert to a slack channel
func (s Slack) Alert(a *Alert) error {
	alerts := a.Dump()

	json := fmt.Sprintf("{\"text\": \"%s\"}", alerts)
	body := bytes.NewReader([]byte(json))
	resp, err := http.Post(s.WebhookURL, "application/json", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Println("sent alert to slack")
	return nil
}
