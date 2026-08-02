package sender

import "time"

//nolint:recvcheck // UnmarshalJSON requires pointer receiver per json.Unmarshaler interface
type StatusType int

const (
	StatusTypeOk StatusType = iota
	StatusTypeError
)

const (
	statusStringOk    = "ok"
	statusStringError = "error"
)

func (st StatusType) String() string {
	switch st {
	case StatusTypeOk:
		return statusStringOk
	case StatusTypeError:
		return statusStringError
	default:
		return "unknown"
	}
}

func (st StatusType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + st.String() + `"`), nil
}

func (st *StatusType) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case `"` + statusStringOk + `"`:
		*st = StatusTypeOk
	case `"` + statusStringError + `"`:
		*st = StatusTypeError
	default:
		*st = StatusTypeOk
	}
	return nil
}

type Status struct {
	Type      StatusType `json:"type"`
	Timestamp time.Time  `json:"timestamp"`
	Message   string     `json:"message"`
}

func NewStatus(typ StatusType, message string) Status {
	return Status{
		Type:      typ,
		Timestamp: time.Now(),
		Message:   message,
	}
}

func OkStatus() Status {
	return NewStatus(StatusTypeOk, "ok")
}

func ErrStatus(message string) Status {
	return NewStatus(StatusTypeError, message)
}
