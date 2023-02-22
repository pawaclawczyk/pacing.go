package dispatcher

import (
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
	"time"
)

func TestWithSubject(t *testing.T) {
	type args struct {
		subject string
	}
	type want struct {
		subject string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{"empty", args{""}, want{subject: ""}},
		{"not empty", args{"test-subject"}, want{subject: "test-subject"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &announcementsOptions{}
			opt := WithSubject(tt.args.subject)
			opt(opts)
			assert.Equal(t, tt.want.subject, opts.subject)
		})
	}
}

func TestWithPeriod(t *testing.T) {
	type args struct {
		period time.Duration
	}
	type want struct {
		period time.Duration
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{"negative", args{-time.Second}, want{-time.Second}},
		{"zero", args{0}, want{0}},
		{"positive", args{time.Second}, want{time.Second}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &announcementsOptions{}
			opt := WithPeriod(tt.args.period)
			opt(opts)
			assert.Equal(t, tt.want.period, opts.period)
		})
	}
}

func TestNewAnnouncer(t *testing.T) {
	srv := RunTestServer()
	defer srv.Shutdown()
	nc := MakeTestConnection(t)
	defer nc.Close()
	broken := MakeTestConnection(t)
	broken.Close()
	assert.False(t, broken.IsConnected())

	type args struct {
		nc   *nats.Conn
		opts []AnnouncementsOption
	}
	type want struct {
		subject string
		period  time.Duration
	}
	tests := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{"defaults", args{nc, nil}, want{DefaultSubject, DefaultPeriod}, false},
		{"no connection", args{nil, nil}, want{}, true},
		{"broken connection", args{broken, nil}, want{}, true},
		{"custom subject", args{nc, []AnnouncementsOption{WithSubject("test-subject")}}, want{"test-subject", DefaultPeriod}, false},
		{"invalid subject (empty)", args{nc, []AnnouncementsOption{WithSubject("")}}, want{DefaultSubject, DefaultPeriod}, false},
		{"custom period", args{nc, []AnnouncementsOption{WithPeriod(time.Second)}}, want{DefaultSubject, time.Second}, false},
		{"invalid period (zero)", args{nc, []AnnouncementsOption{WithPeriod(0)}}, want{DefaultSubject, DefaultPeriod}, false},
		{"invalid period (negative)", args{nc, []AnnouncementsOption{WithPeriod(-time.Second)}}, want{DefaultSubject, DefaultPeriod}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAnnouncer(tt.args.nc, tt.args.opts...)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				// Test internal properties.
				assert.True(t, got.nc.IsConnected())
				assert.Equal(t, tt.want.subject, got.opts.subject)
				assert.Equal(t, tt.want.period, got.opts.period)
				assert.NotEmpty(t, got.address)
				assert.NotEmpty(t, got.msg)
				assert.Equal(t, got.address, string(got.msg))
				assert.NotNil(t, got.done)
				// Test methods.
				assert.Equal(t, got.address, got.Address())
				got.Stop()
			}
		})
	}
}

func TestIntegrationBetweenAnnouncersAndObserver(t *testing.T) {
	var alice *Observer
	var bob, charlie *Announcer
	var err error

	srv := RunTestServer()
	defer srv.Shutdown()

	// The period in observer is set twice as large as in announcer to avoid expiration bouncing.
	if alice, err = NewObserver(MakeTestConnection(t), WithPeriod(2*time.Millisecond)); err != nil {
		t.Fatalf("cannot create observer: %v", err)
	}
	defer func() {
		if err := alice.Stop(); err != nil {
			t.Fatalf("failed to stop observer: %v", err)
		}
	}()

	assert.Equal(t, []string{}, alice.Consumers())

	if bob, err = NewAnnouncer(MakeTestConnection(t), WithPeriod(time.Millisecond)); err != nil {
		t.Fatalf("cannot create announcer: %v", err)
	}
	bobRunning := true
	defer func() {
		if !bobRunning {
			return
		}
		bob.Stop()
	}()

	// Delay
	time.Sleep(time.Millisecond)

	assert.Equal(t, []string{bob.Address()}, alice.Consumers())

	if charlie, err = NewAnnouncer(MakeTestConnection(t), WithPeriod(time.Millisecond)); err != nil {
		t.Fatalf("cannot create announcer: %v", err)
	}
	charlieRunning := true
	defer func() {
		if !charlieRunning {
			return
		}
		charlie.Stop()
	}()

	// Delay
	time.Sleep(time.Millisecond)

	want := []string{bob.Address(), charlie.Address()}
	sort.Strings(want)
	got := alice.Consumers()
	sort.Strings(got)

	assert.Equal(t, want, got)

	charlie.Stop()
	charlieRunning = false

	// Delay + expiration period
	time.Sleep(3 * time.Millisecond)

	assert.Equal(t, []string{bob.Address()}, alice.Consumers())

	bob.Stop()
	bobRunning = false

	// Delay + expiration period
	time.Sleep(3 * time.Millisecond)

	assert.Equal(t, []string{}, alice.Consumers())
}
