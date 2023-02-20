package dispatcher

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
	"time"
)

func TestAddAndListConsumers(t *testing.T) {
	cs := newConsumers()
	cs.add("consumer-1")
	cs.add("consumer-2")
	assert.Len(t, cs.list(), 2)
	assert.Contains(t, cs.list(), "consumer-1")
	assert.Contains(t, cs.list(), "consumer-2")
	cs.add("consumer-2")
	assert.Len(t, cs.list(), 2)
}

func TestTTLCleaner(t *testing.T) {
	cs := newConsumers()
	cs.ttl = 5 * time.Millisecond
	cs.checkTTLPeriod = time.Millisecond
	cs.enableTTL()

	cs.add("consumer-1")
	assert.Len(t, cs.list(), 1)
	assert.Contains(t, cs.list(), "consumer-1")

	time.Sleep(2 * time.Millisecond)
	cs.add("consumer-2")
	assert.Len(t, cs.list(), 2)
	assert.Contains(t, cs.list(), "consumer-1")
	assert.Contains(t, cs.list(), "consumer-2")

	time.Sleep(4 * time.Millisecond)
	assert.Len(t, cs.list(), 1)
	assert.Contains(t, cs.list(), "consumer-2")

	time.Sleep(2 * time.Millisecond)
	assert.Len(t, cs.list(), 0)
}

func TestTTLCleanerNotEnabled(t *testing.T) {
	cs := newConsumers()
	cs.ttl = 5 * time.Millisecond
	cs.checkTTLPeriod = time.Millisecond

	cs.add("consumer-1")
	assert.Len(t, cs.list(), 1)
	assert.Contains(t, cs.list(), "consumer-1")

	time.Sleep(10 * time.Millisecond)
	assert.Len(t, cs.list(), 1)
	assert.Contains(t, cs.list(), "consumer-1")
}

func TestTTLCleanerDisabled(t *testing.T) {
	cs := newConsumers()
	cs.ttl = 5 * time.Millisecond
	cs.checkTTLPeriod = time.Millisecond
	cs.enableTTL()
	cs.disableTTL()

	cs.add("consumer-1")
	assert.Len(t, cs.list(), 1)
	assert.Contains(t, cs.list(), "consumer-1")

	time.Sleep(10 * time.Millisecond)
	assert.Len(t, cs.list(), 1)
	assert.Contains(t, cs.list(), "consumer-1")
}

func TestEnableTTLTwice(t *testing.T) {
	cs := newConsumers()
	cs.ttl = 5 * time.Millisecond
	cs.checkTTLPeriod = time.Millisecond
	cs.enableTTL()
	cs.enableTTL()
}

func TestDisableTTLWhenNotEnabled(t *testing.T) {
	cs := newConsumers()
	cs.ttl = 5 * time.Millisecond
	cs.checkTTLPeriod = time.Millisecond
	cs.disableTTL()
}

func TestDisableTTLTwice(t *testing.T) {
	cs := newConsumers()
	cs.ttl = 5 * time.Millisecond
	cs.checkTTLPeriod = time.Millisecond
	cs.enableTTL()
	cs.disableTTL()
	cs.disableTTL()
}

func BenchmarkConsumers(b *testing.B) {
	var m *consumers

	m = newConsumers()
	b.Run("add", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m.add(fmt.Sprintf("item-%d", i))
		}
	})

	m = newConsumers()
	for i := 0; i < 100; i++ {
		m.add(fmt.Sprintf("item-%d", i))
	}
	b.Run("list_100", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m.list()
		}
	})
}

func TestDeleteOutdated(t *testing.T) {
	refT := time.Date(2023, 2, 20, 13, 38, 0, 0, time.Local)
	type args struct {
		consumers map[string]time.Time
		t         time.Time
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"empty", args{map[string]time.Time{}, refT}, []string{}},
		{"before", args{map[string]time.Time{"consumer-a": refT, "consumer-b": refT}, refT.Add(-time.Second)}, []string{"consumer-a", "consumer-b"}},
		{"equal to older", args{map[string]time.Time{"consumer-a": refT, "consumer-b": refT.Add(time.Second)}, refT}, []string{"consumer-b"}},
		{"between", args{map[string]time.Time{"consumer-a": refT, "consumer-b": refT.Add(2 * time.Second)}, refT.Add(time.Second)}, []string{"consumer-b"}},
		{"equal to newer", args{map[string]time.Time{"consumer-a": refT, "consumer-b": refT.Add(time.Second)}, refT.Add(time.Second)}, []string{}},
		{"after", args{map[string]time.Time{"consumer-a": refT, "consumer-b": refT}, refT.Add(time.Second)}, []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := newConsumers()
			cs.consumers = tt.args.consumers
			cs.deleteOutdated(tt.args.t)
			sort.Strings(tt.want)
			actual := cs.list()
			sort.Strings(actual)
			assert.Equal(t, tt.want, actual)
		})
	}
}
