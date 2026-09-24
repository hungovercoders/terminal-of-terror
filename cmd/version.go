package cmd

import (
	"runtime/debug"
	"strings"
)

// version is set at release time by GoReleaser:
//
//	-ldflags "-X github.com/hungovercoders/terminal-of-terror/cmd.version=1.0.0"
var version string

// appVersion is the version to report. Release builds carry it in version.
// Other builds use Go's build info: `go install ...@v1.0.0` reports 1.0.0,
// and a local `go build` reports a pseudo-version naming the commit. "dev"
// is the last resort when neither is available.
func appVersion() string {
	if version != "" {
		return strings.TrimPrefix(version, "v")
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return strings.TrimPrefix(v, "v")
		}
	}
	return "dev"
}
