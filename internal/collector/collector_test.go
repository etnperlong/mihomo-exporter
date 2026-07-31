package collector_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/etnperlong/mihomo-exporter/internal/collector"
	"github.com/etnperlong/mihomo-exporter/internal/mihomo"
)

func TestCollector_CollectAndCompare(t *testing.T) {
	t.Parallel()

	probeTime := time.Date(2100, time.January, 1, 0, 0, 0, 0, time.UTC)
	collector := newCollector(t, collector.Snapshot{
		Fast: collector.FastSnapshot{
			Version:       mihomo.Version{Version: "v1.2.3", Meta: true},
			Connected:     true,
			LastConnected: probeTime,
			ProbeDuration: 125 * time.Millisecond,
			Errors:        map[string]uint64{"version": 1},
			Providers: map[string]mihomo.Provider{
				"primary": {
					Name:        "primary",
					Type:        "Proxy",
					VehicleType: "HTTP",
					TestURL:     "https://example.test/204",
					UpdatedAt:   probeTime,
					Proxies:     []mihomo.Proxy{{Name: "healthy"}, {Name: "failed"}},
				},
			},
		},
		Slow: collector.SlowSnapshot{
			LastSuccessful: probeTime,
			ProbeDuration:  250 * time.Millisecond,
			Errors:         map[string]uint64{"provider_healthcheck": 2},
			FallbackTotal:  3,
			UnknownTypes:   map[string]uint64{"Novel": 4},
			HealthChecks: map[string]collector.ProviderHealthSnapshot{
				"primary": {Duration: 50 * time.Millisecond, Success: true},
			},
		},
		Nodes: map[string]collector.NodeSnapshot{
			"failed":  {Name: "failed", Type: "Trojan", Provider: "primary", Available: false, ProbedAt: probeTime},
			"healthy": {Name: "healthy", Type: "Vmess", Provider: "primary", Delay: 42, Available: true, ProbedAt: probeTime},
		},
	})
	want := testdata(t, "golden/collector.metrics")

	if err := testutil.CollectAndCompare(collector, strings.NewReader(want),
		"mihomo_proxy_delay_milliseconds",
		"mihomo_proxy_available",
		"mihomo_proxy_last_probe_timestamp_seconds",
		"mihomo_provider_info",
		"mihomo_provider_proxy_count",
		"mihomo_provider_proxy_available_count",
		"mihomo_provider_updated_timestamp_seconds",
		"mihomo_provider_healthcheck_duration_seconds",
		"mihomo_provider_healthcheck_success",
		"mihomo_up",
		"mihomo_version_info",
		"mihomo_exporter_scrape_errors_total",
		"mihomo_exporter_probe_duration_seconds",
		"mihomo_exporter_last_success_timestamp_seconds",
		"mihomo_exporter_probe_fallback_total",
		"mihomo_exporter_unknown_proxy_type_total",
	); err != nil {
		t.Errorf("CollectAndCompare() error = %v", err)
	}
}

func TestCollector_CollectReadsOnlySnapshot(t *testing.T) {
	t.Parallel()

	source := &snapshotSource{snapshot: &collector.Snapshot{Fast: collector.FastSnapshot{Errors: map[string]uint64{}}, Slow: collector.SlowSnapshot{Errors: map[string]uint64{}, UnknownTypes: map[string]uint64{}}}}
	metricCollector, err := collector.NewCollector(source, time.Minute)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}
	want := "# HELP mihomo_up Whether the latest fast probe connected to Mihomo.\n# TYPE mihomo_up gauge\nmihomo_up 0\n"

	if err := testutil.CollectAndCompare(metricCollector, strings.NewReader(want), "mihomo_up"); err != nil {
		t.Errorf("CollectAndCompare() error = %v", err)
	}
	if got, want := source.loads, 1; got != want {
		t.Errorf("snapshot loads = %d, want %d", got, want)
	}
}

func TestCollector_StaleNodeKeepsOnlyLastProbe(t *testing.T) {
	t.Parallel()

	metricCollector := newCollector(t, collector.Snapshot{Nodes: map[string]collector.NodeSnapshot{
		"stale": {Name: "stale", Type: "Vmess", Provider: "primary", Delay: 42, Available: true, ProbedAt: time.Unix(1, 0)},
	}})
	want := "# HELP mihomo_proxy_last_probe_timestamp_seconds Unix timestamp of the latest proxy probe.\n# TYPE mihomo_proxy_last_probe_timestamp_seconds gauge\nmihomo_proxy_last_probe_timestamp_seconds{name=\"stale\",provider=\"primary\",type=\"Vmess\"} 1\n"

	if err := testutil.CollectAndCompare(metricCollector, strings.NewReader(want), "mihomo_proxy_delay_milliseconds", "mihomo_proxy_available", "mihomo_proxy_last_probe_timestamp_seconds"); err != nil {
		t.Errorf("CollectAndCompare() error = %v", err)
	}
}

func TestCollector_FailedNodeEmitsAvailabilityWithoutDelay(t *testing.T) {
	t.Parallel()

	probeTime := time.Date(2100, time.January, 1, 0, 0, 0, 0, time.UTC)
	metricCollector := newCollector(t, collector.Snapshot{Nodes: map[string]collector.NodeSnapshot{
		"failed": {Name: "failed", Type: "Trojan", Provider: "primary", Available: false, ProbedAt: probeTime},
	}})
	want := "# HELP mihomo_proxy_available Whether the proxy is available.\n# TYPE mihomo_proxy_available gauge\nmihomo_proxy_available{name=\"failed\",provider=\"primary\",type=\"Trojan\"} 0\n"

	if err := testutil.CollectAndCompare(metricCollector, strings.NewReader(want), "mihomo_proxy_delay_milliseconds", "mihomo_proxy_available"); err != nil {
		t.Errorf("CollectAndCompare() error = %v", err)
	}
}

func TestCollector_UpReflectsFailedFastProbe(t *testing.T) {
	t.Parallel()

	metricCollector := newCollector(t, collector.Snapshot{Fast: collector.FastSnapshot{Errors: map[string]uint64{}}, Slow: collector.SlowSnapshot{Errors: map[string]uint64{}, UnknownTypes: map[string]uint64{}}})
	want := "# HELP mihomo_up Whether the latest fast probe connected to Mihomo.\n# TYPE mihomo_up gauge\nmihomo_up 0\n"

	if err := testutil.CollectAndCompare(metricCollector, strings.NewReader(want), "mihomo_up"); err != nil {
		t.Errorf("CollectAndCompare() error = %v", err)
	}
}

func newCollector(t *testing.T, snapshot collector.Snapshot) *collector.Collector {
	t.Helper()

	metricCollector, err := collector.NewCollector(&snapshotSource{snapshot: &snapshot}, time.Minute)
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}

	return metricCollector
}

func testdata(t *testing.T, name string) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", name, err)
	}

	return string(data)
}

type snapshotSource struct {
	snapshot *collector.Snapshot
	loads    int
}

func (s *snapshotSource) Load() *collector.Snapshot {
	s.loads++

	return s.snapshot
}
