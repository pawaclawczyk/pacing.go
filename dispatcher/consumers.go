package dispatcher

import (
	"sync"
	"time"
)

type consumers struct {
	mu             sync.RWMutex
	consumers      map[string]time.Time
	ttl            time.Duration
	checkTTLPeriod time.Duration
	done           chan byte
}

func newConsumers() *consumers {
	return &consumers{
		mu:             sync.RWMutex{},
		consumers:      make(map[string]time.Time),
		ttl:            DefaultConsumersTTL,
		checkTTLPeriod: DefaultConsumersCheckTTLPeriod,
	}
}

func (cs *consumers) add(c string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.consumers[c] = time.Now().Add(cs.ttl)
}

func (cs *consumers) list() []string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	res := make([]string, len(cs.consumers))[:0]
	for c := range cs.consumers {
		res = append(res, c)
	}
	return res
}

func (cs *consumers) deleteOutdated(t time.Time) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	for c, ts := range cs.consumers {
		if ts.Compare(t) <= 0 {
			delete(cs.consumers, c)
		}
	}
}

func (cs *consumers) enableTTL() {
	if cs.done != nil {
		return
	}
	cs.done = make(chan byte)
	go cs.ttlCleaner()
}

func (cs *consumers) disableTTL() {
	if cs.done != nil {
		cs.done <- 1
		close(cs.done)
		cs.done = nil
	}
}

func (cs *consumers) ttlCleaner() {
	var t time.Time
	ticker := time.NewTicker(cs.checkTTLPeriod)
	for {
		select {
		case t = <-ticker.C:
			cs.deleteOutdated(t)
		case <-cs.done:
			ticker.Stop()
			return
		}
	}
}
