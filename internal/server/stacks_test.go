package server

import (
	"os"
	"path/filepath"
	"strings"
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

func TestMergeComposeEnvDoesNotLetControllerEnvOverrideStackEnv(t *testing.T) {
	t.Setenv("POSTGRES_DB", "controller-value")
	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte("POSTGRES_DB=stack-value\n"), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	merged := mergeComposeEnv(map[string]string{"DOCKER_HOST": "unix:///custom.sock"}, envPath)
	values := make(map[string]string, len(merged))
	for _, item := range merged {
		key, value, found := strings.Cut(item, "=")
		if found {
			values[key] = value
		}
	}

	if _, exists := values["POSTGRES_DB"]; exists {
		t.Fatal("expected POSTGRES_DB to be resolved from --env-file, not controller environment")
	}
	if got := values["DOCKER_HOST"]; got != "unix:///custom.sock" {
		t.Fatalf("expected Docker host override, got %q", got)
	}
}

func TestRunComposeFromPayloadPlacesEnvNextToCompose(t *testing.T) {
	binDir := t.TempDir()
	dockerPath := filepath.Join(binDir, "docker")
	script := `#!/bin/sh
set -eu
test -f "$PWD/docker-compose.yml"
test -f "$PWD/.env"
grep -q '^POSTGRES_DB=stack-value$' "$PWD/.env"
case " $* " in
  *" --env-file $PWD/.env "*) ;;
  *) echo "missing colocated --env-file argument: $*" >&2; exit 1 ;;
esac
`
	if err := os.WriteFile(dockerPath, []byte(script), 0o700); err != nil {
		t.Fatalf("write fake docker: %v", err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	err := runComposeFromPayload(stackActionRequest{
		Name:       "database",
		Action:     "up",
		ComposeYml: "services:\n  db:\n    image: postgres\n    env_file:\n      - .env\n",
		Env:        "POSTGRES_DB=stack-value\n",
		UseEnv:     true,
	}, "")
	if err != nil {
		t.Fatalf("run compose payload: %v", err)
	}
}

func TestRunComposeConfigValidatesInterpolatedConfiguration(t *testing.T) {
	binDir := t.TempDir()
	dockerPath := filepath.Join(binDir, "docker")
	script := `#!/bin/sh
set -eu
test -f "$PWD/.env"
case " $* " in
  *" config --quiet "*) ;;
  *) echo "expected config --quiet: $*" >&2; exit 1 ;;
esac
case " $* " in
  *"--no-interpolate"*) echo "interpolation unexpectedly disabled" >&2; exit 1 ;;
esac
`
	if err := os.WriteFile(dockerPath, []byte(script), 0o700); err != nil {
		t.Fatalf("write fake docker: %v", err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	err := runComposeConfigFromPayload(stackValidateRequest{
		ComposeYml: "services:\n  db:\n    image: postgres:${POSTGRES_TAG:-latest}\n",
		Env:        "POSTGRES_TAG=16\n",
		UseEnv:     true,
	})
	if err != nil {
		t.Fatalf("validate compose payload: %v", err)
	}
}
