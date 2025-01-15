// Test the primitive data formatters.

package table

import (
	"fmt"
	"testing"

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
