package percent

import "strconv"

// Percent is a percentage value constrained to the inclusive range [0, 100]
//
//nolint:recvcheck // Percent mixes value and pointer receivers: value for read-only accessors, pointer for UnmarshalJSON
type Percent float64

// NewPercent returns a Percent clamped to [0, 100]
func NewPercent(v float64) Percent {
	switch {
	case v < 0:
		return 0
	case v > 100:
		return 100
	default:
		return Percent(v)
	}
}

func (p Percent) Value() float64 {
	return float64(p)
}

func (p Percent) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatFloat(float64(p), 'f', -1, 64)), nil
}

func (p *Percent) UnmarshalJSON(data []byte) error {
	v, err := strconv.ParseFloat(string(data), 64)
	if err != nil {
		return err
	}
	*p = NewPercent(v)
	return nil
}
