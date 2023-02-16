package pacing

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
