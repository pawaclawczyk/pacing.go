package dispatcher

import (
	"sync"
	"time"
)

type consumersConcurrentExpiration struct {
	mu             sync.RWMutex
	consumers      map[string]time.Time
	ttl            time.Duration
	checkTTLPeriod time.Duration
	done           chan byte
}

func newConsumersConcurrentExpiration() *consumersConcurrentExpiration {
	return &consumersConcurrentExpiration{
		mu:             sync.RWMutex{},
		consumers:      make(map[string]time.Time),
		ttl:            DefaultConsumersTTL,
		checkTTLPeriod: DefaultConsumersCheckTTLPeriod,
	}
}

func (cs *consumersConcurrentExpiration) add(c string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.consumers[c] = time.Now().Add(cs.ttl)
}

func (cs *consumersConcurrentExpiration) list() []string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	res := make([]string, 0, len(cs.consumers))
	for c := range cs.consumers {
		res = append(res, c)
	}
	return res
}

func (cs *consumersConcurrentExpiration) deleteExpired(t time.Time) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	for c, ts := range cs.consumers {
		if ts.Compare(t) <= 0 {
			delete(cs.consumers, c)
		}
	}
}

func (cs *consumersConcurrentExpiration) enableTTL() {
	if cs.done != nil {
		return
	}
	cs.done = make(chan byte)
	go cs.ttlCleaner()
}

func (cs *consumersConcurrentExpiration) disableTTL() {
	if cs.done != nil {
		cs.done <- 1
		close(cs.done)
		cs.done = nil
	}
}

func (cs *consumersConcurrentExpiration) ttlCleaner() {
	var t time.Time
	ticker := time.NewTicker(cs.checkTTLPeriod)
	for {
		select {
		case t = <-ticker.C:
			cs.deleteExpired(t)
		case <-cs.done:
			ticker.Stop()
			return
		}
	}
}

type consumersAmortizedExpiration struct {
	mu        sync.RWMutex
	consumers map[string]time.Time
	ttl       time.Duration
	now       func() time.Time
}

func newConsumersAmortizedExpiration() *consumersAmortizedExpiration {
	return &consumersAmortizedExpiration{
		mu:        sync.RWMutex{},
		consumers: make(map[string]time.Time),
		ttl:       DefaultConsumersTTL,
		now:       time.Now,
	}
}

func (cs *consumersAmortizedExpiration) add(c string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.consumers[c] = time.Now().Add(cs.ttl)
}

func (cs *consumersAmortizedExpiration) list() []string {
	cs.mu.RLock()
	n := len(cs.consumers)
	expired := 0
	now := cs.now()
	res := make([]string, 0, n)
	for c, exp := range cs.consumers {
		if exp.After(now) {
			res = append(res, c)
		} else {
			expired++
		}
	}
	cs.mu.RUnlock()
	if expired*2 > n {
		cs.deleteExpired(now)
	}
	return res
}

func (cs *consumersAmortizedExpiration) deleteExpired(t time.Time) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	for c, ts := range cs.consumers {
		if ts.Compare(t) <= 0 {
			delete(cs.consumers, c)
		}
	}
}
