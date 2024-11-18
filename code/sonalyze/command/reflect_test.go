package command

import (
	"reflect"
	"testing"

	uslices "go-utils/slices"
)

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
