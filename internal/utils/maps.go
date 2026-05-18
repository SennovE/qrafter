package utils

func UnionSets[T comparable](ms ...map[T]struct{}) map[T]struct{} {
	if len(ms) == 0 {
		return make(map[T]struct{}, 0)
	}
	l := len(ms[0])
	for _, m := range ms {
		l = maxInt(l, len(m))
	}
	res := make(map[T]struct{}, l)
	for _, m := range ms {
		for k, v := range m {
			res[k] = v
		}
	}
	return res
}

func maxInt(elems ...int) int {
	m := elems[0]
	for _, e := range elems {
		if m < e {
			m = e
		}
	}
	return m
}
