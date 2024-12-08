// Annotation types; formatters for them and other types, mostly taking a print context argument and
// returning *skip* when appropriate; parsers for them and other types.

package table

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"go-utils/gpuset"
	"go-utils/hostglob"
	. "sonalyze/common"
)

// Timestamp types.  These types hold an int64 unix-timestamp-since-epoch or second count and
// require a particular kind of formatting.

type DateTimeValue = int64        // yyyy-mm-dd hh:mm
type DateTimeValueOrBlank = int64 // yyyy-mm-dd hh:mm or 16 blanks
type IsoDateTimeValue = int64
type IsoDateTimeOrUnknown = int64 // yyyy-mm-ddThh:mmZhh:mm or "Unknown"
type DateValue = int64            // yyyy-mm-dd
type TimeValue = int64            // hh:mm

// Other types

type DurationValue = int64 // _d_h_m for d(ays) h(ours) m(inutes), rounded to minute, round up on ties
type U32Duration = uint32  // ditto, but different underlying type
type F64Ceil = float64     // float64 that rounds up to integer on output
type U64Div1M = uint64     // uint64 that is scaled by 2^20
type IntOrEmpty = int      // the int value, but "" if zero
type UstrMax30 = Ustr      // the string value but only max 30 first chars in fixed mode

func FormatIntOrEmpty(val IntOrEmpty, _ PrintMods) string {
	if val == 0 {
		return ""
	}
	return strconv.FormatInt(int64(val), 10)
}

func FormatDateValue(val DateValue, ctx PrintMods) string {
	if (ctx&PrintModNoDefaults) != 0 && val == 0 {
		return "*skip*"
	}
	return time.Unix(int64(val), 0).UTC().Format("2006-01-02")
}

func FormatTimeValue(val TimeValue, ctx PrintMods) string {
	if (ctx&PrintModNoDefaults) != 0 && val == 0 {
		return "*skip*"
	}
	return time.Unix(int64(val), 0).UTC().Format("15:04")
}

func FormatUstr(val Ustr, ctx PrintMods) string {
	if (ctx&PrintModNoDefaults) != 0 && val == UstrEmpty {
		return "*skip*"
	}
	return val.String()
}

func FormatUstrMax30(val UstrMax30, ctx PrintMods) string {
	if (ctx&PrintModNoDefaults) != 0 && Ustr(val) == UstrEmpty {
		return "*skip*"
	}
	s := Ustr(val).String()
	if (ctx & PrintModFixed) != 0 {
		// TODO: really the rune length, no?
		if len(s) > 30 {
			return s[:30]
		}
	}
	return s
}

func FormatInt64(val int64, ctx PrintMods) string {
	if (ctx&PrintModNoDefaults) != 0 && val == 0 {
		return "*skip*"
	}
	return fmt.Sprint(val)
}

func FormatUint8(val uint8, ctx PrintMods) string {
	if (ctx&PrintModNoDefaults) != 0 && val == 0 {
		return "*skip*"
	}
	return fmt.Sprint(val)
}

func FormatUint32(val uint32, ctx PrintMods) string {
	if (ctx&PrintModNoDefaults) != 0 && val == 0 {
		return "*skip*"
	}
	return fmt.Sprint(val)
}

func FormatUint64(val uint64, ctx PrintMods) string {
	if (ctx&PrintModNoDefaults) != 0 && val == 0 {
		return "*skip*"
	}
	return fmt.Sprint(val)
}

func FormatInt(val int, ctx PrintMods) string {
	if (ctx&PrintModNoDefaults) != 0 && val == 0 {
		return "*skip*"
	}
	return fmt.Sprint(val)
}

func FormatU64Div1M(val uint64, ctx PrintMods) string {
	val /= 1024 * 1024
	if (ctx&PrintModNoDefaults) != 0 && val == 0 {
		return "*skip*"
	}
	return fmt.Sprint(val)
}

func FormatF64Ceil(val float64, ctx PrintMods) string {
	return FormatInt64(int64(math.Ceil(val)), ctx)
}

func FormatFloat32(val float32, ctx PrintMods) string {
	if (ctx&PrintModNoDefaults) != 0 && val == 0 {
		return "*skip*"
	}
	return strconv.FormatFloat(float64(val), 'g', -1, 32)
}

func FormatFloat64(val float64, ctx PrintMods) string {
	if (ctx&PrintModNoDefaults) != 0 && val == 0 {
		return "*skip*"
	}
	return strconv.FormatFloat(val, 'g', -1, 64)
}

func FormatString(val string, ctx PrintMods) string {
	if (ctx&PrintModNoDefaults) != 0 && val == "" {
		return "*skip*"
	}
	if (ctx & PrintModMax30) != 0 {
		if len(val) > 30 {
			return val[:30]
		}
	}
	return val
}

func FormatStrings(val []string, ctx PrintMods) string {
	if (ctx&PrintModNoDefaults) != 0 && len(val) == 0 {
		return "*skip*"
	}
	sortable := slices.Clone(val)
	slices.Sort(sortable)
	return strings.Join(sortable, ",")
}

func FormatGpuSet(val gpuset.GpuSet, ctx PrintMods) string {
	if (ctx&PrintModNoDefaults) != 0 && val.IsEmpty() {
		return "*skip*"
	}
	return val.String()
}

func FormatBool(val bool, ctx PrintMods) string {
	if (ctx&PrintModNoDefaults) != 0 && !val {
		return "*skip*"
	}
	// These are backwards compatible values.
	if val {
		return "yes"
	}
	return "no"
}

// The DurationValue had two different formats in older code: %2dd%2dh%2dm and %dd%2dh%2dm.  The
// embedded spaces made things line up properly in fixed-format outputs of jobs, and most scripts
// would likely use "duration/sec" etc instead, but the variability in the leading space is weird
// and probably an accident (the "duration" field of the jobs report is the oldest one in the code
// and did not have leading space).  Additionally, in older code there was a difference in rounding
// behavior in different printers (some would round, some would truncate).
//
// The fact that the output format does not correspond to the "duration" option format (which does
// not allow spaces) is also annoying.
//
// Here, we settle on the following interpretation and hope it won't break anything.  For fixed
// output, always use %2dd%2dh%2dm.  For other outputs, always use %dd%dh%dm.  Also, always round to
// the nearerest minute, rounding up on ties.

func FormatDurationValue(secs DurationValue, ctx PrintMods) string {
	if (ctx & PrintModSec) != 0 {
		if (ctx&PrintModNoDefaults) != 0 && secs == 0 {
			return "*skip*"
		}
		return fmt.Sprint(secs)
	}

	if secs%60 >= 30 {
		secs += 30
	}
	minutes := (secs / 60) % 60
	hours := (secs / (60 * 60)) % 24
	days := secs / (60 * 60 * 24)
	if (ctx&PrintModNoDefaults) != 0 && minutes == 0 && hours == 0 && days == 0 {
		return "*skip*"
	}
	if (ctx & PrintModFixed) != 0 {
		return fmt.Sprintf("%2dd%2dh%2dm", days, hours, minutes)
	}
	return fmt.Sprintf("%dd%dh%dm", days, hours, minutes)
}

func FormatU32Duration(secs U32Duration, ctx PrintMods) string {
	return FormatDurationValue(DurationValue(secs), ctx)
}

func FormatDateTimeValue(timestamp DateTimeValue, ctx PrintMods) string {
	if (ctx&PrintModNoDefaults) != 0 && timestamp == 0 {
		return "*skip*"
	}
	// Note, it is part of the API that PrintModSec takes precedence over PrintModIso (this
	// simplifies various other paths).
	if (ctx & PrintModSec) != 0 {
		return fmt.Sprint(timestamp)
	}
	if (ctx & PrintModIso) != 0 {
		return FormatIsoUtc(timestamp)
	}
	return FormatYyyyMmDdHhMmUtc(timestamp)
}

func FormatDateTimeValueOrBlank(val DateTimeValueOrBlank, ctx PrintMods) string {
	// I guess if we want Iso then the blank string should be one longer.
	if val == 0 {
		return "                "
	}
	return FormatDateTimeValue(DateTimeValue(val), ctx)
}

func FormatIsoDateTimeValue(t IsoDateTimeValue, ctx PrintMods) string {
	return FormatDateTimeValue(DateTimeValue(t), ctx|PrintModIso)
}

func FormatIsoDateTimeOrUnknown(t IsoDateTimeOrUnknown, ctx PrintMods) string {
	if t == 0 {
		return "Unknown"
	}
	return FormatDateTimeValue(DateTimeValue(t), ctx|PrintModIso)
}

func FormatYyyyMmDdHhMmUtc(t int64) string {
	return time.Unix(t, 0).UTC().Format("2006-01-02 15:04")
}

func FormatIsoUtc(t int64) string {
	return time.Unix(t, 0).UTC().Format(time.RFC3339)
}

func FormatHostnames(x *Hostnames, ctx PrintMods) string {
	if ctx&PrintModFixed != 0 {
		return x.FormatBrief()
	}
	return x.FormatFull()
}

// This returns error to conform to an interface but never returns non-nil error.
func CvtString2Strings(s string) (any, error) {
	if s == "" {
		return make([]string, 0), nil
	}
	ss := strings.Split(s, ",")
	// Sorted is required by SetCompareStrings
	slices.Sort(ss)
	return ss, nil
}

func CvtString2GpuSet(s string) (any, error) {
	return gpuset.NewGpuSet(s)
}

func CvtString2Ustr(s string) (any, error) {
	return StringToUstr(s), nil
}

func CvtString2UstrMax30(s string) (any, error) {
	return StringToUstr(s), nil
}

func CvtString2IsoDateTimeValue(s string) (any, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return int64(0), err
	}
	return t.Unix(), nil
}

func CvtString2IsoDateTimeOrUnknown(s string) (any, error) {
	return CvtString2IsoDateTimeValue(s)
}

var durationRe = regexp.MustCompile(`^((\d+)[wW])?((\d+)[dD])?((\d+)[hH])?((\d+)[mM])?$`)

// The value is seconds as int64
func CvtString2DurationValue(s string) (any, error) {
	m := durationRe.FindStringSubmatch(s)
	if m == nil {
		seconds, err := strconv.Atoi(s)
		if err != nil {
			return 0, fmt.Errorf("Bad duration %s", s)
		}
		return int64(seconds), nil
	}
	var weeks, days, hours, minutes int
	if m[1] != "" {
		weeks, _ = strconv.Atoi(m[2])
	}
	if m[3] != "" {
		days, _ = strconv.Atoi(m[4])
	}
	if m[5] != "" {
		hours, _ = strconv.Atoi(m[6])
	}
	if m[7] != "" {
		minutes, _ = strconv.Atoi(m[8])
	}
	return (((int64(weeks)*7+int64(days))*24+int64(hours))*60 + int64(minutes)) * 60, nil
}

func CvtString2U32Duration(s string) (any, error) {
	d, err := CvtString2DurationValue(s)
	if err != nil {
		return uint32(0), err
	}
	return uint32(d.(int64)), nil
}

func CvtString2DateTimeValue(s string) (any, error) {
	t, err := time.Parse(time.DateTime, s)
	if err != nil {
		return int64(0), err
	}
	return t.Unix(), nil
}

func CvtString2DateTimeValueOrBlank(s string) (any, error) {
	return CvtString2DateTimeValue(s)
}

func CvtString2DateValue(s string) (any, error) {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return int64(0), err
	}
	return t.Unix(), nil
}

func CvtString2TimeValue(s string) (any, error) {
	t, err := time.Parse(time.TimeOnly, s)
	if err != nil {
		return int64(0), err
	}
	return t.Unix(), nil
}

func CvtString2Bool(s string) (any, error) {
	s = strings.ToLower(s)
	switch s {
	case "true", "yes", "1":
		return true, nil
	case "false", "no", "0":
		return false, nil
	default:
		return nil, fmt.Errorf("Not a boolean value: %s", s)
	}
}

func CvtString2Int(s string) (any, error) {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, err
	}
	return int(i), nil
}

func CvtString2Int64(s string) (any, error) {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func CvtString2Uint8(s string) (any, error) {
	i, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		return nil, err
	}
	return uint8(i), nil
}

func CvtString2Uint32(s string) (any, error) {
	i, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return nil, err
	}
	return uint32(i), nil
}

func CvtString2Uint64(s string) (any, error) {
	i, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func CvtString2Float32(s string) (any, error) {
	i, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return nil, err
	}
	return float32(i), nil
}

func CvtString2Float64(s string) (any, error) {
	i, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func CvtString2Hostnames(s string) (any, error) {
	ps, err := hostglob.SplitMultiPattern(s)
	if err != nil {
		return nil, err
	}
	xs := make([]string, 0)
	for _, p := range ps {
		ss, err := hostglob.ExpandPattern(p)
		if err != nil {
			return nil, err
		}
		xs = append(xs, ss...)
	}
	return xs, nil
}

// true > false
func CompareBool(b1, b2 bool) int {
	if b1 == b2 {
		return 0
	}
	if b1 {
		return 1
	}
	return -1
}

func SetCompareGpuSets(a, b gpuset.GpuSet, op int) bool {
	switch op {
	case opEq:
		return a.Equal(b)
	case opLt:
		return b.HasSubset(a, true)
	case opLe:
		return b.HasSubset(a, false)
	case opGt:
		return a.HasSubset(b, true)
	case opGe:
		return a.HasSubset(b, false)
	default:
		panic("Unknown op")
	}
}

// This is special purpose. `a` comes from the data record, `b` comes from CvtString2Hostnames.

func SetCompareHostnames(a, b *Hostnames, op int) bool {
	switch op {
	case opEq:
		return a.Equal(b)
	case opLt:
		return b.HasSubset(a, true)
	case opLe:
		return b.HasSubset(a, false)
	case opGt:
		return a.HasSubset(b, true)
	case opGe:
		return a.HasSubset(b, false)
	default:
		panic("Unknown op")
	}
}

// This is special purpose.  `a` comes from the data record, `b` comes from CvtString2Strings.
//
// We currently require `b` to be sorted!  We could probably impose some restrictions on `a` with
// suitable types.  There are probably other representations that would be better anyway.

func SetCompareStrings(a, b []string, op int) bool {
	// TODO: assert b is sorted, when running in debug mode
	slices.Sort(a)
Again:
	switch op {
	case opEq:
		if len(a) != len(b) {
			return false
		}
		for k := range a {
			if a[k] != b[k] {
				return false
			}
		}
		return true
	case opLt, opLe:
		bx := 0
		bSkipped := false
		for _, s := range a {
			for bx < len(b) && b[bx] < s {
				bSkipped = true
				bx++
			}
			if bx == len(b) || s != b[bx] {
				return false
			}
			bx++
		}
		if bx < len(b) {
			bSkipped = true
		}
		if op == opLt {
			return bSkipped
		}
		return true
	case opGt:
		op = opLt
		a, b = b, a
		goto Again
	case opGe:
		op = opLe
		a, b = b, a
		goto Again
	default:
		panic("Bad operation")
	}
}
