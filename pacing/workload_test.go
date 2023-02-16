package pacing

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPlannedSpendLoad(t *testing.T) {
	path, recs, err := CreateSnapshot(3)
	assert.NoError(t, err)

	s := NewPlannedSpend()
	err = s.Load(path)
	assert.NoError(t, err)

	for _, rec := range recs {
		v, ok := s.ps[rec.LineItemID]
		assert.True(t, ok)
		assert.Equal(t, rec.DailyBudget, Sum(v))
	}
}

func TestCurrentSlot(t *testing.T) {
	type args struct {
		t time.Time
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"zero", args{time.Date(2023, 2, 16, 0, 0, 0, 0, time.Local)}, 0},
		{"ending", args{time.Date(2023, 2, 16, 0, 0, 59, 0, time.Local)}, 0},
		{"beginning", args{time.Date(2023, 2, 16, 0, 1, 0, 0, time.Local)}, 1},
		{"last", args{time.Date(2023, 2, 16, 23, 59, 59, 0, time.Local)}, 1439},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, CurrentSlot(tt.args.t), "CurrentSlot(%v)", tt.args.t)
		})
	}
}
