package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetupStepForms(t *testing.T) {
	c := readTest(t, "copy:\n  - .env\nsetup:\n  - npm install\n  - run: npm --prefix web install\n    dir: web\n")
	if len(c.Setup) != 2 {
		t.Fatalf("got %d steps: %+v", len(c.Setup), c.Setup)
	}
	if c.Setup[0].Run != "npm install" || c.Setup[0].Dir != "" {
		t.Errorf("string form: %+v", c.Setup[0])
	}
	if c.Setup[1].Run != "npm --prefix web install" || c.Setup[1].Dir != "web" {
		t.Errorf("map form: %+v", c.Setup[1])
	}
}

func TestTeardownStepForms(t *testing.T) {
	c := readTest(t, "teardown:\n  - ./scripts/down.sh\n  - run: docker compose down\n    dir: web\n")
	if len(c.Teardown) != 2 {
		t.Fatalf("got %d steps: %+v", len(c.Teardown), c.Teardown)
	}
	if c.Teardown[0].Run != "./scripts/down.sh" || c.Teardown[0].Dir != "" {
		t.Errorf("string form: %+v", c.Teardown[0])
	}
	if c.Teardown[1].Run != "docker compose down" || c.Teardown[1].Dir != "web" {
		t.Errorf("map form: %+v", c.Teardown[1])
	}
}

func TestLoadMergesFieldsIndependently(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	global := "copy:\n  - .env\nsetup:\n  - run: echo global\nteardown:\n  - run: echo global-down\n"
	if err := os.MkdirAll(filepath.Join(home, ".config", "wtm"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".config", "wtm", "config.yaml"), []byte(global), 0o644); err != nil {
		t.Fatal(err)
	}
	// Repo file sets copy only: global setup and teardown must survive.
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, ".wtm.yaml"), []byte("copy:\n  - web/.env\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := Load(repo)
	if len(cfg.Copy) != 1 || cfg.Copy[0] != "web/.env" {
		t.Errorf("copy: %+v", cfg.Copy)
	}
	if len(cfg.Setup) != 1 || cfg.Setup[0].Run != "echo global" {
		t.Errorf("setup should fall back to global: %+v", cfg.Setup)
	}
	if len(cfg.Teardown) != 1 || cfg.Teardown[0].Run != "echo global-down" {
		t.Errorf("teardown should fall back to global: %+v", cfg.Teardown)
	}
}

func TestLoadRepoTeardownReplacesGlobal(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	global := "teardown:\n  - run: echo global-down\n"
	if err := os.MkdirAll(filepath.Join(home, ".config", "wtm"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".config", "wtm", "config.yaml"), []byte(global), 0o644); err != nil {
		t.Fatal(err)
	}
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, ".wtm.yaml"), []byte("teardown:\n  - run: echo local-down\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := Load(repo)
	if len(cfg.Teardown) != 1 || cfg.Teardown[0].Run != "echo local-down" {
		t.Errorf("repo teardown should replace global: %+v", cfg.Teardown)
	}
}

func readTest(t *testing.T, body string) *Config {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, ".wtm.yaml")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	c := read(p)
	if c == nil {
		t.Fatal("read returned nil")
	}
	return c
}
