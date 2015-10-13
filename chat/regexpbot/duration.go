package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	durRE                  = regexp.MustCompile(`(?i)^[^\d]*(\d+)([a-z]+)?`)
	defaultDuration uint64 = 60000000000
	units                  = map[string]uint64{
		"s":       1000000000,
		"sec":     1000000000,
		"secs":    1000000000,
		"second":  1000000000,
		"seconds": 1000000000,
		"m":       60000000000,
		"min":     60000000000,
		"mins":    60000000000,
		"minute":  60000000000,
		"minutes": 60000000000,
		"h":       3600000000000,
		"hr":      3600000000000,
		"hrs":     3600000000000,
		"hour":    3600000000000,
		"hours":   3600000000000,
		"d":       86400000000000,
		"day":     86400000000000,
		"days":    86400000000000,
	}
)

func init() {
	durRE.Longest()
}

func parseDuration(str, defunit string) uint64 {
	var dur uint64
	m := durRE.FindStringSubmatch(str)
	if len(m) == 0 {
		return defaultDuration
	}

	var err error
	dur, err = strconv.ParseUint(m[1], 10, 64)
	if err != nil {
		return defaultDuration
	}

	var u uint64
	var ok bool
	if len(m[2]) > 0 {
		u, ok = units[strings.ToLower(m[2])]
		if !ok {
			u = units["min"]
		}
	} else {
		u = units["min"]
	}

	return dur * u
}

func humanizeDuration(dur uint64) string {
	var str string
	d := dur / 86400000000000
	if d > 0 {
		str = fmt.Sprintf("%d days", d)
		dur -= d * 86400000000000
	}
	h := dur / 3600000000000
	if h > 0 {
		str = fmt.Sprintf("%s%d hours ", str, h)
		dur -= h * 3600000000000
	}
	m := dur / 60000000000
	if m > 0 {
		str = fmt.Sprintf("%s%d minutes ", str, m)
		dur -= m * 60000000000
	}
	s := dur / 1000000000
	if s > 0 || str == "" {
		str = fmt.Sprintf("%s%d seconds", str, s)
	}

	return strings.TrimSpace(str)
}
