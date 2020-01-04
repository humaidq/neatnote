package common

// ContainsInt64 returns whether an array if it contains a specific elements.
func ContainsInt64(a []int64, i int64) bool {
	for _, v := range a {
		if v == i {
			return true
		}
	}
	return false
}

func RemoveInt64(a []int64, v int64) []int64 {
	for i, k := range a {
		if k == v {
			return append(a[:i], a[i+1:]...)
		}
	}
	return a
}
