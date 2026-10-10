package espn

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFallbackDNS(t *testing.T) {
	prefix := t.TempDir()
	os.MkdirAll(filepath.Join(prefix, "etc"), 0o755)
	os.WriteFile(filepath.Join(prefix, "etc", "resolv.conf"), []byte("# termux\nnameserver 8.8.8.8\nnameserver 8.8.4.4\n"), 0o644)

	t.Setenv("PREFIX", prefix)
	if got := fallbackDNS(); got != "8.8.8.8:53" {
		t.Errorf("got %q", got)
	}

	t.Setenv("PREFIX", "")
	if got := fallbackDNS(); got != "1.1.1.1:53" {
		t.Errorf("got %q", got)
	}
}
