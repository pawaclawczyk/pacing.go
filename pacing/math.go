package pacing

func Sum(a []int64) int64 {
	var s int64
	for _, v := range a {
		s += v
	}
	return s
}
