package dispatcher

import (
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
	"time"
)

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
