package gui

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestValidNewTunnelName(t *testing.T) {
	valid := []string{"vpn1", "us-east", "my.tunnel", "a", "abcdefghijklmno", "WG_0", "a-b.c_d"}
	for _, n := range valid {
		if !validNewTunnelName(n) {
			t.Errorf("validNewTunnelName(%q) = false, want true", n)
		}
	}
	invalid := []string{
		"",                 // empty
		"../../tmp/evil",   // traversal
		"..",               // parent
		".",                // self
		"-evil",            // option injection
		".hidden",          // dotfile
		"a/b",              // separator
		"abcdefghijklmnop", // 16 chars, over IFNAMSIZ-1
		"name with spaces", // whitespace
		"semi;colon",       // shell metachar
		"back`tick",        // shell metachar
		"dollar$(x)",       // shell metachar
	}
	for _, n := range invalid {
		if validNewTunnelName(n) {
			t.Errorf("validNewTunnelName(%q) = true, want false", n)
		}
	}
	if !validNewTunnelName("fifteen-chars-x") { // exactly 15 chars
		t.Error("validNewTunnelName(15-char name) = false, want true")
	}
}

func TestTunnelConfPathStaysInsideTunnelDir(t *testing.T) {
	for _, n := range []string{"vpn1", "us-east", "my.tunnel", "fifteen-chars-x"} {
		p, err := tunnelConfPath(n, ".conf")
		if err != nil {
			t.Fatalf("tunnelConfPath(%q) unexpected error: %v", n, err)
		}
		rel, err := filepath.Rel(TunnelsPath, p)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			t.Errorf("tunnelConfPath(%q) = %q escapes %q", n, p, TunnelsPath)
		}
	}
	for _, n := range []string{"", "..", ".", "-x", "../../tmp/evil", "a/b", "toolongname123456"} {
		if _, err := tunnelConfPath(n, ".conf"); err == nil {
			t.Errorf("tunnelConfPath(%q) = nil error, want rejection", n)
		}
	}
}

func TestSafeExistingTunnelName(t *testing.T) {
	// Grandfathered: dots and pre-existing odd names keep working.
	for _, n := range []string{"vpn1", "my.tunnel", "old name with spaces", "abcdefghijklmnop"} {
		if !safeExistingTunnelName(n) {
			t.Errorf("safeExistingTunnelName(%q) = false, want true (grandfathered)", n)
		}
	}
	for _, n := range []string{"", ".", "..", "-inject", "sub/dir", "\x00null"} {
		if safeExistingTunnelName(n) {
			t.Errorf("safeExistingTunnelName(%q) = true, want false", n)
		}
	}
}
