package collector

import (
	"sync/atomic"

	"github.com/etnperlong/mihomo-exporter/internal/mihomo"
)

// Store atomically publishes immutable probe snapshots.
type Store struct {
	snapshot atomic.Pointer[Snapshot]
}

// NewStore creates a Store with an empty snapshot.
func NewStore() *Store {
	store := &Store{}
	store.snapshot.Store(&Snapshot{
		Fast: FastSnapshot{
			Proxies:   make(map[string]mihomo.Proxy),
			Providers: make(map[string]mihomo.Provider),
			Errors:    make(map[string]uint64),
		},
		Slow: SlowSnapshot{
			Errors:       make(map[string]uint64),
			UnknownTypes: make(map[string]uint64),
			HealthChecks: make(map[string]ProviderHealthSnapshot),
		},
		Nodes: make(map[string]NodeSnapshot),
	})

	return store
}

// Load returns the current immutable snapshot.
func (s *Store) Load() *Snapshot {
	return s.snapshot.Load()
}

func (s *Store) update(transform func(Snapshot) Snapshot) {
	for {
		current := s.snapshot.Load()
		next := transform(*current)
		if s.snapshot.CompareAndSwap(current, &next) {
			return
		}
	}
}
