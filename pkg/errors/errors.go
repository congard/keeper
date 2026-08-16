package errors

import "errors"

// IsErrorOfType reports whether err is of type T, using errors.As.
// It is a generic convenience wrapper around the standard library's errors.As.
//
// Note: pass a pointer type as the generic argument, e.g.
// IsErrorOfType[*MyError](err), not IsErrorOfType[MyError](err).
func IsErrorOfType[T error](err error) bool {
	var target T
	return errors.As(err, &target)
}
