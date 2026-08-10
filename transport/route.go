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
	return strings.Split(s, "/")
}
