package errors

import (
	"fmt"

	"errors"
)

type Errors struct {
	HaveFatalErrors bool
	Errors          []error
}

func FromErrNonFatal(err error) *Errors {
	if err == nil {
		return nil
	}
	return &Errors{
		HaveFatalErrors: false,
		Errors:          []error{err},
	}
}

func FromErr(err error) *Errors {
	if err == nil {
		return nil
	}
	return &Errors{
		HaveFatalErrors: true,
		Errors:          []error{err},
	}
}

func Fatal(err string) *Errors {
	return &Errors{
		HaveFatalErrors: true,
		Errors:          []error{errors.New(err)},
	}
}

func Fatalf(format string, args ...interface{}) *Errors {
	return &Errors{
		HaveFatalErrors: true,
		Errors:          []error{fmt.Errorf(format, args)},
	}
}

func Error(err string) *Errors {
	return &Errors{
		HaveFatalErrors: false,
		Errors:          []error{errors.New(err)},
	}
}

func Errorf(format string, args ...interface{}) *Errors {
	return &Errors{
		HaveFatalErrors: false,
		Errors:          []error{fmt.Errorf(format, args)},
	}
}

func (e *Errors) AddFatal(err error) *Errors {
	if err == nil {
		return e
	}
	e.HaveFatalErrors = true
	e.Errors = append(e.Errors, err)
	return e
}

func (e *Errors) Add(err error) *Errors {
	if err == nil {
		return e
	}
	e.Errors = append(e.Errors, err)
	return e
}

func (e *Errors) Addf(format string, args ...interface{}) *Errors {
	e.Errors = append(e.Errors, fmt.Errorf(format, args))
	return e
}

func (e *Errors) Merge(e2 *Errors) *Errors {
	if e2 == nil {
		return e
	}
	if !e.HaveFatalErrors {
		e.HaveFatalErrors = e2.HaveFatalErrors
	}

	e.Errors = append(e.Errors, e2.Errors...)
	return e
}
