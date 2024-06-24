package slices

// This is a little tricky.  If V == W then we'd ideally want the return type to be S, not []W.  But
// we can't express that.  And it's probably much more important that f's result can have a
// different type than its input.

func Map[S ~[]V, V, W any](xs S, f func(V)W) []W {
	ys := make([]W, 0, len(xs))
	for _, v := range xs {
		ys = append(ys, f(v))
	}
	return ys
}

func Catenate[S ~[]E, E any](xss []S) []E {
	n := 0
	for _, xs := range xss {
		n += len(xs)
	}
	result := make([]E, 0, n)
	for _, xs := range xss {
		result = append(result, xs...)
	}
	return result
}

// This is gross but convenient since maps.Values(InputStreamSet) is []*SampleStream
func CatenateP[S ~[]E, E any](xss []*S) []E {
	n := 0
	for _, xs := range xss {
		n += len(*xs)
	}
	result := make([]E, 0, n)
	for _, xs := range xss {
		result = append(result, *xs...)
	}
	return result
}
