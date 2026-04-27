package endpoints

import (
	"net"
	"strings"
	"unicode"
)

const (
	Unknown = "Unknown"
	IPV4    = "IPV4"
	IPV6    = "IPV6"
	FQDN    = "FQDN"
)

func Categorize(endpoint string) string {
	// no empty or urls
	if endpoint == "" || strings.Contains(endpoint, "://") {
		return Unknown
	}

	host, _, err := net.SplitHostPort(endpoint)
	if err != nil {
		// when no port is present, use the it as the host.
		host = endpoint
	}

	// IPV6 with brackets
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		inner := host[1 : len(host)-1]

		ip := net.ParseIP(inner)
		if ip != nil && ip.To4() == nil {
			return IPV6
		}

		// invalid [IPV6]
		return Unknown
	}

	if ip := net.ParseIP(host); ip != nil {
		if ip.To4() != nil {
			return IPV4
		}

		return IPV6
	}

	// FQDNs have no web url chars like ?, #, @...
	if host == "" || strings.ContainsAny(host, " /?#@:[]") {
		return Unknown
	}

	// FQDNs must be printable and contain no spaces
	// printable means that strange chars like chinese should work
	for _, r := range host {
		if !unicode.IsPrint(r) || unicode.IsSpace(r) {
			return Unknown
		}
	}

	return FQDN
}
