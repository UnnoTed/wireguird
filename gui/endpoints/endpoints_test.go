package endpoints

import (
	"testing"
)

func TestCategorization(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		want     string
	}{
		{"Empty string", "", Unknown},
		{"Just a colon", ":", Unknown},
		{"Empty brackets", "[]", Unknown},
		{"Empty brackets with port", "[]:51820", Unknown},

		{"IPV4 with port", "192.168.1.1:51820", IPV4},
		{"IPV4 without port", "10.0.0.1", IPV4},

		{"IPV6 with brackets & port", "[2001:db8::1]:51820", IPV6},
		{"IPV6 with brackets, no port", "[2001:db8::1]", IPV6},
		{"IPV6 no brackets, no port", "2001:db8::1", IPV6},

		{"Valid FQDN with port", "vpn.example.com:51820", FQDN},
		{"Valid FQDN without port", "wg.my-domain.net", FQDN},

		{"FQDN with spaces", "my vpn.com:51820", Unknown},
		{"URL with scheme", "https://vpn.example.com", Unknown},
		{"Malformed IP in brackets", "[192.168.1.1]", Unknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Categorize(tt.endpoint)
			if got != tt.want {
				t.Errorf("\nCategorize(%q)\n  got:  %s\n  want: %s", tt.endpoint, got, tt.want)
			}
		})
	}
}
