// Test the low-level data formatters including *skip* logic

package table

import (
	"fmt"
	"testing"

	"go-utils/gpuset"
)

const (
	now = 1732518173          // 2024-11-25T07:02:53Z
	dur = 3600*33 + 60*6 + 38 // 1d 9h 7m, rounded up
)

func TestDataFormatting(t *testing.T) {
	if s := FormatDateValue(now, 0); s != "2024-11-25" {
		t.Fatalf("DateValue %s", s)
	}
	if s := FormatTimeValue(now, 0); s != "07:02" {
		t.Fatalf("TimeValue %s", s)
	}

	if s := FormatIntOrEmpty(7, 0); s != "7" {
		t.Fatalf("IntOrEmpty %s", s)
	}
	if s := FormatIntOrEmpty(0, 0); s != "" {
		t.Fatalf("IntOrEmpty %s", s)
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

	if s := FormatInt64(int64(123456), 0); s != "123456" {
		t.Fatalf("Int64 %s", s)
	}
	if s := FormatInt64(int64(-123456), PrintModNoDefaults); s != "-123456" {
		t.Fatalf("Int64 %s", s)
	}
	if s := FormatInt64(int64(0), PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("Int64 %s", s)
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

	if s := FormatBool(true, 0); s != "yes" {
		t.Fatalf("Bool %s", s)
	}
	if s := FormatBool(false, 0); s != "no" {
		t.Fatalf("Bool %s", s)
	}
	if s := FormatBool(false, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("Bool %s", s)
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
}
