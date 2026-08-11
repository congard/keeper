package transport

import (
	"testing"
)

func TestParseRoute(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Route
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "single segment",
			input:    "a",
			expected: Route{"a"},
		},
		{
			name:     "multiple segments",
			input:    "a/b/c",
			expected: Route{"a", "b", "c"},
		},
		{
			name:     "leading slash",
			input:    "/a/b",
			expected: Route{"a", "b"},
		},
		{
			name:     "trailing slash",
			input:    "a/b/",
			expected: Route{"a", "b"},
		},
		{
			name:     "leading and trailing slash",
			input:    "/a/b/",
			expected: Route{"a", "b"},
		},
		{
			name:     "multiple leading slashes",
			input:    "///a/b",
			expected: Route{"a", "b"},
		},
		{
			name:     "multiple trailing slashes",
			input:    "a/b///",
			expected: Route{"a", "b"},
		},
		{
			name:     "only slashes",
			input:    "/",
			expected: nil,
		},
		{
			name:     "multiple slashes only",
			input:    "///",
			expected: nil,
		},
		{
			name:     "multiple slashes between segments",
			input:    "a//b",
			expected: Route{"a", "b"},
		},
		{
			name:     "multiple slashes between multiple segments",
			input:    "a///b//c////d",
			expected: Route{"a", "b", "c", "d"},
		},
		{
			name:     "multiple slashes with leading and trailing",
			input:    "//a///b//",
			expected: Route{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseRoute(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("ParseRoute(%q) = %v, want %v", tt.input, result, tt.expected)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("ParseRoute(%q) = %v, want %v", tt.input, result, tt.expected)
					return
				}
			}
		})
	}
}

func TestRouteString(t *testing.T) {
	tests := []struct {
		name     string
		route    Route
		expected string
	}{
		{
			name:     "nil route",
			route:    nil,
			expected: "",
		},
		{
			name:     "empty route",
			route:    Route{},
			expected: "",
		},
		{
			name:     "single segment",
			route:    Route{"a"},
			expected: "a",
		},
		{
			name:     "multiple segments",
			route:    Route{"a", "b", "c"},
			expected: "a/b/c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.route.String()
			if result != tt.expected {
				t.Errorf("Route(%v).String() = %q, want %q", tt.route, result, tt.expected)
			}
		})
	}
}

func TestParseRouteRoundTrip(t *testing.T) {
	tests := []string{
		"a",
		"a/b",
		"a/b/c",
		"/a/b",
		"a/b/",
		"/a/b/",
		"a.b/c.d",
		"a-b/c-d",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			route := ParseRoute(input)
			output := route.String()
			// Parse again to verify round-trip
			route2 := ParseRoute(output)
			if len(route) != len(route2) {
				t.Errorf("Round-trip failed for %q: got %v, want %v", input, route2, route)
				return
			}
			for i := range route {
				if route[i] != route2[i] {
					t.Errorf("Round-trip failed for %q: got %v, want %v", input, route2, route)
					return
				}
			}
		})
	}
}
