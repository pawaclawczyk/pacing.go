package dispatcher

import (
	"fmt"
	"github.com/nats-io/nats.go"
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
}

// NewAnnouncer creates new Announcer instance.
func NewAnnouncer(nc *nats.Conn, opts ...AnnouncementsOption) (*Announcer, error) {
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
	return &Announcer{
		nc:      nc,
		opts:    options,
		address: nats.NewInbox(),
	}, nil
}

// Observer represents the entity keeping track of announced consumers' addresses.
type Observer struct {
	nc        *nats.Conn
	opts      *announcementsOptions
	consumers *Consumers
}

// NewObserver creates new Observer instance.
func NewObserver(nc *nats.Conn, opts ...AnnouncementsOption) (*Observer, error) {
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
	return &Observer{
		nc:        nc,
		opts:      options,
		consumers: NewConsumers(WithTTL(options.period)),
	}, nil
}
