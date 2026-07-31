package collector_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/etnperlong/mihomo-exporter/internal/collector"
	"github.com/etnperlong/mihomo-exporter/internal/config"
	"github.com/etnperlong/mihomo-exporter/internal/mihomo"
)

func TestProber_ProbeFast(t *testing.T) {
	t.Parallel()

	api := &fakeAPI{
		versionErr: errors.New("version unavailable"),
		proxies:    map[string]mihomo.Proxy{"node": {Name: "node", Type: "Vmess"}},
		providers:  map[string]mihomo.Provider{},
	}
	prober := newProber(t, api, config.ProbeConfig{Concurrency: 2})

	if err := prober.ProbeFast(t.Context()); err == nil {
		t.Fatal("ProbeFast() error = nil, want error")
	}
	first := prober.Snapshot().Fast
	if first.Connected {
		t.Error("first fast snapshot connected = true, want false")
	}
	if got, want := first.Errors["version"], uint64(1); got != want {
		t.Errorf("version error count = %d, want %d", got, want)
	}
	if !first.LastConnected.IsZero() {
		t.Errorf("last connected = %v, want zero time", first.LastConnected)
	}

	api.mu.Lock()
	api.versionErr = nil
	api.version = mihomo.Version{Version: "1.0.0"}
	api.mu.Unlock()
	if err := prober.ProbeFast(t.Context()); err != nil {
		t.Fatalf("ProbeFast() error = %v, want nil", err)
	}
	second := prober.Snapshot().Fast
	if !second.Connected {
		t.Error("second fast snapshot connected = false, want true")
	}
	if second.LastConnected.IsZero() {
		t.Error("second last connected = zero time, want timestamp")
	}
}

func TestProber_ProbeSlow(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	tests := map[string]struct {
		config           config.ProbeConfig
		providers        map[string]mihomo.Provider
		proxies          map[string]mihomo.Proxy
		delays           map[string]uint16
		delayErrors      map[string]error
		wantNodes        map[string]collector.NodeSnapshot
		wantFallback     uint64
		wantHealthCall   int
		wantProviderRead int
		wantErrors       uint64
		wantUnknown      uint64
	}{
		"uses provider extra history and falls back for uncovered nodes": {
			config: config.ProbeConfig{
				Concurrency: 2,
				Staleness:   time.Minute,
				TestURL:     "https://fallback.example/204",
				TestTimeout: time.Second,
			},
			providers: map[string]mihomo.Provider{
				"subscription": {
					Name:        "subscription",
					VehicleType: "HTTP",
					TestURL:     "https://provider.example/204",
					Proxies: []mihomo.Proxy{{
						Name:         "provider-node",
						Type:         "Vmess",
						ProviderName: "subscription",
						Alive:        true,
						History:      []mihomo.Delay{{Time: now, Delay: 900}},
						Extra: map[string]mihomo.ProxyHealth{
							"https://provider.example/204": {Alive: true, History: []mihomo.Delay{{Time: now, Delay: 42}}},
						},
					}},
				},
				"default": {Name: "default", VehicleType: "Compatible"},
			},
			proxies: map[string]mihomo.Proxy{
				"provider-node": {Name: "provider-node", Type: "Vmess", ProviderName: "subscription"},
				"standalone":    {Name: "standalone", Type: "WireGuard", ProviderName: "default"},
				"selector":      {Name: "selector", Type: "Selector"},
			},
			delays: map[string]uint16{"standalone": 77},
			wantNodes: map[string]collector.NodeSnapshot{
				"provider-node": {Name: "provider-node", Type: "Vmess", Provider: "subscription", Delay: 42, Available: true, ProbedAt: now},
				"standalone":    {Name: "standalone", Type: "WireGuard", Provider: "default", Delay: 77, Available: true, ProbedAt: now},
			},
			wantFallback:     1,
			wantHealthCall:   2,
			wantProviderRead: 2,
		},
		"excludes configured types and records failed fallback as unavailable": {
			config: config.ProbeConfig{
				Concurrency:  1,
				Staleness:    time.Minute,
				TestURL:      "https://fallback.example/204",
				TestTimeout:  time.Second,
				ExcludeTypes: "Ssh",
				ExcludeNames: "^notice$",
			},
			providers: map[string]mihomo.Provider{},
			proxies: map[string]mihomo.Proxy{
				"broken":  {Name: "broken", Type: "Trojan"},
				"notice":  {Name: "notice", Type: "Vmess"},
				"ssh":     {Name: "ssh", Type: "Ssh"},
				"unknown": {Name: "unknown", Type: "Unknown"},
			},
			delays:      map[string]uint16{"unknown": 10},
			delayErrors: map[string]error{"broken": errors.New("network unavailable")},
			wantNodes: map[string]collector.NodeSnapshot{
				"broken":  {Name: "broken", Type: "Trojan", Available: false, ProbedAt: now},
				"unknown": {Name: "unknown", Type: "Unknown", Delay: 10, Available: true, ProbedAt: now},
			},
			wantFallback:     2,
			wantErrors:       1,
			wantUnknown:      1,
			wantProviderRead: 2,
		},
	}
	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			api := &fakeAPI{
				providers:   testCase.providers,
				proxies:     testCase.proxies,
				delays:      testCase.delays,
				delayErrors: testCase.delayErrors,
			}
			prober := newProber(t, api, testCase.config)
			if err := prober.ProbeSlow(t.Context()); testCase.wantErrors == 0 && err != nil {
				t.Fatalf("ProbeSlow() error = %v, want nil", err)
			} else if testCase.wantErrors > 0 && err == nil {
				t.Fatal("ProbeSlow() error = nil, want error")
			}

			snapshot := prober.Snapshot()
			for nodeName, want := range testCase.wantNodes {
				got, ok := snapshot.Nodes[nodeName]
				if !ok {
					t.Errorf("snapshot node %q missing", nodeName)
					continue
				}
				if got.Name != want.Name || got.Type != want.Type || got.Provider != want.Provider || got.Delay != want.Delay || got.Available != want.Available {
					t.Errorf("snapshot node %q = %#v, want %#v", nodeName, got, want)
				}
				if got.ProbedAt.IsZero() {
					t.Errorf("snapshot node %q probed at = zero time, want timestamp", nodeName)
				}
			}
			if got, want := snapshot.Slow.FallbackTotal, testCase.wantFallback; got != want {
				t.Errorf("fallback total = %d, want %d", got, want)
			}
			if got, want := api.healthCalls(), testCase.wantHealthCall; got != want {
				t.Errorf("health check calls = %d, want %d", got, want)
			}
			if got, want := api.providerCalls(), testCase.wantProviderRead; got != want {
				t.Errorf("provider reads = %d, want %d", got, want)
			}
			if got, want := snapshot.Slow.Errors["proxy_delay"], testCase.wantErrors; got != want {
				t.Errorf("proxy delay errors = %d, want %d", got, want)
			}
			if got, want := snapshot.Slow.UnknownTypes["Unknown"], testCase.wantUnknown; got != want {
				t.Errorf("unknown proxy type count = %d, want %d", got, want)
			}
		})
	}
}

func TestProber_ProbeSlowLimitsConcurrentRequests(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})
	started := make(chan struct{})
	api := &fakeAPI{
		providers: map[string]mihomo.Provider{
			"one":   {Name: "one", VehicleType: "HTTP"},
			"two":   {Name: "two", VehicleType: "HTTP"},
			"three": {Name: "three", VehicleType: "HTTP"},
		},
		proxies: map[string]mihomo.Proxy{},
	}
	api.healthCheck = func(context.Context, string) error {
		if api.enterProbe() == 2 {
			select {
			case <-started:
			default:
				close(started)
			}
		}
		<-release
		api.leaveProbe()

		return nil
	}
	prober := newProber(t, api, config.ProbeConfig{Concurrency: 2, Staleness: time.Minute, TestTimeout: time.Second})
	done := make(chan error, 1)
	go func() {
		done <- prober.ProbeSlow(t.Context())
	}()
	<-started
	if got, want := api.maximumConcurrentProbes(), 2; got != want {
		t.Errorf("maximum concurrent requests = %d, want %d", got, want)
	}
	close(release)
	if err := <-done; err != nil {
		t.Errorf("ProbeSlow() error = %v, want nil", err)
	}
}

func newProber(t *testing.T, api *fakeAPI, probeConfig config.ProbeConfig) *collector.Prober {
	t.Helper()
	store := collector.NewStore()
	prober, err := collector.NewProber(api, probeConfig, store, nil)
	if err != nil {
		t.Fatalf("NewProber() error = %v", err)
	}

	return prober
}

type fakeAPI struct {
	mu sync.Mutex

	version       mihomo.Version
	versionErr    error
	proxies       map[string]mihomo.Proxy
	providers     map[string]mihomo.Provider
	delays        map[string]uint16
	delayErrors   map[string]error
	healthCheck   func(context.Context, string) error
	healthCount   int
	providerCount int
	active        int
	maximum       int
}

func (f *fakeAPI) Version(context.Context) (mihomo.Version, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.version, f.versionErr
}

func (f *fakeAPI) Proxies(context.Context) (map[string]mihomo.Proxy, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return cloneProxies(f.proxies), nil
}

func (f *fakeAPI) Providers(context.Context) (map[string]mihomo.Provider, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.providerCount++

	return cloneProviders(f.providers), nil
}

func (f *fakeAPI) ProviderHealthCheck(ctx context.Context, name string) error {
	f.mu.Lock()
	f.healthCount++
	hook := f.healthCheck
	f.mu.Unlock()
	if hook != nil {
		return hook(ctx, name)
	}

	return nil
}

func (f *fakeAPI) ProxyDelay(_ context.Context, name, _ string, _ time.Duration) (uint16, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.delays[name], f.delayErrors[name]
}

func (f *fakeAPI) healthCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.healthCount
}

func (f *fakeAPI) providerCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.providerCount
}

func (f *fakeAPI) enterProbe() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.active++
	if f.active > f.maximum {
		f.maximum = f.active
	}

	return f.active
}

func (f *fakeAPI) leaveProbe() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.active--
}

func (f *fakeAPI) maximumConcurrentProbes() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.maximum
}

func cloneProxies(proxies map[string]mihomo.Proxy) map[string]mihomo.Proxy {
	clone := make(map[string]mihomo.Proxy, len(proxies))
	for name, proxy := range proxies {
		clone[name] = proxy
	}

	return clone
}

func cloneProviders(providers map[string]mihomo.Provider) map[string]mihomo.Provider {
	clone := make(map[string]mihomo.Provider, len(providers))
	for name, provider := range providers {
		clone[name] = provider
	}

	return clone
}
