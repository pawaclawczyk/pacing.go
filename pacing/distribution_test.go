package pacing

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEvenDistribution(t *testing.T) {
	type args struct {
		val int64
	}
	type want struct {
		low  int64
		high int64
		sum  int64
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{"divisible without reminder", args{17 * TimeSlots}, want{17, 17, 17 * TimeSlots}},
		{"divisible with reminder", args{17*TimeSlots + 123}, want{17, 18, 17*TimeSlots + 123}},
		{"zero", args{0}, want{0, 0, 0}},
		{"negative", args{-1}, want{0, 0, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := EvenDistribution(tt.args.val)
			for _, v := range dist {
				assert.GreaterOrEqual(t, v, tt.want.low, "EvenDistribution(%v) >=%v", tt.args.val, tt.want.low)
				assert.LessOrEqual(t, v, tt.want.high, "EvenDistribution(%v) <=%v", tt.args.val, tt.want.high)
			}
			assert.Equal(t, tt.want.sum, Sum(dist), "EvenDistribution(%v) sum", tt.want.sum)
		})
	}
}
