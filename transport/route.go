package transport

import (
	"strings"
)

type RouteSegment = string

type Route []RouteSegment

func (r Route) String() string {
	return strings.Join(r, "/")
}

func ParseRoute(s string) Route {
	if s == "" {
		return nil
	}

	// remove leading and trailing slashes
	s = strings.Trim(s, "/")
	if s == "" {
		return nil
	}

	// normalize: deduplicate slashes and convert to Route in one pass
	route := make(Route, 0)
	var b strings.Builder
	prevSlash := false

	for _, r := range s {
		if r == '/' {
			if !prevSlash {
				route = append(route, b.String())
				b.Reset()
				prevSlash = true
			}
		} else {
			b.WriteRune(r)
			prevSlash = false
		}
	}

	route = append(route, b.String())

	return route
}
