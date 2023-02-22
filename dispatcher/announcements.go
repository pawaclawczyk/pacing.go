package dispatcher

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"time"
)

// DefaultSubject is the NATS subject used for announcements if none or invalid is provided.
const DefaultSubject = "announcements"

// DefaultPeriod is the time period between consecutive announcements.
const DefaultPeriod = 5 * time.Second

// announcementsOptions represents configurable options for Announcer and Observer.
type announcementsOptions struct {
	subject string
	period  time.Duration
}

// AnnouncementsOption allows to define configurable options.
type AnnouncementsOption func(opts *announcementsOptions)

// WithSubject configures the NATS subject used for announcements.
func WithSubject(subject string) AnnouncementsOption {
	return func(opts *announcementsOptions) {
		opts.subject = subject
	}
}

// WithPeriod configures period between announcements.
func WithPeriod(period time.Duration) AnnouncementsOption {
	return func(opts *announcementsOptions) {
		opts.period = period
	}
}

// Announcer represents the entity announcing consumer address.
type Announcer struct {
	nc      *nats.Conn
	opts    *announcementsOptions
	address string
	msg     []byte
	done    chan bool
}

// NewAnnouncer creates new Announcer instance.
func NewAnnouncer(nc *nats.Conn, opts ...AnnouncementsOption) (*Announcer, error) {
	if nc == nil {
		return nil, fmt.Errorf("NATS connection is required argument")
	}
	if !nc.IsConnected() {
		return nil, fmt.Errorf("NATS connection has invalid state: %v", nc.Status())
	}
	options := &announcementsOptions{}
	for _, opt := range opts {
		opt(options)
	}
	if options.subject == "" {
		options.subject = DefaultSubject
	}
	if options.period <= 0 {
		options.period = DefaultPeriod
	}
	address := nats.NewInbox()
	a := &Announcer{
		nc:      nc,
		opts:    options,
		address: address,
		// The communication protocol is extremely simple.
		// The announcer publishes its address as string.
		msg:  []byte(address),
		done: make(chan bool),
	}
	go a.loop()
	return a, nil
}

// Address is the address being announced.
func (a *Announcer) Address() string {
	return a.address
}

// Stop gracefully stops internal routines and cleans up resources.
func (a *Announcer) Stop() {
	a.done <- true
}

// announce implements communication protocol and encoding.
func (a *Announcer) announce() error {
	return a.nc.Publish(a.opts.subject, a.msg)
}

// loop is announcing routine skeleton.
func (a *Announcer) loop() {
	var err error
	// Run once immediately
	err = a.announce()
	if err != nil {
		log.Err(err).Msg("error occurred when publishing announcement")
	}
	ticker := time.NewTicker(a.opts.period)
	for {
		select {
		case <-ticker.C:
			err = a.announce()
			if err != nil {
				log.Err(err).Msg("error occurred when publishing announcement")
			}
		case <-a.done:
			ticker.Stop()
			return
		}
	}
}

// Observer represents the entity keeping track of announced consumers' addresses.
type Observer struct {
	nc        *nats.Conn
	opts      *announcementsOptions
	consumers *Consumers
	sub       *nats.Subscription
}

// NewObserver creates new Observer instance.
func NewObserver(nc *nats.Conn, opts ...AnnouncementsOption) (*Observer, error) {
	if nc == nil {
		return nil, fmt.Errorf("NATS connection is required argument")
	}
	if !nc.IsConnected() {
		return nil, fmt.Errorf("NATS connection has invalid state: %v", nc.Status())
	}
	options := &announcementsOptions{}
	for _, opt := range opts {
		opt(options)
	}
	if options.subject == "" {
		options.subject = DefaultSubject
	}
	if options.period <= 0 {
		options.period = DefaultPeriod
	}
	o := &Observer{
		nc:        nc,
		opts:      options,
		consumers: NewConsumers(WithTTL(options.period)),
	}
	var err error
	o.sub, err = o.nc.Subscribe(o.opts.subject, o.process)
	if err != nil {
		return nil, fmt.Errorf("cannot subscribe to announcements: %v", err)
	}
	return o, nil
}

// Consumers returns list of active consumers' addresses.
func (o *Observer) Consumers() []string {
	return o.consumers.List()
}

// Stop gracefully stops internal routines and cleans up resources.
func (o *Observer) Stop() error {
	if err := o.sub.Unsubscribe(); err != nil {
		return fmt.Errorf("cannot unsubscribe from announcements: %v", err)
	}
	return nil
}

// process implements communication protocol, encoding, and domain processing.
func (o *Observer) process(msg *nats.Msg) {
	o.consumers.Join(string(msg.Data))
}
