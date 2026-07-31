// Package collector maintains asynchronous Mihomo probe snapshots.
package collector

import (
	"time"

	"github.com/etnperlong/mihomo-exporter/internal/mihomo"
)

// Snapshot is an immutable view of Mihomo topology and probe results.
type Snapshot struct {
	Fast  FastSnapshot
	Slow  SlowSnapshot
	Nodes map[string]NodeSnapshot
}

// FastSnapshot contains state collected by the fast probe loop.
type FastSnapshot struct {
	Version       mihomo.Version
	Proxies       map[string]mihomo.Proxy
	Providers     map[string]mihomo.Provider
	Connected     bool
	LastConnected time.Time
	ProbeDuration time.Duration
	Errors        map[string]uint64
}

// SlowSnapshot contains state collected by the slow probe loop.
type SlowSnapshot struct {
	Connected      bool
	LastSuccessful time.Time
	ProbeDuration  time.Duration
	Errors         map[string]uint64
	FallbackTotal  uint64
	UnknownTypes   map[string]uint64
	HealthChecks   map[string]ProviderHealthSnapshot
}

// ProviderHealthSnapshot is the outcome of a provider health check.
type ProviderHealthSnapshot struct {
	Duration time.Duration
	Success  bool
}

// NodeSnapshot is the latest probe result for one outbound proxy.
type NodeSnapshot struct {
	Name      string
	Type      string
	Provider  string
	Delay     uint16
	Available bool
	ProbedAt  time.Time
}
