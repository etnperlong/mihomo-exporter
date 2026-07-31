package mihomo

// groupTypes contains Mihomo group and built-in adapter types, not outbound nodes.
var groupTypes = map[string]struct{}{
	"Compatible": {}, "Direct": {}, "Dns": {}, "Fallback": {}, "LoadBalance": {},
	"Pass": {}, "PassRule": {}, "Reject": {}, "RejectDrop": {}, "Relay": {},
	"Rematch": {}, "Selector": {}, "URLTest": {},
}

// IsOutboundNode reports whether proxyType is an outbound protocol rather than a group or built-in adapter.
func IsOutboundNode(proxyType string) bool {
	_, isGroup := groupTypes[proxyType]

	return !isGroup
}
