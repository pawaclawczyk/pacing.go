package pacing

import (
	"github.com/google/uuid"
	"math"
	"sync"
	"time"
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

type Spend struct {
	mu sync.RWMutex
	s  map[uuid.UUID]int64
}

func NewSpend() *Spend {
	return &Spend{
		mu: sync.RWMutex{},
		s:  map[uuid.UUID]int64{},
	}
}

func CurrentSlot(t time.Time) int {
	y, m, d := t.Date()
	a := time.Date(y, m, d, 0, 0, 0, 0, t.Location())
	secs := t.Sub(a).Seconds()
	return int(math.Floor(secs / 60))
}

func DistributeBudgetFragments(bidders []string) interface{} {
	return nil
}
