package utils

func UnionSets[T comparable](ms ...map[T]struct{}) map[T]struct{} {
	l := len(ms[0])
	for _, m := range ms {
		l = max(l, len(m))
	}
	res := make(map[T]struct{}, l)
	for _, m := range ms {
		for k, v := range m {
			res[k] = v
		}
	}
	return res
}

func max(elems ...int) int {
	m := elems[0]
	for _, e := range elems {
		if m < e {
			m = e
		}
	}
	return m
}
