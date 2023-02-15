package dispatcher

import (
	"errors"
	"fmt"
	"github.com/nats-io/nats.go"
	"pacing.go/shared"
	"sync"
	"time"
)

const DefaultConsumersTTL = 30 * time.Second
const DefaultConsumersCheckTTLPeriod = DefaultConsumersTTL / 2
const DefaultPeriod = 1 * time.Minute

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
	var res []string
	for c := range cs.consumers {
		res = append(res, c)
	}
	return res
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
			cs.mu.Lock()
			for c, ts := range cs.consumers {
				if ts.Before(t) {
					delete(cs.consumers, c)
				}
			}
			cs.mu.Unlock()
		case <-cs.done:
			ticker.Stop()
			return
		}
	}
}

type Dispatcher struct {
	url           string
	announcements string
	conn          *nats.Conn
	period        time.Duration
	done          chan bool
	consumers     *consumers
	watcherSub    *nats.Subscription
}

func NewDispatcher() (*Dispatcher, error) {
	return &Dispatcher{
		url:           nats.DefaultURL,
		announcements: DefaultAnnouncements,
		period:        DefaultPeriod,
		done:          make(chan bool),
		consumers:     newConsumers(),
	}, nil
}

func (d *Dispatcher) Run() error {
	var err error
	d.conn, err = nats.Connect(d.url)
	if err != nil {
		return err
	}
	if !d.conn.IsConnected() {
		return errors.New(fmt.Sprintf("Cannot connect, connection status is %s\n", d.conn.Status()))
	}
	d.watcherSub, err = d.conn.Subscribe(d.announcements, d.watchAnnouncements)
	if err != nil {
		return err
	}
	d.consumers.enableTTL()
	go d.DispatchWorkload()
	return nil
}

func (d *Dispatcher) watchAnnouncements(msg *nats.Msg) {
	consumer := string(msg.Data)
	d.consumers.add(consumer)
}

func (d *Dispatcher) DispatchWorkload() {
	var err error
	var consumers []string
	tick := time.Tick(d.period)
	for {
		select {
		case <-tick:
			consumers = d.consumers.list()
			for _, subject := range consumers {
				err = d.conn.Publish(subject, []byte(subject))
				shared.PanicIf(err)
			}
		case <-d.done:
			return
		default:
		}
	}
}

func (d *Dispatcher) Shutdown() {
	d.consumers.disableTTL()
	d.done <- true
	d.conn.Close()
}
