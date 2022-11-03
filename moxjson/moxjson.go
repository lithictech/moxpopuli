package moxjson

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

func NewPrettyEncoder(w io.Writer) *json.Encoder {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc
}

type Path []interface{}

func (p Path) String() string {
	lines := make([]string, len(p))
	for i, x := range p {
		if ss, ok := x.(fmt.Stringer); ok {
			lines[i] = ss.String()
		} else if x == nil {
			panic("should never get a nil path part")
		} else {
			lines[i] = fmt.Sprintf("%v", x)
		}
	}
	return strings.Join(lines, ".")
}

func ParsePath(s string) Path {
	parts := strings.Split(s, ".")
	path := make(Path, len(parts))
	for i, part := range parts {
		if jqInd.MatchString(part) {
			v, err := strconv.Atoi(part[1 : len(part)-1])
			if err != nil {
				panic("should never hit this, for " + part)
			}
			path[i] = v
		} else if numInd.MatchString(part) {
			v, err := strconv.Atoi(part)
			if err != nil {
				panic("should never hit this, for " + part)
			}
			path[i] = v
		} else if len(part) > 0 {
			path[i] = part
		} else {
			path[i] = ""
		}
	}
	return path
}

var jqInd = regexp.MustCompile("^\\[\\d+\\]$")
var numInd = regexp.MustCompile("^\\d+$")

func Get(o interface{}, path Path) (interface{}, error) {
	subject := o
	for i, pathPart := range path {
		if pathInt, ok := pathPart.(int); ok {
			sli, ok := subject.([]interface{})
			if !ok {
				return subject, fmt.Errorf("get: expected []interface{} at %s, got %v", path[:i+1], subject)
			}
			subject = sli[pathInt]
		} else if pathStr, ok := pathPart.(string); ok {
			mp, ok := subject.(map[string]interface{})
			if !ok {
				return subject, fmt.Errorf("get: expected map[string]interface{} at %s, got %v", path[:i+1], subject)
			}
			subject = mp[pathStr]
		} else {
			return nil, fmt.Errorf("get: got invalid Path element, should only be int or string: %s", path)
		}
	}
	return subject, nil
}

func Set(o, value interface{}, path Path) error {
	host, err := Get(o, path[:len(path)-1])
	if err != nil {
		return err
	}
	lastPathPart := path[len(path)-1]
	if pathInt, ok := lastPathPart.(int); ok {
		sli, ok := host.([]interface{})
		if !ok {
			return fmt.Errorf("set: expected []interface{} at %s, got %v", path, host)
		}
		sli[pathInt] = value
	} else if pathStr, ok := lastPathPart.(string); ok {
		mp, ok := host.(map[string]interface{})
		if !ok {
			return fmt.Errorf("set: expected map[string]interface{} at %s, got %v", path, host)
		}
		mp[pathStr] = value
	} else {
		return fmt.Errorf("set: got invalid Path element, should only be int or string: %s", path)
	}
	return nil
}
