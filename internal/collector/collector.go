package collector

import (
	"errors"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	errMissingSnapshotReader = errors.New("snapshot reader is required")
	errInvalidStaleness      = errors.New("probe staleness must be greater than zero")
)

// SnapshotReader returns the latest immutable probe snapshot.
type SnapshotReader interface {
	Load() *Snapshot
}

// Collector exposes immutable Mihomo probe snapshots as Prometheus metrics.
type Collector struct {
	store     SnapshotReader
	staleness time.Duration
	descs     descriptors
	build     buildInfo
}

type buildInfo struct {
	version   string
	revision  string
	goVersion string
}

// NewCollector creates a Collector that reads snapshots from store.
func NewCollector(store SnapshotReader, staleness time.Duration) (*Collector, error) {
	if store == nil {
		return nil, errMissingSnapshotReader
	}
	if staleness <= 0 {
		return nil, errInvalidStaleness
	}

	return &Collector{
		store:     store,
		staleness: staleness,
		descs:     newDescriptors(),
		build:     readBuildInfo(),
	}, nil
}

// Describe sends all metric descriptors to ch.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.descs.proxyDelay
	ch <- c.descs.proxyAvailable
	ch <- c.descs.proxyLastProbe
	ch <- c.descs.providerInfo
	ch <- c.descs.providerProxyCount
	ch <- c.descs.providerAvailableCount
	ch <- c.descs.providerUpdated
	ch <- c.descs.providerHealthDuration
	ch <- c.descs.providerHealthSuccess
	ch <- c.descs.up
	ch <- c.descs.versionInfo
	ch <- c.descs.buildInfo
	ch <- c.descs.scrapeErrors
	ch <- c.descs.probeDuration
	ch <- c.descs.lastSuccess
	ch <- c.descs.probeFallback
	ch <- c.descs.unknownProxyType
}

// Collect emits metrics from the latest immutable snapshot without performing I/O.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	snapshot := c.store.Load()
	if snapshot == nil {
		return
	}

	c.collectNodes(ch, snapshot.Nodes)
	c.collectProviders(ch, snapshot)
	c.collectFast(ch, snapshot.Fast)
	c.collectSlow(ch, snapshot.Slow)
	c.collectErrors(ch, snapshot.Fast.Errors, snapshot.Slow.Errors)
}

func (c *Collector) collectNodes(ch chan<- prometheus.Metric, nodes map[string]NodeSnapshot) {
	names := sortedKeys(nodes)
	now := time.Now()
	for _, name := range names {
		node := nodes[name]
		if !node.ProbedAt.IsZero() {
			ch <- prometheus.MustNewConstMetric(c.descs.proxyLastProbe, prometheus.GaugeValue, float64(node.ProbedAt.Unix()), node.Name, node.Type, node.Provider)
		}
		if node.ProbedAt.IsZero() || now.Sub(node.ProbedAt) > c.staleness {
			continue
		}
		ch <- prometheus.MustNewConstMetric(c.descs.proxyAvailable, prometheus.GaugeValue, boolValue(node.Available), node.Name, node.Type, node.Provider)
		if node.Available {
			ch <- prometheus.MustNewConstMetric(c.descs.proxyDelay, prometheus.GaugeValue, float64(node.Delay), node.Name, node.Type, node.Provider)
		}
	}
}

func (c *Collector) collectProviders(ch chan<- prometheus.Metric, snapshot *Snapshot) {
	names := sortedKeys(snapshot.Fast.Providers)
	now := time.Now()
	for _, name := range names {
		provider := snapshot.Fast.Providers[name]
		available := 0
		for _, node := range snapshot.Nodes {
			if node.Provider == name && node.Available && !node.ProbedAt.IsZero() && now.Sub(node.ProbedAt) <= c.staleness {
				available++
			}
		}
		ch <- prometheus.MustNewConstMetric(c.descs.providerInfo, prometheus.GaugeValue, 1, name, provider.VehicleType, provider.Type, provider.TestURL)
		ch <- prometheus.MustNewConstMetric(c.descs.providerProxyCount, prometheus.GaugeValue, float64(len(provider.Proxies)), name)
		ch <- prometheus.MustNewConstMetric(c.descs.providerAvailableCount, prometheus.GaugeValue, float64(available), name)
		if !provider.UpdatedAt.IsZero() {
			ch <- prometheus.MustNewConstMetric(c.descs.providerUpdated, prometheus.GaugeValue, float64(provider.UpdatedAt.Unix()), name)
		}
		if health, ok := snapshot.Slow.HealthChecks[name]; ok {
			ch <- prometheus.MustNewConstMetric(c.descs.providerHealthDuration, prometheus.GaugeValue, health.Duration.Seconds(), name)
			ch <- prometheus.MustNewConstMetric(c.descs.providerHealthSuccess, prometheus.GaugeValue, boolValue(health.Success), name)
		}
	}
}

func (c *Collector) collectFast(ch chan<- prometheus.Metric, fast FastSnapshot) {
	ch <- prometheus.MustNewConstMetric(c.descs.up, prometheus.GaugeValue, boolValue(fast.Connected))
	ch <- prometheus.MustNewConstMetric(c.descs.versionInfo, prometheus.GaugeValue, 1, fast.Version.Version, strconv.FormatBool(fast.Version.Meta))
	ch <- prometheus.MustNewConstMetric(c.descs.buildInfo, prometheus.GaugeValue, 1, c.build.version, c.build.revision, c.build.goVersion)
	ch <- prometheus.MustNewConstMetric(c.descs.probeDuration, prometheus.GaugeValue, fast.ProbeDuration.Seconds(), "fast")
	if !fast.LastConnected.IsZero() {
		ch <- prometheus.MustNewConstMetric(c.descs.lastSuccess, prometheus.GaugeValue, float64(fast.LastConnected.Unix()), "fast")
	}
}

func (c *Collector) collectSlow(ch chan<- prometheus.Metric, slow SlowSnapshot) {
	ch <- prometheus.MustNewConstMetric(c.descs.probeDuration, prometheus.GaugeValue, slow.ProbeDuration.Seconds(), "slow")
	if !slow.LastSuccessful.IsZero() {
		ch <- prometheus.MustNewConstMetric(c.descs.lastSuccess, prometheus.GaugeValue, float64(slow.LastSuccessful.Unix()), "slow")
	}
	ch <- prometheus.MustNewConstMetric(c.descs.probeFallback, prometheus.CounterValue, float64(slow.FallbackTotal))
	for _, proxyType := range sortedKeys(slow.UnknownTypes) {
		ch <- prometheus.MustNewConstMetric(c.descs.unknownProxyType, prometheus.CounterValue, float64(slow.UnknownTypes[proxyType]), proxyType)
	}
}

func (c *Collector) collectErrors(ch chan<- prometheus.Metric, fastErrors, slowErrors map[string]uint64) {
	errors := make(map[string]uint64, len(fastErrors)+len(slowErrors))
	for collector, count := range fastErrors {
		errors[collector] += count
	}
	for collector, count := range slowErrors {
		errors[collector] += count
	}
	for _, collector := range sortedKeys(errors) {
		ch <- prometheus.MustNewConstMetric(c.descs.scrapeErrors, prometheus.CounterValue, float64(errors[collector]), collector)
	}
}

func readBuildInfo() buildInfo {
	info := buildInfo{version: "(devel)", revision: "unknown", goVersion: runtime.Version()}
	if build, ok := debug.ReadBuildInfo(); ok {
		if build.Main.Version != "" {
			info.version = build.Main.Version
		}
		for _, setting := range build.Settings {
			if setting.Key == "vcs.revision" && setting.Value != "" {
				info.revision = setting.Value
			}
		}
	}

	return info
}

func boolValue(value bool) float64 {
	if value {
		return 1
	}

	return 0
}

func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return keys
}
