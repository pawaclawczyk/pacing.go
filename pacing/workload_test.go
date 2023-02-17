package pacing

import (
	"github.com/google/uuid"
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
			assert.Equalf(t, tt.want, TimeToSlot(tt.args.t), "TimeToSlot(%v)", tt.args.t)
		})
	}
}

func TestMakeWorkloadSplitterReturnsValidNumberOfWorkloads(t *testing.T) {
	planned := NewPlannedSpend()
	spend := NewSpend()
	now := func() time.Time { return time.Date(2023, 2, 17, 0, 0, 0, 0, time.Local) }
	splitter := MakeWorkloadSplitter(planned, spend, now)
	tests := []struct {
		name string
		args []string
		want int
	}{
		{"zero consumers", []string{}, 0},
		{"one consumers", []string{"alice"}, 1},
		{"three consumers", []string{"alice", "bob", "charlie"}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			split := splitter(tt.args)
			assert.Len(t, split, tt.want)
		})
	}
}

func TestMakeWorkloadSplitterDividesWorkEqually(t *testing.T) {
	planned := NewPlannedSpend()
	lineItemId := uuid.New()
	planned.ps = map[uuid.UUID][]int64{
		lineItemId: {9},
	}
	spend := NewSpend()
	now := func() time.Time { return time.Date(2023, 2, 17, 0, 0, 0, 0, time.Local) }
	splitter := MakeWorkloadSplitter(planned, spend, now)
	tests := []struct {
		name string
		args []string
		want int64
	}{
		{"zero consumers", []string{}, 0},
		{"one consumers", []string{"alice"}, 9},
		{"two consumers", []string{"alice", "bob"}, 4},
		{"three consumers", []string{"alice", "bob", "charlie"}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			split := splitter(tt.args)
			for k, v := range split {
				assert.Contains(t, tt.args, k)
				assert.Equal(t, v.(map[uuid.UUID]int64)[lineItemId], tt.want)
			}
		})
	}
}

func TestMakeWorkloadSplitterTakesIntoAccountSpend(t *testing.T) {
	planned := NewPlannedSpend()
	lineItemId := uuid.New()
	planned.ps = map[uuid.UUID][]int64{
		lineItemId: {9},
	}
	spend := NewSpend()
	now := func() time.Time { return time.Date(2023, 2, 17, 0, 0, 0, 0, time.Local) }
	consumers := []string{"alice", "bob", "charlie"}
	splitter := MakeWorkloadSplitter(planned, spend, now)
	tests := []struct {
		name string
		args int64
		want int64
	}{
		{"no spend", 0, 3},
		{"partial spend", 6, 1},
		{"partial spend not enough for all", 7, 0},
		{"full spend", 9, 0},
		{"overspend", 10, 0},
	}
	for _, tt := range tests {
		spend.s[lineItemId] = tt.args
		t.Run(tt.name, func(t *testing.T) {
			split := splitter(consumers)
			assert.Len(t, split, 3)
			for _, v := range split {
				vv, ok := v.(map[uuid.UUID]int64)[lineItemId]
				// Expect that items without budget are not distributed
				if tt.want > 0 {
					assert.True(t, ok)
					assert.Equal(t, vv, tt.want)
				} else {
					assert.False(t, ok)
				}
			}
		})
	}
}

func TestMakeWorkloadSplitterComputesForSpecificSlot(t *testing.T) {
	planned := NewPlannedSpend()
	lineItemId := uuid.New()
	planned.ps = map[uuid.UUID][]int64{
		lineItemId: {9, 18, 27},
	}
	spend := NewSpend()
	consumers := []string{"alice", "bob", "charlie"}
	tests := []struct {
		name string
		args int
		want int64
	}{
		{"time slot 0", 0, 3},
		{"time slot 1", 1, 6},
		{"time slot 2", 2, 9},
	}
	for _, tt := range tests {
		now := func() time.Time { return time.Date(2023, 2, 17, 0, tt.args, 0, 0, time.Local) }
		splitter := MakeWorkloadSplitter(planned, spend, now)
		t.Run(tt.name, func(t *testing.T) {
			split := splitter(consumers)
			assert.Len(t, split, 3)
			for _, v := range split {
				vv, ok := v.(map[uuid.UUID]int64)[lineItemId]
				assert.True(t, ok)
				assert.Equal(t, vv, tt.want)
			}
		})
	}
}
