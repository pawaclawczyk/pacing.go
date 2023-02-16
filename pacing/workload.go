package pacing

import (
	"github.com/google/uuid"
	"sync"
)

type PlannedSpend struct {
	mu sync.RWMutex
	ps map[uuid.UUID][]int64
}

func NewPlannedSpend() *PlannedSpend {
	return &PlannedSpend{
		mu: sync.RWMutex{},
		ps: map[uuid.UUID][]int64{},
	}
}

func (s *PlannedSpend) Load(path string) error {
	snapshot, err := LoadSnapshot(path)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, rec := range snapshot {
		s.ps[rec.LineItemID] = EvenDistribution(rec.DailyBudget)
	}
	return nil
}

func DistributeBudgetFragments(bidders []string) interface{} {
	return nil
}
