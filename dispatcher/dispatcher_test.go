package dispatcher

import (
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const Delay = time.Millisecond

func TestNewDispatcher(t *testing.T) {
	srv := RunServer()
	defer srv.Shutdown()

	dispatcher, err := NewDispatcher()
	assert.Nil(t, err)
	assert.NotNil(t, dispatcher)

	err = dispatcher.Run()
	assert.Nil(t, err)

	dispatcher.Shutdown()
}

func TestWatchAnnouncements(t *testing.T) {
	srv := RunServer()
	defer srv.Shutdown()

	dispatcher, err := NewDispatcher()
	assert.Nil(t, err)

	err = dispatcher.Run()
	assert.Nil(t, err)
	defer dispatcher.Shutdown()

	nc, err := nats.Connect(nats.DefaultURL)
	assert.Nil(t, err)

	err = nc.Publish(DefaultAnnouncements, []byte("consumer-1"))
	assert.Nil(t, err)

	time.Sleep(Delay)

	consumers := dispatcher.consumers.list()
	assert.Len(t, consumers, 1)
	assert.Contains(t, consumers, "consumer-1")

	err = nc.Publish(DefaultAnnouncements, []byte("consumer-2"))
	assert.Nil(t, err)
	err = nc.Publish(DefaultAnnouncements, []byte("consumer-3"))
	assert.Nil(t, err)

	time.Sleep(Delay)

	consumers = dispatcher.consumers.list()
	assert.Len(t, consumers, 3)
	assert.Contains(t, consumers, "consumer-1")
	assert.Contains(t, consumers, "consumer-2")
	assert.Contains(t, consumers, "consumer-3")
}

func TestDispatchWorkload(t *testing.T) {
	srv := RunServer()
	defer srv.Shutdown()

	dispatcher, _ := NewDispatcher()
	dispatcher.period = 100 * time.Millisecond
	_ = dispatcher.Run()
	defer dispatcher.Shutdown()

	nc, err := nats.Connect(nats.DefaultURL)
	assert.Nil(t, err)

	c1, err := nc.SubscribeSync("consumer-1")
	assert.Nil(t, err)
	c2, err := nc.SubscribeSync("consumer-2")
	assert.Nil(t, err)

	err = nc.Publish(DefaultAnnouncements, []byte("consumer-1"))
	assert.Nil(t, err)
	err = nc.Publish(DefaultAnnouncements, []byte("consumer-2"))
	assert.Nil(t, err)

	time.Sleep(dispatcher.period)

	msg1, err := c1.NextMsg(time.Millisecond)
	assert.Nil(t, err)
	err = c1.Unsubscribe()
	assert.Nil(t, err)
	msg2, err := c2.NextMsg(time.Millisecond)
	assert.Nil(t, err)
	err = c2.Unsubscribe()
	assert.Nil(t, err)

	assert.Equal(t, "consumer-1", string(msg1.Data))
	assert.Equal(t, "consumer-2", string(msg2.Data))
}
