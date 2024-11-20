package common

import (
	"fmt"
	"regexp"
	"strconv"
)

// Duration parsing.  The format is WwDdHhMm for weeks/days/hours/minutes, all parts are optional
// and the default value is zero.  We return seconds because that's what everyone wants.

var durationRe = regexp.MustCompile(`^(?:(\d+)w)?(?:(\d+)d)?(?:(\d+)h)?(?:(\d+)m)?$`)

func DurationToSeconds(option, s string) (int64, error) {
	if matches := durationRe.FindStringSubmatch(s); matches != nil {
		var weeks, days, hours, minutes int64
		var x1, x2, x3, x4 error
		if matches[1] != "" {
			weeks, x1 = strconv.ParseInt(matches[1], 10, 64)
		}
		if matches[2] != "" {
			days, x2 = strconv.ParseInt(matches[2], 10, 64)
		}
		if matches[3] != "" {
			hours, x3 = strconv.ParseInt(matches[3], 10, 64)
		}
		if matches[4] != "" {
			minutes, x4 = strconv.ParseInt(matches[4], 10, 64)
		}
		if x1 != nil || x2 != nil || x3 != nil || x4 != nil {
			return 0, fmt.Errorf("Invalid %s specifier, try -h", option)
		}
		return (((weeks*7+days)*24+hours)*60 + minutes) * 60, nil
	}
	return 0, fmt.Errorf("Invalid %s specifier, try -h", option)
}
