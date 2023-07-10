package utils

func Contains[T string | int](elems []T, item T) bool {
	for _, v := range elems {
		if v == item {
			return true
		}
	}
	return false
}
