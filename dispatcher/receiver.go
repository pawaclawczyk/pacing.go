package dispatcher

import (
	"errors"
	"fmt"
	"github.com/nats-io/nats.go"
	"pacing.go/shared"
	"time"
)

type ConsumeCallback func(w string)

type Receiver struct {
	url                 string
	announcements       string
	announcementsPeriod time.Duration
	ccb                 ConsumeCallback

	conn  *nats.Conn
	inbox string
	done  chan byte
	sub   *nats.Subscription
}

func NewReceiver(ccb func(workload string)) (*Receiver, error) {
	return &Receiver{
		url:                 nats.DefaultURL,
		announcements:       DefaultAnnouncements,
		announcementsPeriod: DefaultAnnouncementPeriod,
		ccb:                 ccb,
	}, nil
}

func (r *Receiver) Run() error {
	var err error
	// Bootstrap resources
	r.conn, err = nats.Connect(r.url)
	if err != nil {
		return err
	}
	if !r.conn.IsConnected() {
		return errors.New(fmt.Sprintf("Cannot connect, connection status is %s\n", r.conn.Status()))
	}
	r.inbox = nats.NewInbox()
	// Subscribe for workloads
	r.sub, err = r.conn.Subscribe(r.inbox, func(msg *nats.Msg) {
		r.ccb(string(msg.Data))
	})
	// Run processes
	r.done = make(chan byte)
	go r.Announcer(r.done)
	return nil
}

func (r *Receiver) Announcer(done <-chan byte) {
	var err error
	tick := time.Tick(r.announcementsPeriod)
	for {
		select {
		case <-tick:
			err = r.conn.Publish(r.announcements, []byte(r.inbox))
			shared.PanicIf(err)
		case <-done:
			return
		default:
		}
	}
}

func (r *Receiver) Shutdown() error {
	var err error
	// Shutdown processes
	if r.done != nil {
		r.done <- 1
		close(r.done)
		r.done = nil
	}
	// Free resources
	if r.sub != nil && r.sub.IsValid() {
		err = r.sub.Unsubscribe()
	}
	if r.conn != nil {
		r.conn.Close()
	}
	return err
}
