package pacing

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSum(t *testing.T) {
	type args struct {
		a []int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{"empty", args{[]int64{}}, 0},
		{"positive", args{[]int64{1, 2, 3, 4, 5, 6, 7}}, 28},
		{"negative", args{[]int64{-1, -2, -3, -4, -5, -6, -7}}, -28},
		{"positive_and_negative", args{[]int64{1, -2, 3, -4, 5, -6, 7}}, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, Sum(tt.args.a), "Sum(%v)", tt.args.a)
		})
	}
}
