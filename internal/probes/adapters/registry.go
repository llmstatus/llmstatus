// Package adapters holds one Go file per AI-provider adapter plus a
// central registry that maps provider IDs to Provider implementations.
package adapters

import (
	"sort"
	"sync"

	"github.com/llmstatus/llmstatus/internal/probes"
)

var (
	mu       sync.RWMutex
	registry = make(map[string]probes.Provider)
)

// Register associates a Provider with its ID. It is intended to be called
// from each adapter's init() function. Calling Register with an ID that is
// already registered overwrites the previous entry; adapters are expected
// to use unique, snake_case IDs.
func Register(p probes.Provider) {
	mu.Lock()
	defer mu.Unlock()
	registry[p.ID()] = p
}

// Get returns the Provider registered under id, or (nil, false) if no
// adapter has registered that ID.
func Get(id string) (probes.Provider, bool) {
	mu.RLock()
	defer mu.RUnlock()
	p, ok := registry[id]
	return p, ok
}

// All returns every registered provider, sorted by ID. The caller must
// not modify the returned slice's elements' internal state; Provider
// implementations are assumed to be safe for concurrent use.
func All() []probes.Provider {
	mu.RLock()
	defer mu.RUnlock()
	ps := make([]probes.Provider, 0, len(registry))
	for _, p := range registry {
		ps = append(ps, p)
	}
	sort.Slice(ps, func(i, j int) bool { return ps[i].ID() < ps[j].ID() })
	return ps
}
