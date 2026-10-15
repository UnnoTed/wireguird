package gui

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// newTunnelNamePattern constrains names the user can assign to a tunnel
// (editor rename flow). Rules:
//   - 1-15 chars (kernel IFNAMSIZ limit, see issue #65)
//   - letters, digits, '_' , '-' and '.' only — no '/' so the name can
//     never escape TunnelsPath via filepath.Join
//   - must not start with '-' or '.' (no option-injection into wg-quick,
//     no '.' / '..' basenames)
var newTunnelNamePattern = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,14}$`)

// validNewTunnelName reports whether name may be assigned to a tunnel.
func validNewTunnelName(name string) bool {
	return newTunnelNamePattern.MatchString(name)
}

// safeExistingTunnelName reports whether an already-on-disk tunnel name is
// safe to pass to wg-quick. Existing configs are grandfathered in (dots and
// longer names keep working); only values that could escape the tunnel
// directory or inject options are rejected. An empty name is rejected too.
func safeExistingTunnelName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}
	if strings.Contains(name, "/") || strings.ContainsRune(name, 0) {
		return false
	}
	if strings.HasPrefix(name, "-") {
		return false
	}
	return true
}

// tunnelConfPath resolves a user-supplied tunnel name to its config path,
// guaranteeing the result stays inside TunnelsPath. New names must pass the
// strict allowlist; anything else is rejected before any rename/write.
// ext is the caller-computed extension suffix ("" or ".conf").
func tunnelConfPath(name, ext string) (string, error) {
	if !validNewTunnelName(name) {
		return "", fmt.Errorf("invalid tunnel name %q: use 1-15 chars of [A-Za-z0-9_.-], not starting with '-' or '.'", name)
	}
	p := filepath.Join(TunnelsPath, name+ext)
	rel, err := filepath.Rel(TunnelsPath, p)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("tunnel name %q escapes the tunnel directory", name)
	}
	return p, nil
}
