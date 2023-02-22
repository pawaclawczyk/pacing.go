package dispatcher

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const Delay = time.Millisecond

func ConsumerNameCallback(consumers []string) map[string]interface{} {
	res := make(map[string]interface{}, len(consumers))
	for _, c := range consumers {
		res[c] = c
	}
	return res
}

func TestNewDispatcher(t *testing.T) {
	srv := RunTestServer()
	defer srv.Shutdown()

	dispatcher, err := NewDispatcher(ConsumerNameCallback)
	assert.Nil(t, err)
	assert.NotNil(t, dispatcher)

	err = dispatcher.Run()
	assert.Nil(t, err)

	dispatcher.Shutdown()
}

func TestWatchAnnouncements(t *testing.T) {
	srv := RunTestServer()
	defer srv.Shutdown()

	dispatcher, err := NewDispatcher(ConsumerNameCallback)
	assert.Nil(t, err)

	err = dispatcher.Run()
	assert.Nil(t, err)
	defer dispatcher.Shutdown()

	nc, err := nats.Connect(nats.DefaultURL)
	assert.Nil(t, err)

	err = nc.Publish(DefaultAnnouncements, []byte("consumer-1"))
	assert.Nil(t, err)

	time.Sleep(Delay)

	consumers := dispatcher.cons.List()
	assert.Len(t, consumers, 1)
	assert.Contains(t, consumers, "consumer-1")

	err = nc.Publish(DefaultAnnouncements, []byte("consumer-2"))
	assert.Nil(t, err)
	err = nc.Publish(DefaultAnnouncements, []byte("consumer-3"))
	assert.Nil(t, err)

	time.Sleep(Delay)

	consumers = dispatcher.cons.List()
	assert.Len(t, consumers, 3)
	assert.Contains(t, consumers, "consumer-1")
	assert.Contains(t, consumers, "consumer-2")
	assert.Contains(t, consumers, "consumer-3")
}

func TestDispatchWorkload(t *testing.T) {
	srv := RunTestServer()
	defer srv.Shutdown()

	dispatcher, _ := NewDispatcher(ConsumerNameCallback)
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

	var workload string
	err = json.Unmarshal(msg1.Data, &workload)
	assert.Nil(t, err)
	assert.Equal(t, "consumer-1", workload)
	err = json.Unmarshal(msg2.Data, &workload)
	assert.Nil(t, err)
	assert.Equal(t, "consumer-2", workload)
}
