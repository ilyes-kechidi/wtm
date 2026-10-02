package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// SetupStep is one post-create (or teardown) command. Dir is relative to the
// worktree root (empty means the root). The command runs via `sh -c`.
type SetupStep struct {
	Run string `yaml:"run"`
	Dir string `yaml:"dir"`
}

// UnmarshalYAML accepts a plain string ("npm install") or a {run, dir} map.
func (s *SetupStep) UnmarshalYAML(value *yaml.Node) error {
	var cmd string
	if err := value.Decode(&cmd); err == nil {
		s.Run = cmd
		return nil
	}
	type raw SetupStep
	var r raw
	if err := value.Decode(&r); err != nil {
		return err
	}
	*s = SetupStep(r)
	if s.Run == "" {
		return &yaml.TypeError{Errors: []string{"setup step needs a `run` command"}}
	}
	return nil
}

// Config holds the copy list and post-create / teardown steps. Paths/globs are
// relative to repo root.
type Config struct {
	Copy     []string    `yaml:"copy"`
	Setup    []SetupStep `yaml:"setup"`
	Teardown []SetupStep `yaml:"teardown"`
}

// Defaults covers the common single-.env case.
func Defaults() Config {
	return Config{Copy: []string{".env", ".env.local"}}
}

// Load merges global (~/.config/wtm/config.yaml) with repo-local .wtm.yaml
// field by field: a non-empty repo-local list replaces the global one.
func Load(repoRoot string) Config {
	cfg := Defaults()
	merge := func(c *Config) {
		if len(c.Copy) > 0 {
			cfg.Copy = c.Copy
		}
		if len(c.Setup) > 0 {
			cfg.Setup = c.Setup
		}
		if len(c.Teardown) > 0 {
			cfg.Teardown = c.Teardown
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		if g := read(filepath.Join(home, ".config", "wtm", "config.yaml")); g != nil {
			merge(g)
		}
	}
	if l := read(filepath.Join(repoRoot, ".wtm.yaml")); l != nil {
		merge(l)
	}
	return cfg
}

func read(path string) *Config {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil
	}
	return &c
}
