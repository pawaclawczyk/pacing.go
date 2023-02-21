package dispatcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"pacing.go/shared"
	"time"
)

type WorkloadCallback func(consumers []string) map[string]interface{}

type Dispatcher struct {
	url           string
	announcements string
	cons          *Consumers
	period        time.Duration
	wcb           WorkloadCallback

	conn *nats.Conn
	sub  *nats.Subscription
	done chan byte
}

func NewDispatcher(wcb WorkloadCallback) (*Dispatcher, error) {
	return &Dispatcher{
		url:           nats.DefaultURL,
		announcements: DefaultAnnouncements,
		cons:          NewConsumers(),
		period:        DefaultDispatcherPeriod,
		wcb:           wcb,
	}, nil
}

func (d *Dispatcher) Run() error {
	var err error
	// Connect to NATs server
	d.conn, err = nats.Connect(d.url)
	if err != nil {
		return err
	}
	if !d.conn.IsConnected() {
		return errors.New(fmt.Sprintf("Cannot connect, connection status is %s\n", d.conn.Status()))
	}
	// Subscribe to announcements
	d.sub, err = d.conn.Subscribe(d.announcements, d.watchAnnouncements)
	if err != nil {
		return err
	}
	// Run dispatcher routine
	d.done = make(chan byte)
	go d.dispatcher()
	// Done
	return nil
}

func (d *Dispatcher) watchAnnouncements(msg *nats.Msg) {
	d.cons.Join(string(msg.Data))
}

func (d *Dispatcher) dispatcher() {
	var err error
	var enc []byte
	ticker := time.NewTicker(d.period)
	for {
		select {
		case <-ticker.C:
			for c, w := range d.wcb(d.cons.List()) {
				log.Debug().Msg(fmt.Sprintf("(dispatcher) sending workload to consumer: %v", c))
				enc, err = json.Marshal(w)
				shared.PanicIf(err)
				err = d.conn.Publish(c, enc)
				shared.PanicIf(err)
			}
		case <-d.done:
			ticker.Stop()
			return
		}
	}
}

func (d *Dispatcher) Shutdown() {
	// Shutdown dispatcher routine
	if d.done != nil {
		d.done <- 1
		close(d.done)
		d.done = nil
	}
	// Unsubscribe from announcements
	if d.sub != nil {
		// The subscription is either not valid or connection is broken
		_ = d.sub.Unsubscribe()
		d.sub = nil
	}
	// Disconnect fromNATS server
	d.conn.Close()
}
