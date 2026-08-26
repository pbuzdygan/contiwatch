package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWritePrivateFileAtomicallyKeepsRestrictedPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "docker-compose.yml")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	if err := writePrivateFileAtomically(path, []byte("services: {}\n")); err != nil {
		t.Fatalf("write private file: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat private file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("expected permissions 0600, got %04o", got)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read private file: %v", err)
	}
	if string(content) != "services: {}\n" {
		t.Fatalf("unexpected content %q", content)
	}
}
