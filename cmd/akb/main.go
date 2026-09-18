package main

import (
	"errors"
	"fmt"
	"os"
)

var version = "dev"

// Process exit codes: 0 = success, 1 = the command ran but produced a result
// the caller can act on and retry (failed page validation, raw drift detected),
// 2 = the command could not do its work (bad invocation or an akb fault).
const (
	exitSuccess = 0
	exitFailure = 1
	exitFault   = 2
)

// usageError reports a failure caused by how akb was invoked, such as a
// knowledge base without the directory a command needs. main reports it with a
// usage: prefix, so agents can tell invocation mistakes from akb faults.
type usageError struct {
	msg string
}

func (e *usageError) Error() string { return e.msg }

// internalError reports a failure inside akb itself. main reports it with an
// internal: prefix, so agents can tell akb faults from invocation mistakes.
type internalError struct {
	err error
}

func (e *internalError) Error() string { return e.err.Error() }

func (e *internalError) Unwrap() error { return e.err }

// validationFailure reports page content that failed write-time validation. The
// command that ran the rules has already printed every failed rule to stderr,
// so main reports nothing further and exits 1.
type validationFailure struct{}

func (validationFailure) Error() string { return "page validation failed" }

// driftDetected is `akb raw status`'s result signal for raw files that drifted
// from the manifest: a result, not a failure. The drift report has already been
// printed, so main reports nothing further and exits 1.
type driftDetected struct{}

func (driftDetected) Error() string { return "raw drift detected" }

// classifyExit maps a command error to its process exit code and the text to
// report on stderr. An empty report means the command already reported the
// reason itself.
func classifyExit(err error) (code int, report string) {
	var validationErr validationFailure
	var driftErr driftDetected
	var usageErr *usageError
	var internalErr *internalError

	switch {
	case err == nil:
		return exitSuccess, ""
	case errors.As(err, &validationErr), errors.As(err, &driftErr):
		return exitFailure, ""
	case errors.As(err, &usageErr):
		return exitFault, "usage: " + usageErr.Error()
	case errors.As(err, &internalErr):
		return exitFault, "internal: " + internalErr.Error()
	default:
		return exitFailure, "Error: " + err.Error()
	}
}

func main() {
	// main is the single place that reports failures and maps them to exit
	// codes, so cobra must not print errors or usage text of its own.
	RootCmd.SilenceErrors = true
	RootCmd.SilenceUsage = true

	code, report := classifyExit(Execute())
	if report != "" {
		fmt.Fprintln(os.Stderr, report)
	}
	os.Exit(code)
}
