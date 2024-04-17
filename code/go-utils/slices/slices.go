// This may go away (in some part) when we move to Go 1.21 or 1.22

package slices

func Map[V, W any](xs []V, f func(V)W) []W {
	ys := make([]W, 0, len(xs))
	for _, v := range xs {
		ys = append(ys, f(v))
	}
	return ys
}

func Copy[V any](xs []V) []V {
	return append(make([]V, 0, len(xs)), xs...)
}
