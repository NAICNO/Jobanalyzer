// This may go away (in some part) when we move to Go 1.21 or 1.22

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

func Copy[S ~[]E, E any](xs S) S {
	return append(make(S, 0, len(xs)), xs...)
}

func Insert[S ~[]E, E any](s S, i int, v ...E) S {
	n := len(v)
	if cap(s) - len(s) >= n {
		// update in-place
		t := s[:len(s)+n]
		copy(t[i+n:], s[i:])
		copy(t[i:], v)
		return t
	}

	// make a new one
	t := make(S, len(s) + n)
	copy(t, s[:i])
	copy(t[i:], v)
	copy(t[i+len(v):], s[i:])
	return t
}

func BinarySearchFunc[S ~[]E, E, T any](xs S, target T, cmp func(E, T) int) (int, bool) {
	lo := 0
	lim := len(xs)
	for lo < lim {
		mid := lo + (lim - lo) / 2
		rel := cmp(xs[mid], target)
		switch {
		case rel == 0:
			return mid, true
		case rel < 0:
			lo = mid+1
		default:
			lim = mid
		}
	}
	return lo, false
}
