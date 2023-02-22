package dispatcher

import (
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

// nop is the do nothing workload callback
func nop(_ string) {}

type workloads struct {
	mu sync.Mutex
	ws []string
}

func (ws *workloads) get(i int) string {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.ws[i]
}

// collect creates a callback and a pointer where the workload is stored once callback is invoked.
func collect(ws *workloads) func(w string) {
	return func(w string) {
		ws.mu.Lock()
		defer ws.mu.Unlock()
		ws.ws = append(ws.ws, w)
	}
}

func TestWorkloadInterceptor(t *testing.T) {
	ws := new(workloads)
	cb := collect(ws)
	assert.Empty(t, ws.ws)
	cb("first workload")
	assert.Equal(t, "first workload", ws.get(0))
	cb("second workload")
	cb("third workload")
	assert.Equal(t, "second workload", ws.get(1))
	assert.Equal(t, "third workload", ws.get(2))
}

func TestNewReceiver(t *testing.T) {
	srv := RunTestServer()
	defer srv.Shutdown()

	receiver, err := NewReceiver(nop)
	assert.NotNil(t, receiver)
	assert.Nil(t, err)
	defer func() {
		err := receiver.Shutdown()
		assert.Nil(t, err)
	}()

	assert.Nil(t, receiver.conn)
}

func TestRunReceiver(t *testing.T) {
	srv := RunTestServer()
	defer srv.Shutdown()

	receiver, _ := NewReceiver(nop)
	defer func() {
		err := receiver.Shutdown()
		assert.Nil(t, err)
	}()

	err := receiver.Run()
	assert.Nil(t, err)

	assert.True(t, receiver.conn.IsConnected())
	assert.NotEmpty(t, receiver.inbox)
}

func TestAnnouncer(t *testing.T) {
	srv := RunTestServer()
	defer srv.Shutdown()

	receiver, _ := NewReceiver(nop)
	defer func() {
		err := receiver.Shutdown()
		assert.Nil(t, err)
	}()

	_ = receiver.Run()

	nc, err := nats.Connect(nats.DefaultURL)
	assert.Nil(t, err)

	sub, err := nc.SubscribeSync(receiver.announcements)
	assert.Nil(t, err)

	msg, err := sub.NextMsg(2 * receiver.announcementsPeriod)
	assert.Nil(t, err)

	assert.Equal(t, receiver.inbox, string(msg.Data))
}

func TestWorkload(t *testing.T) {
	srv := RunTestServer()
	defer srv.Shutdown()

	ws := new(workloads)
	cb := collect(ws)
	receiver, _ := NewReceiver(cb)
	defer func() {
		err := receiver.Shutdown()
		assert.Nil(t, err)
	}()

	_ = receiver.Run()

	nc, err := nats.Connect(nats.DefaultURL)
	assert.Nil(t, err)

	err = nc.Publish(receiver.inbox, []byte("first workload"))
	assert.Nil(t, err)

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, "first workload", ws.get(0))

	err = nc.Publish(receiver.inbox, []byte("second workload"))
	assert.Nil(t, err)
	err = nc.Publish(receiver.inbox, []byte("third workload"))
	assert.Nil(t, err)

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, "second workload", ws.get(1))
	assert.Equal(t, "third workload", ws.get(2))
}

func TestDoubleShutdown(t *testing.T) {
	srv := RunTestServer()
	defer srv.Shutdown()

	receiver, err := NewReceiver(nop)
	assert.Nil(t, err)
	err = receiver.Run()
	assert.Nil(t, err)

	err = receiver.Shutdown()
	assert.Nil(t, err)
	err = receiver.Shutdown()
	assert.Nil(t, err)
}
