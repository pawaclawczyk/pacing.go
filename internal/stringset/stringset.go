package stringset

import "sync"

// StringSet is a data structure representing a set (collection of unique values) of strings
type StringSet struct {
	mu  sync.RWMutex
	set map[string]bool
}

// NewStringSet creates empty StringSet
func NewStringSet() *StringSet {
	return &StringSet{
		mu:  sync.RWMutex{},
		set: map[string]bool{},
	}
}

// NewStringSet2 creates empty StringSet
func NewStringSet2() *StringSet {
	return &StringSet{
		mu:  sync.RWMutex{},
		set: make(map[string]bool, 1024),
	}
}

// Add adds a string to a set
func (set *StringSet) Add(s string) {
	set.mu.Lock()
	defer set.mu.Unlock()
	set.set[s] = true
}

// List returns all strings in a set
func (set *StringSet) List() []string {
	set.mu.RLock()
	defer set.mu.RUnlock()
	res := make([]string, len(set.set))[:0]
	for v := range set.set {
		res = append(res, v)
	}
	return res
}
