// Test the primitive data formatters, parsers, and comparators.

package table

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"go-utils/gpuset"

	. "sonalyze/common"
)

const (
	now = 1732518173          // 2024-11-25T07:02:53Z
	dur = 3600*33 + 60*6 + 38 // 1d 9h 7m, rounded up
)

func TestDataFormatting(t *testing.T) {
	// Try to keep these in the same order as the formatters in data.go.
	// Try to test all print modifiers that are possible for each formatters.

	if s := FormatIntOrEmpty(7, 0); s != "7" {
		t.Fatalf("IntOrEmpty %s", s)
	}
	if s := FormatIntOrEmpty(0, 0); s != "" {
		t.Fatalf("IntOrEmpty %s", s)
	}

	if s := FormatDateValue(now, 0); s != "2024-11-25" {
		t.Fatalf("DateValue %s", s)
	}
	if s := FormatDateValue(0, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("DateValue %s", s)
	}

	if s := FormatTimeValue(now, 0); s != "07:02" {
		t.Fatalf("TimeValue %s", s)
	}
	if s := FormatTimeValue(0, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("TimeValue %s", s)
	}

	if s := FormatUstr(StringToUstr("hello"), 0); s != "hello" {
		t.Fatalf("Ustr %s", s)
	}
	if s := FormatUstr(UstrEmpty, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("Ustr %s", s)
	}

	if s := FormatUstrMax30(StringToUstr("supercalifragilisticexpialidocious"), PrintModFixed); s != "supercalifragilisticexpialidoc" {
		t.Fatalf("UstrMax30 %s", s)
	}
	if s := FormatUstrMax30(StringToUstr("supercalifragilisticexpialidocious"), 0); s != "supercalifragilisticexpialidocious" {
		t.Fatalf("UstrMax30 %s", s)
	}
	if s := FormatUstrMax30(UstrEmpty, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("UstrMax30 %s", s)
	}

	if s := FormatInt64(int64(123456), 0); s != "123456" {
		t.Fatalf("Int64 %s", s)
	}
	if s := FormatInt64(int64(-123456), PrintModNoDefaults); s != "-123456" {
		t.Fatalf("Int64 %s", s)
	}
	if s := FormatInt64(int64(0), PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("Int64 %s", s)
	}

	if s := FormatUint8(127, 0); s != "127" {
		t.Fatalf("Uint8 %s", s)
	}
	if s := FormatUint8(0, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("Uint8 %s", s)
	}

	if s := FormatUint32(12735, 0); s != "12735" {
		t.Fatalf("Uint32 %s", s)
	}
	if s := FormatUint32(0, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("Uint32 %s", s)
	}

	if s := FormatUint64(1273599, 0); s != "1273599" {
		t.Fatalf("Uint64 %s", s)
	}
	if s := FormatUint64(0, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("Uint64 %s", s)
	}

	if s := FormatInt(12735, 0); s != "12735" {
		t.Fatalf("Int %s", s)
	}
	if s := FormatInt(-12735, 0); s != "-12735" {
		t.Fatalf("Int %s", s)
	}
	if s := FormatInt(0, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("Int %s", s)
	}

	// Divides by 2^20
	if s := FormatU64Div1M(127359907, 0); s != "121" {
		t.Fatalf("U64Div1M %s", s)
	}
	if s := FormatU64Div1M(0, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("U64Div1M %s", s)
	}

	if s := FormatF64Ceil(12735.3, 0); s != "12736" {
		t.Fatalf("F64Ceil %s", s)
	}
	if s := FormatF64Ceil(0, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("F64Ceil %s", s)
	}

	if s := FormatFloat32(12735.5, 0); s != "12735.5" {
		t.Fatalf("Float32 %s", s)
	}
	if s := FormatFloat32(0, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("Float32 %s", s)
	}

	if s := FormatFloat64(12735.5, 0); s != "12735.5" {
		t.Fatalf("Float64 %s", s)
	}
	if s := FormatFloat64(0, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("Float64 %s", s)
	}

	if s := FormatString("hi", 0); s != "hi" {
		t.Fatalf("String %s", s)
	}
	if s := FormatString("", 0); s != "" {
		t.Fatalf("String %s", s)
	}
	if s := FormatString("", PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("String %s", s)
	}

	if s := FormatStrings([]string{"a", "b", "c"}, 0); s != "a,b,c" {
		t.Fatalf("Strings %s", s)
	}
	if s := FormatStrings([]string{}, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("Strings %s", s)
	}
	if s := FormatStrings(nil, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("Strings %s", s)
	}

	set, _ := gpuset.NewGpuSet("1,3")
	if s := FormatGpuSet(set, 0); s != "1,3" {
		t.Fatalf("GpuSet %s", s)
	}
	set, _ = gpuset.NewGpuSet("unknown")
	if s := FormatGpuSet(set, 0); s != "unknown" {
		t.Fatalf("GpuSet %s", s)
	}
	if s := FormatGpuSet(gpuset.EmptyGpuSet(), 0); s != "none" {
		t.Fatalf("GpuSet %s", s)
	}
	if s := FormatGpuSet(gpuset.EmptyGpuSet(), PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("GpuSet %s", s)
	}

	if s := FormatBool(true, 0); s != "yes" {
		t.Fatalf("Bool %s", s)
	}
	if s := FormatBool(false, 0); s != "no" {
		t.Fatalf("Bool %s", s)
	}
	if s := FormatBool(false, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("Bool %s", s)
	}

	if s := FormatDurationValue(dur, 0); s != "1d9h7m" {
		t.Fatalf("Duration %s", s)
	}
	if s := FormatDurationValue(dur, PrintModFixed); s != " 1d 9h 7m" {
		t.Fatalf("Duration %s", s)
	}
	if s := FormatDurationValue(dur, PrintModSec); s != fmt.Sprint(dur) {
		t.Fatalf("Duration %s", s)
	}
	if s := FormatDurationValue(0, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("Duration %s", s)
	}

	// U32Duration forwards to DurationValue with all attributes so just do one test
	if s := FormatU32Duration(uint32(dur), 0); s != "1d9h7m" {
		t.Fatalf("U32Duration %s", s)
	}

	if s := FormatDateTimeValue(now, 0); s != "2024-11-25 07:02" {
		t.Fatalf("DateTimeValue %s", s)
	}
	if s := FormatDateTimeValue(now, PrintModSec|PrintModIso); s != fmt.Sprint(now) {
		t.Fatalf("DateTimeValue %s", s)
	}
	if s := FormatDateTimeValue(now, PrintModIso); s != "2024-11-25T07:02:53Z" {
		t.Fatalf("DateTimeValue %s", s)
	}
	if s := FormatDateTimeValue(0, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("DateTimeValue %s", s)
	}

	// DateTimeValueOrBlank is uses DateTimeValue except in the case of 0, when it returns exactly
	// 16 blanks, the length of "2024-11-25 07:02".  The modifier should be ignored.
	if s := FormatDateTimeValueOrBlank(0, PrintModNoDefaults); s != "                " {
		t.Fatalf("DateTimeValueOrBlank <%s>", s)
	}

	// IsoDateTimeValue uses DateTimeValue but with the PrintModIso modifier.
	// No-op - tested above.

	// IsoDateTimeOrUnknown.  The modifier should be ignored.
	if s := FormatIsoDateTimeOrUnknown(0, PrintModNoDefaults); s != "Unknown" {
		t.Fatalf("IsoDateTimeOrUnknown %s", s)
	}
}

func TestCompareBool(t *testing.T) {
	if CompareBool(false, false) != 0 {
		t.Fatal("Bool")
	}
	if CompareBool(true, true) != 0 {
		t.Fatal("Bool")
	}
	if CompareBool(false, true) != -1 {
		t.Fatal("Bool")
	}
	if CompareBool(true, false) != 1 {
		t.Fatal("Bool")
	}
}

func TestCompareGpuSets(t *testing.T) {
	// Just basic stuff, since the implementation defers to subset operations that are tested
	// elsewhere.
	as, _ := gpuset.NewGpuSet("1,2")
	bs, _ := gpuset.NewGpuSet("1,2,3")
	cs, _ := gpuset.NewGpuSet("2,3")
	if !SetCompareGpuSets(as, as, opEq) {
		t.Fatal("GpuSet")
	}
	if SetCompareGpuSets(as, bs, opEq) {
		t.Fatal("GpuSet")
	}
	if !SetCompareGpuSets(as, bs, opLt) {
		t.Fatal("GpuSet")
	}
	if !SetCompareGpuSets(as, as, opLe) {
		t.Fatal("GpuSet")
	}
	if !SetCompareGpuSets(as, bs, opLe) {
		t.Fatal("GpuSet")
	}
	if SetCompareGpuSets(as, cs, opLe) {
		t.Fatal("GpuSet")
	}
}

func TestSetCompareStrings(t *testing.T) {
	as := []string{"a", "b", "c"}
	bs := []string{"a", "b", "c", "d"}
	cs := []string{"a", "b", "c", "e"}
	if !SetCompareStrings(as, as, opEq) {
		t.Fatal("Equal")
	}
	if SetCompareStrings(as, as, opLt) {
		t.Fatal("Less")
	}
	if !SetCompareStrings(as, as, opLe) {
		t.Fatal("LessOrEqual")
	}
	if SetCompareStrings(as, as, opGt) {
		t.Fatal("Greater")
	}
	if !SetCompareStrings(as, as, opGe) {
		t.Fatal("GreaterOrEqual")
	}

	if SetCompareStrings(as, bs, opEq) {
		t.Fatal("Equal")
	}
	if !SetCompareStrings(as, bs, opLt) {
		t.Fatal("Less")
	}
	if !SetCompareStrings(as, bs, opLe) {
		t.Fatal("LessOrEqual")
	}
	if SetCompareStrings(as, bs, opGt) {
		t.Fatal("Greater")
	}
	if SetCompareStrings(as, bs, opGe) {
		t.Fatal("GreaterOrEqual")
	}

	if SetCompareStrings(cs, bs, opEq) {
		t.Fatal("Equal")
	}
	if SetCompareStrings(cs, bs, opLt) {
		t.Fatal("Less")
	}
	if SetCompareStrings(cs, bs, opLe) {
		t.Fatal("LessOrEqual")
	}
	if SetCompareStrings(cs, bs, opGt) {
		t.Fatal("Greater")
	}
	if SetCompareStrings(cs, bs, opGe) {
		t.Fatal("GreaterOrEqual")
	}

	// A little harder, this uncovered a bug
	ds := []string{"b", "d"}
	if !SetCompareStrings(ds, bs, opLt) {
		t.Fatal("Less")
	}
	if !SetCompareStrings(ds, bs, opLe) {
		t.Fatal("Less")
	}
}

func check(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func shouldfail(t *testing.T, err error) {
	if err == nil {
		t.Fatal("Expected error")
	}
}

func same(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Fatal("Should be equal")
	}
}

func TestCvt2Strings(t *testing.T) {
	// No actual error cases here
	xs, err := CvtString2Strings("")
	check(t, err)
	same(t, xs, []string{})
	xs, err = CvtString2Strings("a")
	check(t, err)
	same(t, xs, []string{"a"})
	xs, err = CvtString2Strings("a,b,c,d")
	check(t, err)
	same(t, xs, []string{"a", "b", "c", "d"})
}

func TestCvt2GpuSet(t *testing.T) {
	// "" is not a valid GPU set
	_, err := CvtString2GpuSet("")
	shouldfail(t, err)

	xs, err := CvtString2GpuSet("none")
	check(t, err)
	same(t, xs, gpuset.EmptyGpuSet())

	xs, err = CvtString2GpuSet("unknown")
	check(t, err)
	same(t, xs, gpuset.UnknownGpuSet())

	xs, err = CvtString2GpuSet("1,2,3")
	check(t, err)
	s := gpuset.EmptyGpuSet()
	s, err = gpuset.Adjoin(s, 1, 2, 3)
	check(t, err)
	same(t, xs, s)
}

func TestCvt2Misc(t *testing.T) {
	// No actual error cases for Ustr
	x, err := CvtString2Ustr("hi there")
	check(t, err)
	same(t, x, StringToUstr("hi there"))

	// Does *not* chop
	x, err = CvtString2UstrMax30("supercalifragilisticexpialidocious")
	check(t, err)
	same(t, x, StringToUstr("supercalifragilisticexpialidocious"))

	x, err = CvtString2Bool("1")
	check(t, err)
	same(t, x, true)

	x, err = CvtString2Bool("yes")
	check(t, err)
	same(t, x, true)

	x, err = CvtString2Bool("true")
	check(t, err)
	same(t, x, true)

	x, err = CvtString2Bool("tRUe")
	check(t, err)
	same(t, x, true)

	x, err = CvtString2Bool("0")
	check(t, err)
	same(t, x, false)

	x, err = CvtString2Bool("no")
	check(t, err)
	same(t, x, false)

	x, err = CvtString2Bool("nO")
	check(t, err)
	same(t, x, false)

	x, err = CvtString2Bool("false")
	check(t, err)
	same(t, x, false)

	_, err = CvtString2Bool("maybe")
	shouldfail(t, err)

	x, err = CvtString2Int("-312")
	check(t, err)
	same(t, x, int(-312))

	_, err = CvtString2Int("hello")
	shouldfail(t, err)

	x, err = CvtString2Int64("-312")
	check(t, err)
	same(t, x, int64(-312))

	_, err = CvtString2Int64("hello")
	shouldfail(t, err)

	x, err = CvtString2Uint8("114")
	check(t, err)
	same(t, x, uint8(114))

	_, err = CvtString2Uint8("312")
	shouldfail(t, err)

	x, err = CvtString2Uint32("114")
	check(t, err)
	same(t, x, uint32(114))

	_, err = CvtString2Uint32("1234567890123")
	shouldfail(t, err)

	x, err = CvtString2Uint64("114")
	check(t, err)
	same(t, x, uint64(114))

	_, err = CvtString2Uint64("hello")
	shouldfail(t, err)

	x, err = CvtString2Float32("114.5")
	check(t, err)
	same(t, x, float32(114.5))

	_, err = CvtString2Float32("14f")
	shouldfail(t, err)

	x, err = CvtString2Float64("114.5")
	check(t, err)
	same(t, x, float64(114.5))

	_, err = CvtString2Float64("14.1.2")
	shouldfail(t, err)
}

func TestCvt2DateTime(t *testing.T) {
	s := "2025-02-18T19:13:27+01:00"
	x, err := CvtString2IsoDateTimeValue(s)
	check(t, err)
	w, _ := time.Parse(time.RFC3339, s)
	v := w.Unix()
	same(t, x, v)

	_, err = CvtString2IsoDateTimeValue(strings.Replace(s, "T", " ", 1))
	shouldfail(t, err)

	// CvtString2IsoDateTimevalueOrUnknown just calls 2IsoDateTimeValue

	y := "2025-02-18 19:13:27" // implied localtime
	x, err = CvtString2DateTimeValue(y)
	check(t, err)
	w, _ = time.Parse(time.DateTime, y)
	v = w.Unix()
	same(t, x, v)

	_, err = CvtString2DateTimeValue(s)
	shouldfail(t, err)

	// CvtString2DateTimeValueOrBlank just calls 2DateTimeValue

	z := "2025-02-18"
	x, err = CvtString2DateValue(z)
	check(t, err)
	w, _ = time.Parse(time.DateOnly, z)
	v = w.Unix()
	same(t, x, v)

	_, err = CvtString2DateValue(z + " ")
	shouldfail(t, err)

	u := "12:14:05"
	x, err = CvtString2TimeValue(u)
	check(t, err)
	w, _ = time.Parse(time.TimeOnly, u)
	v = w.Unix()
	same(t, x, v)

	_, err = CvtString2TimeValue("12:14")
	shouldfail(t, err)
}

func TestCvt2DurationValue(t *testing.T) {
	x, err := CvtString2DurationValue("1h")
	check(t, err)
	same(t, x, int64(3600))

	x, err = CvtString2DurationValue("1h2m")
	check(t, err)
	same(t, x, int64(3600+120))

	x, err = CvtString2DurationValue("2d")
	check(t, err)
	same(t, x, int64(3600*24*2))

	x, err = CvtString2DurationValue("3w")
	check(t, err)
	same(t, x, int64(3600*24*7*3))

	x, err = CvtString2DurationValue("312")
	check(t, err)
	same(t, x, int64(312))

	_, err = CvtString2DurationValue("1d2w")
	shouldfail(t, err)

	x, err = CvtString2U32Duration("1h")
	check(t, err)
	same(t, x, uint32(3600))

	_, err = CvtString2U32Duration("1d2w")
	shouldfail(t, err)
}
