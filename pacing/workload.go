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

func (s *PlannedSpend) Get(t int) map[uuid.UUID]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make(map[uuid.UUID]int64)
	for id, dist := range s.ps {
		res[id] = dist[t]
	}
	return res
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

func (s *Spend) Get(id uuid.UUID) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res, ok := s.s[id]
	if !ok {
		return int64(0)
	}
	return res
}

func TimeToSlot(t time.Time) int {
	y, m, d := t.Date()
	a := time.Date(y, m, d, 0, 0, 0, 0, t.Location())
	secs := t.Sub(a).Seconds()
	return int(math.Floor(secs / 60))
}

func MakeWorkloadSplitter(planned *PlannedSpend, spend *Spend, now func() time.Time) func(consumers []string) map[string]interface{} {
	return func(consumers []string) map[string]interface{} {
		if len(consumers) == 0 {
			return map[string]interface{}{}
		}
		slot := TimeToSlot(now())
		consWrk := map[uuid.UUID]int64{}
		for id, planned := range planned.Get(slot) {
			diff := planned - spend.Get(id)
			// skip line item if the available budget will be 0 or less per consumer
			if diff < int64(len(consumers)) {
				continue
			}
			fragment := diff / int64(len(consumers))
			consWrk[id] = fragment
		}
		wrk := map[string]interface{}{}
		for _, c := range consumers {
			wrk[c] = consWrk
		}
		return wrk
	}
}
