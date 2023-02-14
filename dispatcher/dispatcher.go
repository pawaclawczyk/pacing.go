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
const DefaultConsumersTTLCheck = 30 * time.Second
const DefaultPeriod = 1 * time.Minute

type Dispatcher struct {
	url                  string
	announcements        string
	conn                 *nats.Conn
	period               time.Duration
	done                 chan bool
	consumers            map[string]time.Time
	consumersTTL         time.Duration
	consumersTTLCheck    time.Duration
	consumersCleanerDone chan bool
	consumersMutex       sync.Mutex
	watcherSub           *nats.Subscription
}

func NewDispatcher() (*Dispatcher, error) {
	return &Dispatcher{
		url:                  nats.DefaultURL,
		announcements:        DefaultAnnouncements,
		period:               DefaultPeriod,
		done:                 make(chan bool),
		consumers:            make(map[string]time.Time),
		consumersTTL:         DefaultConsumersTTL,
		consumersTTLCheck:    DefaultConsumersTTLCheck,
		consumersCleanerDone: make(chan bool),
		consumersMutex:       sync.Mutex{},
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
	go d.DeleteUnseenConsumers()
	go d.DispatchWorkload()
	return nil
}

func (d *Dispatcher) watchAnnouncements(msg *nats.Msg) {
	consumer := string(msg.Data)
	d.consumersMutex.Lock()
	defer d.consumersMutex.Unlock()
	d.consumers[consumer] = time.Now()
}

func (d *Dispatcher) DeleteUnseenConsumers() {
	tick := time.Tick(d.consumersTTLCheck)
	for {
		select {
		case <-tick:
			d.consumersMutex.Lock()
			for k, v := range d.consumers {
				if v.Before(time.Now().Add(-d.consumersTTL)) {
					delete(d.consumers, k)
				}
			}
			d.consumersMutex.Unlock()
		case <-d.consumersCleanerDone:
			return
		}
	}
}

func (d *Dispatcher) Consumers() []string {
	var watchers []string
	d.consumersMutex.Lock()
	defer d.consumersMutex.Unlock()
	for k := range d.consumers {
		watchers = append(watchers, k)
	}
	return watchers
}

func (d *Dispatcher) DispatchWorkload() {
	var err error
	var consumers []string
	tick := time.Tick(d.period)
	for {
		select {
		case <-tick:
			consumers = d.Consumers()
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
	d.consumersCleanerDone <- true
	d.done <- true
	d.conn.Close()
}
