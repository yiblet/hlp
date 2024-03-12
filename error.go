package main

import "fmt"

type terminateSilentlyError struct {
	err error
}

func (e *terminateSilentlyError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("terminating silently: %s", e.err)
}

func (e *terminateSilentlyError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

func terminateSilently(err error) *terminateSilentlyError {
	return &terminateSilentlyError{err}
}
