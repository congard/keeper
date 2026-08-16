package transport

import "fmt"

type UnexpectedResponseTypeError struct {
	Expected string
	Actual   string
}

func (e *UnexpectedResponseTypeError) Error() string {
	return fmt.Sprintf("unexpected response type: expected %s, got %s", e.Expected, e.Actual)
}
