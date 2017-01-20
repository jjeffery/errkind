// Package errkind is used to create and detect specific kinds of errors,
// based on single method interfaces that the errors support.
//
// Supported interfaces
//
// Temporary errors are detected using the ``temporaryer'' interface. Some
// errors in the Go standard library implement this interface. (See net.AddrError,
// net.DNSConfigError, and net.DNSError for examples).
//  type temporaryer interface {
//      Temporary() bool
//  }
//
// Some packages return errors which implement the ``coder''
// interface, which allows the error to report an application-specific
// error condition.
//  type coder interface {
//      Code() string
//  }
// The AWS SDK for Go is a popular third party library that follows this
// convention.
//
// In addition some third party packages (including the AWS SDK) follow the
// convention of reporting HTTP status values using the ``statusCoder'' interface.
//  type statusCoder interface {
//      StatusCode() int
//  }
// An alternative interface provides the same meaning:
//  type statuser interface {
//      Status() int
//  }
//
// The publicer interface identifies an error as being suitable for displaying
// to a requesting client. The error message does not contain any implementation
// details that could leak sensitive information.
//  type publicer interface {
//      Public() bool
//  }
package errkind

import (
	"net/http"
	"strings"

	"github.com/go-stack/stack"
	"github.com/jjeffery/errors"
)

// cause is an interface implemented by errors that have a cause error.
type causer interface {
	Cause() error
}

// temporaryer is an interface implemented by errors that communicate
// if they are temporary or not. Temporary errors can be retried.
type temporaryer interface {
	Temporary() bool
}

// coder is an interface implemented by errors that return a string code.
// Useful for checking AWS error codes.
type coder interface {
	Code() string
}

// statusCode is an interface implemented by errors that return an integer status code.
// Userful for checking AWS status codes.
type statusCoder interface {
	StatusCode() int
}

// statuser is an alternative interface implemented by errors that return an integer status code.
type statuser interface {
	Status() int
}

// publicer is an interface implemented by errors whose contents are suitable
// for returning to requesting clients. Their message does not include implementation details.
type publicer interface {
	Public() bool
}

// HasCode determines whether the error has any of the codes associated with it.
func HasCode(err error, codes ...string) bool {
	err = errors.Cause(err)
	if err == nil {
		return false
	}
	if errCoder, ok := err.(coder); ok {
		errCode := errCoder.Code()
		for _, code := range codes {
			if errCode == code {
				return true
			}
		}
	}
	return false
}

// HasStatus determines whether the error has any of the statuses associated with it.
func HasStatus(err error, statuses ...int) bool {
	err = errors.Cause(err)
	if err == nil {
		return false
	}
	if errStatusCoder, ok := err.(statusCoder); ok {
		errStatus := errStatusCoder.StatusCode()
		for _, status := range statuses {
			if errStatus == status {
				return true
			}
		}
	}
	if errStatuser, ok := err.(statuser); ok {
		errStatus := errStatuser.Status()
		for _, status := range statuses {
			if errStatus == status {
				return true
			}
		}
	}
	return false
}

// Status returns the status code associated with err, or
// zero if there is no status.
func Status(err error) int {
	err = errors.Cause(err)
	if err == nil {
		return 0
	}
	if errStatusCoder, ok := err.(statusCoder); ok {
		return errStatusCoder.StatusCode()
	}
	if errStatuser, ok := err.(statuser); ok {
		return errStatuser.Status()
	}
	return 0
}

// Code returns the string error code associated with err, or
// a blank string if there is no code.
func Code(err error) string {
	err = errors.Cause(err)
	if err == nil {
		return ""
	}
	if errCoder, ok := err.(coder); ok {
		return errCoder.Code()
	}
	return ""
}

// IsTemporary returns true for errors that indicate
// an error condition that may succeed if retried.
//
// An error is considered temporary if it implements
// the following interface and its Temporary method returns true.
//  type temporaryer interface {
//      Temporary() bool
//  }
func IsTemporary(err error) bool {
	err = errors.Cause(err)
	for err == nil {
		return false
	}
	if temporary, ok := err.(temporaryer); ok {
		return temporary.Temporary()
	}
	return false
}

// publicStatusError implements error, statusCoder and publicer interfaces.
type publicStatusError struct {
	message string
	status  int
}

func (s publicStatusError) Error() string {
	return s.message
}

func (s publicStatusError) StatusCode() int {
	return s.status
}

func (s publicStatusError) Status() int {
	return s.status
}

func (s publicStatusError) Public() bool {
	return true
}

func (s publicStatusError) With(keyvals ...interface{}) errors.Error {
	return errors.Wrap(s).With(keyvals...)
}

// publicStatusCodeError implements error, statusCoder, coder and publicer interfaces.
type publicStatusCodeError struct {
	message string
	status  int
	code    string
}

func (s publicStatusCodeError) Error() string {
	return s.message
}

func (s publicStatusCodeError) StatusCode() int {
	return s.status
}

func (s publicStatusCodeError) Status() int {
	return s.status
}

func (s publicStatusCodeError) Code() string {
	return s.code
}

func (s publicStatusCodeError) Public() bool {
	return true
}

func (s publicStatusCodeError) With(keyvals ...interface{}) errors.Error {
	return errors.Wrap(s).With(keyvals...)
}

// makeMessage returns a string message based on a default message,
// and zero or more strings in the msg slice. If there is one or more
// non-blank messages in the msg slice, then they are concatenated and
// returned. Usually there will be one non-blank message. If there are no
// non-blank messages, then the default message is returned.
func makeMessage(defaultMsg string, msgs []string) string {
	var messages []string
	if len(msgs) > 0 {
		messages = make([]string, 0, len(msgs))
	}
	for _, msg := range msgs {
		msg = strings.TrimSpace(msg)
		if msg != "" {
			messages = append(messages, msg)
		}
	}
	if len(messages) == 0 {
		return defaultMsg
	}
	return strings.Join(messages, " ")
}

// Public returns an error with the message and status.
// The message should not contain any implementation details as
// it may be displayed to a requesting client.
//
// Note that if you attach any key/value pairs to the public
// error using the With method, then that will return a new error that
// is not public, as implementation details may be present in the key/value pairs.
// The cause of the new error, however, will still be public.
func Public(message string, status int) errors.Error {
	return publicStatusError{
		message: message,
		status:  status,
	}
}

// PublicWithCode returns an error with the message, status and code.
// The code can be useful for indicating specific error conditions to
// a requesting client.
//
// The message and code should not contain any implementation details as
// it may be displayed to a requesting client.
//
// Note that if you attach any key/value pairs to the public
// error using the With method, then that will return a new error that
// is not public, as implementation details may be present in the key/value pairs.
// The cause of the new error, however, will still be public.
func PublicWithCode(message string, status int, code string) errors.Error {
	code = strings.TrimSpace(code)
	if code == "" {
		// no code supplied
		return Public(message, status)
	}
	return publicStatusCodeError{
		message: message,
		status:  status,
		code:    code,
	}
}

// IsPublic returns true for errors that indicate
// that their content does not contain sensitive information
// and can be displayed to external clients.
//
// An error is considered public if it implements
// the following interface and its Public method returns true.
//  type publicer interface {
//      Public() bool
//  }
//
// It usually makes sense to obtain the cause of an error first
// before testing to see if it is public. Any public error that
// is wrapped using errors.Wrap, or errors.With will return a
// new error that is no longer public.
//  // get the cause of the error
//  err = errors.Cause(err)
//  if errkind.IsPublic(err) {
//      // ... can provide err.Error() to the client
//  }
func IsPublic(err error) bool {
	if public, ok := err.(publicer); ok {
		return public.Public()
	}
	return false
}

// BadRequest returns an client error that has a status of 400 (bad request).
// The optional msg should not contain sensitive implementation details, as it
// may be returned to the requesting client.
func BadRequest(msg ...string) errors.Error {
	return Public(makeMessage("bad request", msg), http.StatusBadRequest)
}

// Forbidden returns an error that has a status of 403 (forbidden).
// The optional msg should not contain sensitive implementation details, as it
// may be returned to the requesting client.
func Forbidden(msg ...string) errors.Error {
	return Public(makeMessage("forbidden", msg), http.StatusForbidden)
}

// NotImplemented returns an error with a status of not implemented.
// The optional msg should not contain sensitive implementation details, as it
// may be returned to the requesting client.
func NotImplemented(msg ...string) errors.Error {
	return Public(makeMessage("not implemented", msg), http.StatusNotImplemented).With(
		"caller", stack.Caller(1),
	)
}

type temporaryError string

func (t temporaryError) Error() string {
	return string(t)
}

func (t temporaryError) Temporary() bool {
	return true
}

// Temporary returns an error that indicates it is temporary.
func Temporary(msg string) errors.Error {
	return errors.Wrap(temporaryError(msg))
}
