package collector

import (
	"context"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/etnperlong/mihomo-exporter/internal/config"
	"github.com/etnperlong/mihomo-exporter/internal/mihomo"
)

func TestProber_StartSlowSkipsReentryAndCancels(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		api := &blockingAPI{started: make(chan struct{})}
		prober, err := NewProber(api, config.ProbeConfig{
			Concurrency:  1,
			Staleness:    time.Minute,
			TestTimeout:  time.Second,
			FastInterval: time.Minute,
			SlowInterval: time.Minute,
		}, NewStore(), nil)
		if err != nil {
			t.Fatalf("NewProber() error = %v", err)
		}

		var probes errgroup.Group
		probes.SetLimit(1)
		if started := prober.startSlow(ctx, &probes); !started {
			t.Fatal("first startSlow() = false, want true")
		}
		<-api.started
		synctest.Wait()
		if started := prober.startSlow(ctx, &probes); started {
			t.Error("second startSlow() = true, want false while first probe runs")
		}

		cancel()
		if err := probes.Wait(); err != nil {
			t.Errorf("wait for canceled slow probe = %v, want nil", err)
		}
		if got, want := api.healthCalls(), 1; got != want {
			t.Errorf("health check calls = %d, want %d", got, want)
		}
	})
}

type blockingAPI struct {
	mu      sync.Mutex
	started chan struct{}
	calls   int
}

func (f *blockingAPI) Version(context.Context) (mihomo.Version, error) {
	return mihomo.Version{}, nil
}

func (f *blockingAPI) Proxies(context.Context) (map[string]mihomo.Proxy, error) {
	return map[string]mihomo.Proxy{}, nil
}

func (f *blockingAPI) Providers(context.Context) (map[string]mihomo.Provider, error) {
	return map[string]mihomo.Provider{"provider": {Name: "provider", VehicleType: "HTTP"}}, nil
}

func (f *blockingAPI) ProviderHealthCheck(ctx context.Context, _ string) error {
	f.mu.Lock()
	f.calls++
	if f.calls == 1 {
		close(f.started)
	}
	f.mu.Unlock()
	<-ctx.Done()

	return ctx.Err()
}

func (f *blockingAPI) ProxyDelay(context.Context, string, string, time.Duration) (uint16, error) {
	return 0, nil
}

func (f *blockingAPI) healthCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.calls
}
