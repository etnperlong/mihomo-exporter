package collector

import "github.com/prometheus/client_golang/prometheus"

type descriptors struct {
	proxyDelay             *prometheus.Desc
	proxyAvailable         *prometheus.Desc
	proxyLastProbe         *prometheus.Desc
	providerInfo           *prometheus.Desc
	providerProxyCount     *prometheus.Desc
	providerAvailableCount *prometheus.Desc
	providerUpdated        *prometheus.Desc
	providerHealthDuration *prometheus.Desc
	providerHealthSuccess  *prometheus.Desc
	up                     *prometheus.Desc
	versionInfo            *prometheus.Desc
	buildInfo              *prometheus.Desc
	scrapeErrors           *prometheus.Desc
	probeDuration          *prometheus.Desc
	lastSuccess            *prometheus.Desc
	probeFallback          *prometheus.Desc
	unknownProxyType       *prometheus.Desc
}

func newDescriptors() descriptors {
	return descriptors{
		proxyDelay:             prometheus.NewDesc("mihomo_proxy_delay_milliseconds", "Latest successful proxy delay in milliseconds.", []string{"name", "type", "provider"}, nil),
		proxyAvailable:         prometheus.NewDesc("mihomo_proxy_available", "Whether the proxy is available.", []string{"name", "type", "provider"}, nil),
		proxyLastProbe:         prometheus.NewDesc("mihomo_proxy_last_probe_timestamp_seconds", "Unix timestamp of the latest proxy probe.", []string{"name", "type", "provider"}, nil),
		providerInfo:           prometheus.NewDesc("mihomo_provider_info", "Provider metadata.", []string{"provider", "vehicle_type", "type", "test_url"}, nil),
		providerProxyCount:     prometheus.NewDesc("mihomo_provider_proxy_count", "Number of proxies in the provider.", []string{"provider"}, nil),
		providerAvailableCount: prometheus.NewDesc("mihomo_provider_proxy_available_count", "Number of available proxies in the provider.", []string{"provider"}, nil),
		providerUpdated:        prometheus.NewDesc("mihomo_provider_updated_timestamp_seconds", "Unix timestamp of the provider update.", []string{"provider"}, nil),
		providerHealthDuration: prometheus.NewDesc("mihomo_provider_healthcheck_duration_seconds", "Duration of the latest provider health check.", []string{"provider"}, nil),
		providerHealthSuccess:  prometheus.NewDesc("mihomo_provider_healthcheck_success", "Whether the latest provider health check succeeded.", []string{"provider"}, nil),
		up:                     prometheus.NewDesc("mihomo_up", "Whether the latest fast probe connected to Mihomo.", nil, nil),
		versionInfo:            prometheus.NewDesc("mihomo_version_info", "Mihomo version information.", []string{"version", "meta"}, nil),
		buildInfo:              prometheus.NewDesc("mihomo_exporter_build_info", "Mihomo exporter build information.", []string{"version", "revision", "goversion"}, nil),
		scrapeErrors:           prometheus.NewDesc("mihomo_exporter_scrape_errors_total", "Total Mihomo scrape errors.", []string{"collector"}, nil),
		probeDuration:          prometheus.NewDesc("mihomo_exporter_probe_duration_seconds", "Duration of the latest probe loop.", []string{"loop"}, nil),
		lastSuccess:            prometheus.NewDesc("mihomo_exporter_last_success_timestamp_seconds", "Unix timestamp of the latest successful probe loop.", []string{"loop"}, nil),
		probeFallback:          prometheus.NewDesc("mihomo_exporter_probe_fallback_total", "Total proxies measured by fallback probes.", nil, nil),
		unknownProxyType:       prometheus.NewDesc("mihomo_exporter_unknown_proxy_type_total", "Total unknown proxy types encountered.", []string{"type"}, nil),
	}
}
