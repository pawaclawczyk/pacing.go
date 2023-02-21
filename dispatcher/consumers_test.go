package dispatcher

import (
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
	"time"
)

func TestNewConsumers(t *testing.T) {
	cs := NewConsumers()
	assert.Equal(t, cs.opts.ttl, DefaultTTL)
	assert.Len(t, cs.ttlSet, 0)
}

func TestConsumers_Join(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{"none", []string{}, []string{}},
		{"alice bob", []string{"alice", "bob"}, []string{"alice", "bob"}},
		{"alice alice", []string{"alice", "alice"}, []string{"alice"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure test is not affected by expiring cons
			cs := NewConsumers(WithTTL(time.Hour))
			for _, con := range tt.args {
				cs.Join(con)
			}
			got := cs.List()
			sort.Strings(got)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConsumers_Leave(t *testing.T) {
	type args struct {
		current []string
		leaving []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"{} - {alice}", args{[]string{}, []string{"alice"}}, []string{}},
		{"{alice} - {alice}", args{[]string{"alice"}, []string{"alice"}}, []string{}},
		{"{alice, bob} - {alice}", args{[]string{"alice", "bob"}, []string{"alice"}}, []string{"bob"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure test is not affected by expiring cons
			cs := NewConsumers(WithTTL(time.Hour))
			for _, con := range tt.args.current {
				cs.Join(con)
			}
			for _, con := range tt.args.leaving {
				cs.Leave(con)
			}
			got := cs.List()
			sort.Strings(got)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithTTL(t *testing.T) {
	tests := []struct {
		name string
		args time.Duration
		want time.Duration
	}{
		{"positive", time.Second, time.Second},
		{"zero", 0, DefaultTTL},
		{"negative", -time.Second, DefaultTTL},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := NewConsumers(WithTTL(tt.args))
			assert.Equal(t, tt.want, cs.opts.ttl)
		})
	}
}
