package mihomo_test

import (
	"testing"

	"github.com/etnperlong/mihomo-exporter/internal/mihomo"
)

func TestIsOutboundNode(t *testing.T) {
	t.Parallel()

	tests := map[string]bool{
		"AnyTLS":       true,
		"Compatible":   false,
		"Direct":       false,
		"Dns":          false,
		"Fallback":     false,
		"GostRelay":    true,
		"Http":         true,
		"Hysteria":     true,
		"Hysteria2":    true,
		"LoadBalance":  false,
		"Masque":       true,
		"Mieru":        true,
		"OpenVPN":      true,
		"Pass":         false,
		"PassRule":     false,
		"Reject":       false,
		"RejectDrop":   false,
		"Relay":        false,
		"Rematch":      false,
		"Selector":     false,
		"ShadowQuic":   true,
		"Shadowsocks":  true,
		"ShadowsocksR": true,
		"Snell":        true,
		"Socks5":       true,
		"Ssh":          true,
		"Sudoku":       true,
		"Tailscale":    true,
		"Trojan":       true,
		"TrustTunnel":  true,
		"Tuic":         true,
		"URLTest":      false,
		"Unknown":      true,
		"Vless":        true,
		"Vmess":        true,
		"WireGuard":    true,
	}
	for proxyType, want := range tests {
		t.Run(proxyType, func(t *testing.T) {
			t.Parallel()

			if got := mihomo.IsOutboundNode(proxyType); got != want {
				t.Errorf("IsOutboundNode(%q) = %t, want %t", proxyType, got, want)
			}
		})
	}
}
