package collector

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/etnperlong/mihomo-exporter/internal/config"
	"github.com/etnperlong/mihomo-exporter/internal/mihomo"
)

// mihomoAPI is the Mihomo API subset required by Prober.
type mihomoAPI interface {
	Version(ctx context.Context) (mihomo.Version, error)
	Proxies(ctx context.Context) (map[string]mihomo.Proxy, error)
	Providers(ctx context.Context) (map[string]mihomo.Provider, error)
	ProviderHealthCheck(ctx context.Context, name string) error
	ProxyDelay(ctx context.Context, name, testURL string, timeout time.Duration) (uint16, error)
}

var (
	errMissingAPI         = errors.New("mihomo API is required")
	errMissingStore       = errors.New("snapshot store is required")
	errInvalidConcurrency = errors.New("probe concurrency must be at least one")
)

// Prober updates a Store with fast topology and slow delay probes.
type Prober struct {
	api          mihomoAPI
	config       config.ProbeConfig
	store        *Store
	logger       *zap.Logger
	excludeNames *regexp.Regexp
	excludeTypes map[string]struct{}
	slowRunning  atomic.Bool
}

// NewProber creates a Prober that publishes results to store.
func NewProber(api mihomoAPI, probeConfig config.ProbeConfig, store *Store, logger *zap.Logger) (*Prober, error) {
	if api == nil {
		return nil, errMissingAPI
	}
	if store == nil {
		return nil, errMissingStore
	}
	if probeConfig.Concurrency < 1 {
		return nil, errInvalidConcurrency
	}

	var excludeNames *regexp.Regexp
	if probeConfig.ExcludeNames != "" {
		var err error
		excludeNames, err = regexp.Compile(probeConfig.ExcludeNames)
		if err != nil {
			return nil, fmt.Errorf("compile excluded proxy names: %w", err)
		}
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Prober{
		api:          api,
		config:       probeConfig,
		store:        store,
		logger:       logger,
		excludeNames: excludeNames,
		excludeTypes: splitTypes(probeConfig.ExcludeTypes),
	}, nil
}

// Snapshot returns the latest immutable probe snapshot.
func (p *Prober) Snapshot() *Snapshot {
	return p.store.Load()
}

// Run starts the fast and slow probe loops and returns after ctx is canceled.
func (p *Prober) Run(ctx context.Context) error {
	var loops errgroup.Group
	loops.SetLimit(2)
	loops.Go(func() error {
		p.runFastLoop(ctx)

		return nil
	})
	loops.Go(func() error {
		p.runSlowLoop(ctx)

		return nil
	})

	if err := loops.Wait(); err != nil {
		return fmt.Errorf("wait for probe loops: %w", err)
	}

	return nil
}

// ProbeFast performs one topology probe and atomically publishes its result.
func (p *Prober) ProbeFast(ctx context.Context) error {
	started := time.Now()
	var (
		version   mihomo.Version
		proxies   map[string]mihomo.Proxy
		providers map[string]mihomo.Provider
		errs      = make(map[string]error)
		mutex     sync.Mutex
	)

	var group errgroup.Group
	group.SetLimit(p.config.Concurrency)
	group.Go(func() error {
		result, err := p.api.Version(ctx)
		mutex.Lock()
		version = result
		errs["version"] = err
		mutex.Unlock()

		return nil
	})
	group.Go(func() error {
		result, err := p.api.Proxies(ctx)
		mutex.Lock()
		proxies = result
		errs["proxies"] = err
		mutex.Unlock()

		return nil
	})
	group.Go(func() error {
		result, err := p.api.Providers(ctx)
		mutex.Lock()
		providers = result
		errs["providers"] = err
		mutex.Unlock()

		return nil
	})
	if err := group.Wait(); err != nil {
		return fmt.Errorf("wait for fast probes: %w", err)
	}

	connected := errs["version"] == nil && errs["proxies"] == nil && errs["providers"] == nil
	now := time.Now()
	p.store.update(func(current Snapshot) Snapshot {
		next := current
		next.Fast.Errors = cloneCounters(current.Fast.Errors)
		for collector, err := range errs {
			if err != nil {
				next.Fast.Errors[collector]++
			}
		}
		if errs["version"] == nil {
			next.Fast.Version = version
		}
		if errs["proxies"] == nil {
			next.Fast.Proxies = cloneProxies(proxies)
		}
		if errs["providers"] == nil {
			next.Fast.Providers = cloneProviders(providers)
		}
		next.Fast.Connected = connected
		next.Fast.ProbeDuration = now.Sub(started)
		if connected {
			next.Fast.LastConnected = now
		}

		return next
	})

	return errors.Join(errs["version"], errs["proxies"], errs["providers"])
}

// ProbeSlow performs one health-check and fallback-delay probe iteration.
func (p *Prober) ProbeSlow(ctx context.Context) error {
	started := time.Now()
	providers, err := p.api.Providers(ctx)
	if err != nil {
		p.recordSlowFailure("providers", time.Since(started))

		return fmt.Errorf("get providers before health checks: %w", err)
	}

	healthChecks, healthErrors := p.healthCheckProviders(ctx, providers)
	updatedProviders, providerErr := p.api.Providers(ctx)
	if providerErr != nil {
		p.recordSlowFailure("providers", time.Since(started))

		return fmt.Errorf("get providers after health checks: %w", providerErr)
	}
	allProxies, proxyErr := p.api.Proxies(ctx)
	if proxyErr != nil {
		p.recordSlowFailure("proxies", time.Since(started))

		return fmt.Errorf("get proxies for fallback probes: %w", proxyErr)
	}

	now := time.Now()
	nodes, covered := p.providerNodes(updatedProviders, now)
	fallbackNodes, fallbackErrors := p.fallbackNodes(ctx, allProxies, covered, now)
	for name, node := range fallbackNodes {
		nodes[name] = node
	}

	allErrors := append(healthErrors, fallbackErrors...)
	p.store.update(func(current Snapshot) Snapshot {
		next := current
		next.Nodes = nodes
		next.Slow.Errors = cloneCounters(current.Slow.Errors)
		next.Slow.UnknownTypes = cloneCounters(current.Slow.UnknownTypes)
		next.Slow.HealthChecks = healthChecks
		next.Slow.FallbackTotal = current.Slow.FallbackTotal + uint64(len(fallbackNodes))
		next.Slow.Connected = len(allErrors) == 0
		next.Slow.ProbeDuration = now.Sub(started)
		if next.Slow.Connected {
			next.Slow.LastSuccessful = now
		}
		for _, probeErr := range allErrors {
			next.Slow.Errors[probeErr.collector]++
		}
		for _, node := range nodes {
			if node.Type == "Unknown" {
				next.Slow.UnknownTypes[node.Type]++
			}
		}

		return next
	})

	return joinProbeErrors(allErrors)
}

func (p *Prober) runFastLoop(ctx context.Context) {
	p.probeFastAndLog(ctx)
	ticker := time.NewTicker(p.config.FastInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.probeFastAndLog(ctx)
		}
	}
}

func (p *Prober) runSlowLoop(ctx context.Context) {
	var probes errgroup.Group
	probes.SetLimit(1)

	p.startSlow(ctx, &probes)
	ticker := time.NewTicker(p.config.SlowInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			if err := probes.Wait(); err != nil {
				p.logger.Error("wait for slow probe", zap.Error(err))
			}

			return
		case <-ticker.C:
			p.startSlow(ctx, &probes)
		}
	}
}

func (p *Prober) startSlow(ctx context.Context, probes *errgroup.Group) bool {
	if !p.slowRunning.CompareAndSwap(false, true) {
		return false
	}
	probes.Go(func() error {
		defer p.slowRunning.Store(false)
		p.probeSlowAndLog(ctx)

		return nil
	})

	return true
}

func (p *Prober) probeFastAndLog(ctx context.Context) {
	if err := p.ProbeFast(ctx); err != nil {
		p.logger.Warn("fast probe failed", zap.Error(err))
	}
}

func (p *Prober) probeSlowAndLog(ctx context.Context) {
	if err := p.ProbeSlow(ctx); err != nil {
		p.logger.Warn("slow probe failed", zap.Error(err))
	}
}

func (p *Prober) healthCheckProviders(ctx context.Context, providers map[string]mihomo.Provider) (map[string]ProviderHealthSnapshot, []probeError) {
	names := make([]string, 0, len(providers))
	for name, provider := range providers {
		if healthCheckable(provider.VehicleType) {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	var (
		group   errgroup.Group
		mutex   sync.Mutex
		errors  []probeError
		results = make(map[string]ProviderHealthSnapshot, len(names))
	)
	group.SetLimit(p.config.Concurrency)
	for _, name := range names {
		group.Go(func() error {
			started := time.Now()
			if err := p.api.ProviderHealthCheck(ctx, name); err != nil {
				mutex.Lock()
				errors = append(errors, probeError{collector: "provider_healthcheck", err: fmt.Errorf("health check provider %q: %w", name, err)})
				results[name] = ProviderHealthSnapshot{Duration: time.Since(started)}
				mutex.Unlock()

				return nil
			}
			mutex.Lock()
			results[name] = ProviderHealthSnapshot{Duration: time.Since(started), Success: true}
			mutex.Unlock()

			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return results, append(errors, probeError{collector: "provider_healthcheck", err: fmt.Errorf("wait for provider health checks: %w", err)})
	}

	return results, errors
}

func (p *Prober) providerNodes(providers map[string]mihomo.Provider, now time.Time) (map[string]NodeSnapshot, map[string]struct{}) {
	nodes := make(map[string]NodeSnapshot)
	covered := make(map[string]struct{})
	names := make([]string, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		provider := providers[name]
		for _, proxy := range provider.Proxies {
			if !p.shouldProbe(proxy) {
				continue
			}
			delay, ok := proxy.LastDelay(provider.TestURL)
			if !ok || delay.Time.IsZero() || now.Sub(delay.Time) > p.config.Staleness {
				continue
			}
			providerName := proxy.ProviderName
			if providerName == "" {
				providerName = provider.Name
			}
			if providerName == "" {
				providerName = name
			}
			nodes[proxy.Name] = NodeSnapshot{
				Name:      proxy.Name,
				Type:      proxy.Type,
				Provider:  providerName,
				Delay:     delay.Delay,
				Available: proxy.Health(provider.TestURL).Alive && delay.Delay != 0,
				ProbedAt:  delay.Time,
			}
			covered[proxy.Name] = struct{}{}
		}
	}

	return nodes, covered
}

func (p *Prober) fallbackNodes(ctx context.Context, proxies map[string]mihomo.Proxy, covered map[string]struct{}, now time.Time) (map[string]NodeSnapshot, []probeError) {
	names := make([]string, 0, len(proxies))
	for name, proxy := range proxies {
		if _, ok := covered[name]; !ok && p.shouldProbe(proxy) {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	var (
		group  errgroup.Group
		mutex  sync.Mutex
		nodes  = make(map[string]NodeSnapshot, len(names))
		errors []probeError
	)
	group.SetLimit(p.config.Concurrency)
	for _, name := range names {
		proxy := proxies[name]
		group.Go(func() error {
			delay, err := p.api.ProxyDelay(ctx, name, p.config.TestURL, p.config.TestTimeout)
			node := NodeSnapshot{
				Name:      name,
				Type:      proxy.Type,
				Provider:  proxy.ProviderName,
				Delay:     delay,
				Available: err == nil && delay != 0,
				ProbedAt:  now,
			}
			mutex.Lock()
			nodes[name] = node
			if err != nil {
				errors = append(errors, probeError{collector: "proxy_delay", err: fmt.Errorf("probe proxy %q: %w", name, err)})
			}
			mutex.Unlock()

			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nodes, append(errors, probeError{collector: "proxy_delay", err: fmt.Errorf("wait for fallback probes: %w", err)})
	}

	return nodes, errors
}

func (p *Prober) shouldProbe(proxy mihomo.Proxy) bool {
	if !mihomo.IsOutboundNode(proxy.Type) || p.excludeNames != nil && p.excludeNames.MatchString(proxy.Name) {
		return false
	}
	_, excluded := p.excludeTypes[proxy.Type]

	return !excluded
}

func (p *Prober) recordSlowFailure(collector string, duration time.Duration) {
	p.store.update(func(current Snapshot) Snapshot {
		next := current
		next.Slow.Connected = false
		next.Slow.ProbeDuration = duration
		next.Slow.Errors = cloneCounters(current.Slow.Errors)
		next.Slow.Errors[collector]++

		return next
	})
}

func cloneCounters(counters map[string]uint64) map[string]uint64 {
	clone := make(map[string]uint64, len(counters))
	for key, value := range counters {
		clone[key] = value
	}

	return clone
}

func cloneProxies(proxies map[string]mihomo.Proxy) map[string]mihomo.Proxy {
	clone := make(map[string]mihomo.Proxy, len(proxies))
	for name, proxy := range proxies {
		clone[name] = cloneProxy(proxy)
	}

	return clone
}

func cloneProviders(providers map[string]mihomo.Provider) map[string]mihomo.Provider {
	clone := make(map[string]mihomo.Provider, len(providers))
	for name, provider := range providers {
		provider.Proxies = make([]mihomo.Proxy, len(provider.Proxies))
		for index, proxy := range providers[name].Proxies {
			provider.Proxies[index] = cloneProxy(proxy)
		}
		clone[name] = provider
	}

	return clone
}

func cloneProxy(proxy mihomo.Proxy) mihomo.Proxy {
	proxy.History = append([]mihomo.Delay(nil), proxy.History...)
	if proxy.Extra == nil {
		return proxy
	}
	proxy.Extra = make(map[string]mihomo.ProxyHealth, len(proxy.Extra))
	for testURL, health := range proxy.Extra {
		health.History = append([]mihomo.Delay(nil), health.History...)
		proxy.Extra[testURL] = health
	}

	return proxy
}

func splitTypes(value string) map[string]struct{} {
	types := make(map[string]struct{})
	for _, proxyType := range strings.Split(value, ",") {
		if proxyType = strings.TrimSpace(proxyType); proxyType != "" {
			types[proxyType] = struct{}{}
		}
	}

	return types
}

func healthCheckable(vehicleType string) bool {
	switch vehicleType {
	case "HTTP", "File", "Inline", "Compatible":
		return true
	default:
		return false
	}
}

type probeError struct {
	collector string
	err       error
}

func joinProbeErrors(probeErrors []probeError) error {
	joined := make([]error, 0, len(probeErrors))
	for _, probeErr := range probeErrors {
		joined = append(joined, probeErr.err)
	}

	return errors.Join(joined...)
}
