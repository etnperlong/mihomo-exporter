package mihomo_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/etnperlong/mihomo-exporter/internal/mihomo"
)

func TestProxyHealth(t *testing.T) {
	t.Parallel()

	topLevelDelay := mihomo.Delay{Time: time.Date(2026, time.July, 31, 12, 0, 0, 0, time.UTC), Delay: 220}
	extraDelay := mihomo.Delay{Time: time.Date(2026, time.July, 31, 12, 1, 0, 0, time.UTC), Delay: 120}
	tests := map[string]struct {
		proxy   mihomo.Proxy
		testURL string
		want    mihomo.Delay
		ok      bool
	}{
		"uses matching extra bucket": {
			proxy: mihomo.Proxy{
				History: []mihomo.Delay{topLevelDelay},
				Extra: map[string]mihomo.ProxyHealth{
					"https://health.example/generate_204": {History: []mihomo.Delay{extraDelay}},
				},
			},
			testURL: "https://health.example/generate_204",
			want:    extraDelay,
			ok:      true,
		},
		"falls back when bucket is absent": {
			proxy:   mihomo.Proxy{History: []mihomo.Delay{topLevelDelay}},
			testURL: "https://health.example/generate_204",
			want:    topLevelDelay,
			ok:      true,
		},
		"falls back when bucket has no history": {
			proxy: mihomo.Proxy{
				History: []mihomo.Delay{topLevelDelay},
				Extra:   map[string]mihomo.ProxyHealth{"https://health.example/generate_204": {}},
			},
			testURL: "https://health.example/generate_204",
			want:    topLevelDelay,
			ok:      true,
		},
		"returns false without history": {
			proxy:   mihomo.Proxy{},
			testURL: "https://health.example/generate_204",
			ok:      false,
		},
	}
	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, ok := testCase.proxy.LastDelay(testCase.testURL)
			if diff := cmp.Diff(testCase.want, got); diff != "" {
				t.Errorf("LastDelay(%q) mismatch (-want +got):\n%s", testCase.testURL, diff)
			}
			if ok != testCase.ok {
				t.Errorf("LastDelay(%q) exists = %t, want %t", testCase.testURL, ok, testCase.ok)
			}
		})
	}
}
