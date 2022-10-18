package timestring

import (
	"github.com/rickb777/date/period"
	"time"
)

type TimeString struct {
	S string
	U int64
	T time.Time
	P period.Period
}

func From(layout, s string) TimeString {
	t := mustTime(time.Parse(layout, s))
	return TimeString{T: t, S: s, U: t.UnixNano()}
}

func FromPeriod(s string) TimeString {
	p := mustPeriod(period.Parse(s))
	return TimeString{P: p, S: s, U: p.DurationApprox().Nanoseconds()}
}

func mustTime(t time.Time, err error) time.Time {
	if err != nil {
		panic(err)
	}
	return t
}

func mustPeriod(p period.Period, err error) period.Period {
	if err != nil {
		panic(err)
	}
	return p
}

func Many(layout string, ss ...string) []TimeString {
	r := make([]TimeString, len(ss))
	for i, s := range ss {
		r[i] = From(layout, s)
	}
	return r
}

func MinMax(ts []TimeString) (TimeString, TimeString) {
	if len(ts) == 0 {
		return TimeString{}, TimeString{}
	}
	minidx, maxidx := 0, 0
	minval, maxval := ts[0].U, ts[0].U
	for i, t := range ts[1:] {
		if t.U < minval {
			minval = t.U
			minidx = i + 1
		}
		if t.U > maxval {
			maxval = t.U
			maxidx = i + 1
		}
	}
	return ts[minidx], ts[maxidx]
}
