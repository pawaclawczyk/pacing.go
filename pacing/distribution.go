package pacing

const TimeSlots = 1440

// EvenDistribution creates even or nearly even integer distribution for TimeSlots slots of given integer value.
// If the value is not divisible by TimeSlots then ending slots have values smaller by 1.
// The values in distribution sum up to the given value.
// If the given value is negative EvenDistribution returns distribution with zeros.
func EvenDistribution(val int64) []int64 {
	dist := make([]int64, TimeSlots)
	// Fallback for negative or zero input value.
	if val <= 0 {
		return dist
	}
	slotVal := val / TimeSlots
	for i := range dist {
		dist[i] = slotVal
	}
	// Integer division ignores the remainder,
	// so in the case of non-zero reminder the sum of distribution is less than the input value.
	// The difference is spread over distribution by adding one to the first slots.
	distSum := Sum(dist)
	i := 0
	for distSum < val {
		dist[i]++
		distSum++
		i++
	}
	return dist
}
