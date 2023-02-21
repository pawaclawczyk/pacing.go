package dispatcher

import (
	"sync"
	"time"
)

// DefaultTTL is the default TTL value used in case when none or invalid provided.
const DefaultTTL = time.Second

// consumerOptions contains configurable options.
type consumerOptions struct {
	ttl time.Duration
}

// ConsumerOption type allows defining configurable options.
type ConsumerOption func(opts *consumerOptions)

// WithTTL allows configuring TTL value.
func WithTTL(ttl time.Duration) ConsumerOption {
	return func(opts *consumerOptions) {
		opts.ttl = ttl
	}
}

// Consumers represents a set of expiring cons.
type Consumers struct {
	opts   *consumerOptions
	mu     sync.Mutex
	now    func() time.Time
	ttlSet map[string]time.Time
}

// NewConsumers creates an empty Consumers instance.
// Configurable options are:
// - WithTTL.
func NewConsumers(opts ...ConsumerOption) *Consumers {
	configured := &consumerOptions{}
	for _, opt := range opts {
		opt(configured)
	}

	if configured.ttl <= 0 {
		configured.ttl = DefaultTTL
	}

	return &Consumers{
		opts:   configured,
		mu:     sync.Mutex{},
		now:    time.Now,
		ttlSet: make(map[string]time.Time, 10),
	}
}

// Join adds given consumer to the set.
func (cs *Consumers) Join(consumer string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.ttlSet[consumer] = cs.now().Add(cs.opts.ttl)
}

// Leave removes given consumer from the set.
func (cs *Consumers) Leave(consumer string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	delete(cs.ttlSet, consumer)
}

// List returns not expired cons.
func (cs *Consumers) List() []string {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	now := cs.now()
	consumers := make([]string, 0, len(cs.ttlSet))
	for con, ttl := range cs.ttlSet {
		if ttl.After(now) {
			consumers = append(consumers, con)
		} else {
			delete(cs.ttlSet, con)
		}
	}
	return consumers
}
