package mihomo

import "time"

// Version contains Mihomo build information.
type Version struct {
	Meta    bool   `json:"meta"`
	Version string `json:"version"`
}

// Delay records a proxy delay result at a point in time.
type Delay struct {
	Time  time.Time `json:"time"`
	Delay uint16    `json:"delay"`
}

// ProxyHealth contains availability and delay history for one test URL.
type ProxyHealth struct {
	Alive   bool    `json:"alive"`
	History []Delay `json:"history"`
}

// Proxy represents a Mihomo proxy or proxy group.
type Proxy struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Alive        bool                   `json:"alive"`
	History      []Delay                `json:"history"`
	Extra        map[string]ProxyHealth `json:"extra"`
	UDP          bool                   `json:"udp"`
	UOT          bool                   `json:"uot"`
	XUDP         bool                   `json:"xudp"`
	TFO          bool                   `json:"tfo"`
	MPTCP        bool                   `json:"mptcp"`
	SMUX         bool                   `json:"smux"`
	Interface    string                 `json:"interface"`
	RoutingMark  int                    `json:"routing-mark"`
	ProviderName string                 `json:"provider-name"`
	DialerProxy  string                 `json:"dialer-proxy"`
}

// Health returns health data for testURL, falling back to the proxy's top-level data.
func (p Proxy) Health(testURL string) ProxyHealth {
	if health, ok := p.Extra[testURL]; ok && len(health.History) > 0 {
		return health
	}

	return ProxyHealth{Alive: p.Alive, History: p.History}
}

// LastDelay returns the most recent delay for testURL and whether it exists.
func (p Proxy) LastDelay(testURL string) (Delay, bool) {
	history := p.Health(testURL).History
	if len(history) == 0 {
		return Delay{}, false
	}

	return history[len(history)-1], true
}

// SubscriptionInfo contains usage information returned for a subscription provider.
type SubscriptionInfo struct {
	Upload   int64 `json:"upload"`
	Download int64 `json:"download"`
	Total    int64 `json:"total"`
	Expire   int64 `json:"expire"`
}

// Provider represents a Mihomo proxy provider.
type Provider struct {
	Name             string            `json:"name"`
	Type             string            `json:"type"`
	VehicleType      string            `json:"vehicleType"`
	Proxies          []Proxy           `json:"proxies"`
	TestURL          string            `json:"testUrl"`
	ExpectedStatus   string            `json:"expectedStatus"`
	UpdatedAt        time.Time         `json:"updatedAt"`
	SubscriptionInfo *SubscriptionInfo `json:"subscriptionInfo"`
}

// ConnectionMetadata contains metadata describing an active connection.
type ConnectionMetadata struct {
	Network         string `json:"network"`
	Type            string `json:"type"`
	SourceIP        string `json:"sourceIP"`
	DestinationIP   string `json:"destinationIP"`
	SourcePort      string `json:"sourcePort"`
	DestinationPort string `json:"destinationPort"`
	Host            string `json:"host"`
	DNSMode         string `json:"dnsMode"`
	ProcessPath     string `json:"processPath"`
}

// Connection represents an active Mihomo connection.
type Connection struct {
	ID          string             `json:"id"`
	Metadata    ConnectionMetadata `json:"metadata"`
	Upload      int64              `json:"upload"`
	Download    int64              `json:"download"`
	Start       time.Time          `json:"start"`
	Chains      []string           `json:"chains"`
	Rule        string             `json:"rule"`
	RulePayload string             `json:"rulePayload"`
}

// Connections contains aggregate traffic counters and active connections.
type Connections struct {
	DownloadTotal int64        `json:"downloadTotal"`
	UploadTotal   int64        `json:"uploadTotal"`
	Connections   []Connection `json:"connections"`
}

type proxiesResponse struct {
	Proxies map[string]Proxy `json:"proxies"`
}

type providersResponse struct {
	Providers map[string]Provider `json:"providers"`
}

type delayResponse struct {
	Delay uint16 `json:"delay"`
}
