package xerr

import "fmt"

type SilentTerminationError struct {
	err error
}

func (e *SilentTerminationError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("terminating silently: %s", e.err)
}

func (e *SilentTerminationError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

func TerminateSilently(err error) *SilentTerminationError {
	return &SilentTerminationError{err: err}
}
