package stringset

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewStringSetIsEmpty(t *testing.T) {
	tests := []struct {
		name         string
		newStringSet func() *StringSet
	}{
		{"NewStringSet", NewStringSet},
		{"NewStringSet (with allocation)", NewStringSet2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := tt.newStringSet()
			assert.Len(t, set.List(), 0)
		})
	}
}

func BenchmarkNewStringSet(b *testing.B) {
	tests := []struct {
		name         string
		newStringSet func() *StringSet
	}{
		{"NewStringSet", NewStringSet},
		{"NewStringSet (with allocation)", NewStringSet2},
	}
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = tt.newStringSet()
			}
		})
	}
}
