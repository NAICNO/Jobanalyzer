package table

import (
	"fmt"
	"reflect"
	"testing"

	"go-utils/gpuset"
	uslices "go-utils/slices"

	. "sonalyze/common"
)

// Structure traversal tests

type S1 struct {
	x int `desc:"x field" alias:"xx"`
	*T1
}

type T1 struct {
	y int `desc:"y field" alias:"yy"`
	*U1
}

type U1 struct {
	z int `desc:"z field" alias:"zz"`
}

func TestFormatting1(t *testing.T) {
	v1 := S1{x: 10, T1: &T1{y: 20, U1: &U1{z: 30}}}
	fs := DefineTableFromTags(reflect.TypeOf(v1), nil)
	ss := uslices.Map([]string{"x", "y", "z"}, func(s string) string {
		return fs[s].Fmt(&v1, PrintMods(0))
	})
	if !reflect.DeepEqual(ss, []string{"10", "20", "30"}) {
		t.Fatal(ss)
	}
	ss = uslices.Map([]string{"xx", "yy", "zz"}, func(s string) string {
		return fs[s].Fmt(&v1, PrintMods(0))
	})
	if !reflect.DeepEqual(ss, []string{"10", "20", "30"}) {
		t.Fatal(ss)
	}
}

type S2 struct {
	x int `desc:"x field"`
	T2
}

type T2 struct {
	y int `desc:"y field"`
	U2
}

type U2 struct {
	z int `desc:"z field"`
}

func TestFormatting2(t *testing.T) {
	v1 := S2{x: 10, T2: T2{y: 20, U2: U2{z: 30}}}
	fs := DefineTableFromTags(reflect.TypeOf(v1), nil)
	ss := uslices.Map([]string{"x", "y", "z"}, func(s string) string {
		return fs[s].Fmt(&v1, PrintMods(0))
	})
	if !reflect.DeepEqual(ss, []string{"10", "20", "30"}) {
		t.Fatal(ss)
	}
}

type SFS = SimpleFormatSpec

func TestFormatting3(t *testing.T) {
	v1 := S2{x: 10, T2: T2{y: 20, U2: U2{z: 30}}}
	fs := DefineTableFromMap(
		reflect.TypeFor[S2](),
		map[string]any{
			"x": SFS{"x field", "xx"},
			"y": SFS{"y field", "yy"},
			"z": SFS{"z field", "zz"},
		},
	)
	ss := uslices.Map([]string{"x", "y", "z"}, func(s string) string {
		return fs[s].Fmt(&v1, PrintMods(0))
	})
	if !reflect.DeepEqual(ss, []string{"10", "20", "30"}) {
		t.Fatal(ss)
	}
	ss = uslices.Map([]string{"xx", "yy", "zz"}, func(s string) string {
		return fs[s].Fmt(&v1, PrintMods(0))
	})
	if !reflect.DeepEqual(ss, []string{"10", "20", "30"}) {
		t.Fatal(ss)
	}
}

// Field formatter tests.  Also see data_test.go.

// now and dur are defined in data_test.go

func TestReflectDateTimeValue(t *testing.T) {
	type s0 struct {
		V DateTimeValue
	}
	s0f := reflectTypeFormatter(0, 0, reflect.TypeFor[DateTimeValue]())
	if s := s0f(s0{now}, 0); s != "2024-11-25 07:02" {
		t.Fatalf("DateTimeValue %s", s)
	}
	if s := s0f(s0{0}, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("DateTimeValue %s", s)
	}
}

func TestReflectDateTimeValueOrBlank(t *testing.T) {
	type s1 struct {
		V DateTimeValueOrBlank
	}
	s1f := reflectTypeFormatter(0, 0, reflect.TypeFor[DateTimeValueOrBlank]())
	if s := s1f(s1{now}, 0); s != "2024-11-25 07:02" {
		t.Fatalf("DateTimeValueOrBlank %s", s)
	}
	if s := s1f(s1{0}, 0); s != "                " {
		t.Fatalf("DateTimeValueOrBlank %s", s)
	}
	if s := s1f(s1{0}, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("DateTimeValueOrBlank %s", s)
	}
}

func TestReflectIsoDateTimeOrUnknown(t *testing.T) {
	type s2 struct {
		V IsoDateTimeOrUnknown
	}
	s2f := reflectTypeFormatter(0, 0, reflect.TypeFor[IsoDateTimeOrUnknown]())
	if s := s2f(s2{now}, 0); s != "2024-11-25T07:02:53Z" {
		t.Fatalf("IsoDateTimeOrUnknown %s", s)
	}
	if s := s2f(s2{now}, PrintModSec); s != fmt.Sprint(now) {
		t.Fatalf("IsoDateTimeOrUnknown %s", s)
	}
	if s := s2f(s2{0}, PrintModSec); s != "Unknown" { // "Unknown" wins over "/sec"
		t.Fatalf("IsoDateTimeOrUnknown %s", s)
	}
	if s := s2f(s2{0}, PrintModNoDefaults); s != "Unknown" { // "Unknown" wins over "nodefaults"
		t.Fatalf("IsoDateTimeOrUnknown %s", s)
	}
}

func TestReflectDurationValue(t *testing.T) {
	type s3 struct {
		V DurationValue
	}
	s3f := reflectTypeFormatter(0, 0, reflect.TypeFor[DurationValue]())
	if s := s3f(s3{dur}, 0); s != "1d9h7m" {
		t.Fatalf("DurationValue %s", s)
	}
	if s := s3f(s3{dur}, PrintModSec); s != fmt.Sprint(dur) {
		t.Fatalf("DurationValue %s", s)
	}
	if s := s3f(s3{0}, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("DurationValue %s", s)
	}
}

func TestReflectGpuSet(t *testing.T) {
	type s4 struct {
		V gpuset.GpuSet
	}
	s13, err := gpuset.NewGpuSet("1,3")
	if err != nil {
		t.Fatal(err)
	}
	s4f := reflectTypeFormatter(0, 0, reflect.TypeFor[gpuset.GpuSet]())
	if s := s4f(s4{s13}, 0); s != "1,3" {
		t.Fatalf("GpuSet %s", s)
	}
	if s := s4f(s4{gpuset.EmptyGpuSet()}, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("GpuSet %s", s)
	}
}

func TestReflectUstrMax30(t *testing.T) {
	type st struct {
		V UstrMax30
	}
	sf := reflectTypeFormatter(0, 0, reflect.TypeFor[UstrMax30]())
	d := st{UstrMax30(StringToUstr("supercallifragilisticexpialidocious"))}
	if s := sf(d, 0); s != "supercallifragilisticexpialidocious" {
		t.Fatalf("UstrMax30 %s", s)
	}
	if s := sf(d, PrintModFixed); s != "supercallifragilisticexpialido" {
		t.Fatalf("UstrMax30 %s", s)
	}
	if s := sf(st{UstrMax30(UstrEmpty)}, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("UstrMax30 %s", s)
	}
}

func TestReflectSlice(t *testing.T) {
	type st struct {
		V []string
	}
	sf := reflectTypeFormatter(0, 0, reflect.TypeFor[[]string]())
	if s := sf(st{[]string{"ho", "hi", "z", "a"}}, 0); s != "a,hi,ho,z" {
		t.Fatalf("[]string %s", s)
	}
	if s := sf(st{[]string{}}, 0); s != "" {
		t.Fatalf("[]string %s", s)
	}
	if s := sf(st{[]string{}}, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("[]string %s", s)
	}
}

type tstringer int

func (t tstringer) String() string {
	if t < 0 {
		return ""
	}
	return fmt.Sprintf("+%d+", int(t))
}

func TestReflectStringer(t *testing.T) {
	type st struct {
		V tstringer
	}
	sf := reflectTypeFormatter(0, 0, reflect.TypeFor[tstringer]())
	if s := sf(st{33}, PrintModNoDefaults); s != "+33+" {
		t.Fatalf("stringer %s", s)
	}
	if s := sf(st{-1}, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("stringer %s", s)
	}
}

func TestReflectBool(t *testing.T) {
	type st struct {
		V bool
	}
	sf := reflectTypeFormatter(0, 0, reflect.TypeFor[bool]())
	if s := sf(st{true}, 0); s != "yes" {
		t.Fatalf("bool %s", s)
	}
	if s := sf(st{false}, 0); s != "no" {
		t.Fatalf("bool %s", s)
	}
	if s := sf(st{false}, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("bool %s", s)
	}
}

func TestReflectInt64(t *testing.T) {
	type st struct {
		V int64
	}

	// Plain type, plain printing
	sf := reflectTypeFormatter(0, 0, reflect.TypeFor[int64]())
	if s := sf(st{37}, 0); s != "37" {
		t.Fatalf("int64 %s", s)
	}
	if s := sf(st{0}, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("int64 %s", s)
	}

	// For the following, NoDefaults takes precedence over the formatting directive

	// Scaled
	sf = reflectTypeFormatter(0, FmtDivideBy1M, reflect.TypeFor[int64]())
	if s := sf(st{123456789}, 0); s != "117" {
		t.Fatalf("int64 scaled %s", s)
	}
	if s := sf(st{0}, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("int64 scaled %s", s)
	}

	// As DateTimeValue
	sf = reflectTypeFormatter(0, FmtDateTimeValue, reflect.TypeFor[int64]())
	if s := sf(st{now}, 0); s != "2024-11-25 07:02" {
		t.Fatalf("int64 datetime %s", s)
	}
	if s := sf(st{0}, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("int64 datetime %s", s)
	}

	// As IsoDateTimeValue
	sf = reflectTypeFormatter(0, FmtIsoDateTimeValue, reflect.TypeFor[int64]())
	if s := sf(st{now}, 0); s != "2024-11-25T07:02:53Z" {
		t.Fatalf("int64 isodatetime %s", s)
	}
	if s := sf(st{0}, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("int64 isodatetime %s", s)
	}

	// As Duration
	sf = reflectTypeFormatter(0, FmtDurationValue, reflect.TypeFor[int64]())
	if s := sf(st{dur}, 0); s != "1d9h7m" {
		t.Fatalf("int64 duration %s", s)
	}
	if s := sf(st{0}, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("int64 duration %s", s)
	}
}

// int, int8, int16, int32, uint, uint8, uint16, uint32, uint64 all have the same logic, which is
// the same as the base case logic of int64.  The only complication is the scaling values.
//
// Notably, the default value test applies *before* scaling.  This is probably not desirable.
// Contrast the float case below, where the ceiling operation is performed first and we then check
// for the default value.

func testReflectIntish[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64](
	t *testing.T,
	val, scaled T,
) {
	type st struct {
		V T
	}

	// Plain printing
	sf := reflectTypeFormatter(0, 0, reflect.TypeFor[T]())
	if s := sf(st{val}, 0); s != fmt.Sprint(val) {
		t.Fatalf("intish %s", s)
	}
	if s := sf(st{0}, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("intish %s", s)
	}

	// Scaled.
	sf = reflectTypeFormatter(0, FmtDivideBy1M, reflect.TypeFor[T]())
	if s := sf(st{val}, 0); s != fmt.Sprint(scaled) {
		t.Fatalf("intish scaled %s", s)
	}
	if s := sf(st{0}, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("intish scaled %s", s)
	}
	// If scaling produces zero then the NoDefault filtering kicks in
	if s := sf(st{127}, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("intish scaled %s", s)
	}
}

func TestReflectIntish(t *testing.T) {
	testReflectIntish[int](t, 123456789, 117)
	testReflectIntish[int32](t, 123456789, 117)
	testReflectIntish[uint](t, 123456789, 117)
	testReflectIntish[uint32](t, 123456789, 117)
	testReflectIntish[uint64](t, 123456789, 117)
	testReflectIntish[int8](t, 123, 0)
	testReflectIntish[int16](t, 1234, 0)
	testReflectIntish[uint8](t, 123, 0)
	testReflectIntish[uint16](t, 1234, 0)
	// Re-test int64 to ensure the logic is the same as for the others
	testReflectIntish[int64](t, 123456789, 117)
}

// float32 and float64 have the same logic.  The rounding operations are applied before default
// testing, cf the opposite logic for ints above.

func testReflectFloatish[T float32 | float64](t *testing.T, val, ceiling T) {
	type st struct {
		V T
	}

	sf := reflectTypeFormatter(0, 0, reflect.TypeFor[T]())
	if s := sf(st{val}, 0); s != fmt.Sprint(val) {
		t.Fatalf("floatish %s", s)
	}
	if s := sf(st{0}, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("floatish %s", s)
	}

	sf = reflectTypeFormatter(0, FmtCeil, reflect.TypeFor[T]())
	if val >= 0 {
		if s := sf(st{val}, 0); s != fmt.Sprint(ceiling) {
			t.Fatalf("floatish %s", s)
		}
	} else {
		if s := sf(st{val}, PrintModNoDefaults); s != "*skip*" {
			t.Fatalf("floatish %s", s)
		}
	}
}

func TestReflectFloatish(t *testing.T) {
	testReflectFloatish[float32](t, 13.75, 14)
	testReflectFloatish[float32](t, -0.5, 0)
	testReflectFloatish[float64](t, 13.75, 14)
	testReflectFloatish[float64](t, -0.5, 0)
}

func TestReflectString(t *testing.T) {
	type st struct {
		V string
	}

	sf := reflectTypeFormatter(0, 0, reflect.TypeFor[string]())
	if s := sf(st{"hi there"}, 0); s != "hi there" {
		t.Fatalf("string %s", s)
	}
	if s := sf(st{"hi there"}, PrintModNoDefaults); s != "hi there" {
		t.Fatalf("string %s", s)
	}
	if s := sf(st{""}, PrintModNoDefaults); s != "*skip*" {
		t.Fatalf("string %s", s)
	}
}
