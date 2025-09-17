package math

func ClampUintMin(n, min uint) uint {
	if n < min {
		return min
	} else {
		return n
	}
}
func ClampUintMax(n, max uint) uint {
	if n > max {
		return max
	} else {
		return n
	}
}
func ClampUintMinMax(n, min, max uint) uint {
	return ClampUintMax(ClampUintMin(n, min), max)
}
