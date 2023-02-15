package dispatcher

import (
	"errors"
	"fmt"
	"github.com/nats-io/nats.go"
	"pacing.go/shared"
	"time"
)

type Dispatcher struct {
	url           string
	announcements string
	consumers     *consumers
	period        time.Duration

	conn *nats.Conn
	sub  *nats.Subscription
	done chan byte
}

func NewDispatcher() (*Dispatcher, error) {
	return &Dispatcher{
		url:           nats.DefaultURL,
		announcements: DefaultAnnouncements,
		consumers:     newConsumers(),
		period:        DefaultDispatcherPeriod,
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
	// Enable TTL on consumers
	d.consumers.enableTTL()
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
	d.consumers.add(string(msg.Data))
}

func (d *Dispatcher) dispatcher() {
	var err error
	ticker := time.NewTicker(d.period)
	for {
		select {
		case <-ticker.C:
			for _, subject := range d.consumers.list() {
				err = d.conn.Publish(subject, []byte(subject))
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
	// Shutdown consumers routine
	d.consumers.disableTTL()
	// Disconnect fromNATS server
	d.conn.Close()
}
