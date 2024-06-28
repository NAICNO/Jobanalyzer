package maps

func Keys[K comparable, V any](xs map[K]V) []K {
	ys := make([]K, 0, len(xs))
	for k := range xs {
		ys = append(ys, k)
	}
	return ys
}

func MapKeys[K comparable, V, W any](xs map[K]V, f func(K) W) []W {
	ys := make([]W, 0, len(xs))
	for k := range xs {
		ys = append(ys, f(k))
	}
	return ys
}

func Values[K comparable, V any](xs map[K]V) []V {
	ys := make([]V, 0, len(xs))
	for _, v := range xs {
		ys = append(ys, v)
	}
	return ys
}

func MapValues[K comparable, V, W any](xs map[K]V, f func(V) W) []W {
	ys := make([]W, 0, len(xs))
	for _, v := range xs {
		ys = append(ys, f(v))
	}
	return ys
}
