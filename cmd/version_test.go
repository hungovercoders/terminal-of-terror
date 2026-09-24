package cmd

import "testing"

func TestAppVersion(t *testing.T) {
	saved := version
	t.Cleanup(func() { version = saved })

	version = "v1.2.3"
	if got := appVersion(); got != "1.2.3" {
		t.Errorf("release build: got %q, want 1.2.3", got)
	}
	version = ""
	// Test binaries have no module version, so this is a development build.
	if got := appVersion(); got != "dev" {
		t.Errorf("local build: got %q, want dev", got)
	}
}
