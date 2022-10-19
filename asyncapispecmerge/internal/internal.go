package internal

import (
	"github.com/lithictech/moxpopuli/asyncapispec"
	"github.com/lithictech/moxpopuli/moxio"
	"strings"
)

type MergeInput struct {
	Spec          asyncapispec.Specification
	EventIterator moxio.Iterator
	ExampleLimit  *int
}

func LinesToHeaderNames(raw string) map[string]struct{} {
	lines := strings.Split(raw, "\n")
	result := make(map[string]struct{}, len(lines))
	for _, h := range lines {
		result[CanonicalHeader(h)] = struct{}{}
	}
	return result
}

func CanonicalHeader(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
