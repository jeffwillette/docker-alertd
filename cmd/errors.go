package cmd

import (
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
	ErrSlackNoWebHookURL     = errors.New("no slack webhook url")
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
	ErrPushoverApiToken     = errors.New("no pushover api token")
	ErrPushoverUserKey     = errors.New("no pushover user key")
	ErrPushoverApiURL     = errors.New("no pushover api url")
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
