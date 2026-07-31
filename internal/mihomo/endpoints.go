package mihomo

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

// Version returns Mihomo version information.
func (c *Client) Version(ctx context.Context) (Version, error) {
	var version Version
	if err := c.get(ctx, []string{"version"}, nil, &version); err != nil {
		return Version{}, err
	}

	return version, nil
}

// Proxies returns all configured proxies indexed by name.
func (c *Client) Proxies(ctx context.Context) (map[string]Proxy, error) {
	var response proxiesResponse
	if err := c.get(ctx, []string{"proxies"}, nil, &response); err != nil {
		return nil, err
	}

	return response.Proxies, nil
}

// Providers returns all proxy providers indexed by name.
func (c *Client) Providers(ctx context.Context) (map[string]Provider, error) {
	var response providersResponse
	if err := c.get(ctx, []string{"providers", "proxies"}, nil, &response); err != nil {
		return nil, err
	}

	return response.Providers, nil
}

// ProviderHealthCheck runs Mihomo's side-effect-free health check for provider.
func (c *Client) ProviderHealthCheck(ctx context.Context, provider string) error {
	return c.get(ctx, []string{"providers", "proxies", provider, "healthcheck"}, nil, &struct{}{})
}

// ProxyDelay measures a proxy using testURL and returns the delay in milliseconds.
func (c *Client) ProxyDelay(ctx context.Context, name, testURL string, timeout time.Duration) (uint16, error) {
	query := url.Values{
		"timeout": []string{strconv.FormatInt(timeout.Milliseconds(), 10)},
		"url":     []string{testURL},
	}
	var response delayResponse
	if err := c.get(ctx, []string{"proxies", name, "delay"}, query, &response); err != nil {
		return 0, err
	}

	return response.Delay, nil
}

// Connections returns aggregate traffic counters and active connections.
func (c *Client) Connections(ctx context.Context) (Connections, error) {
	var connections Connections
	if err := c.get(ctx, []string{"connections"}, nil, &connections); err != nil {
		return Connections{}, err
	}

	return connections, nil
}
