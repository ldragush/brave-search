package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const configFileName = ".brave-search.yaml"

type Config struct {
	APIKeys []string `yaml:"api_keys"`
}

func EnsureConfigFile() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(exe)
	cfgPath := filepath.Join(dir, configFileName)

	_, err = os.Stat(cfgPath)
	if err == nil {
		return cfgPath, nil
	}
	if !os.IsNotExist(err) {
		return "", err
	}

	// Create new config with the required structure
	cfg := Config{APIKeys: []string{}}
	b, err := yaml.Marshal(&cfg)
	if err != nil {
		return "", err
	}

	// Strict permissions
	if err := os.WriteFile(cfgPath, b, 0o600); err != nil {
		return "", err
	}

	return cfgPath, nil
}

func LoadConfig(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("invalid yaml: %w", err)
	}
	// Normalize: trim spaces, drop empties
	out := make([]string, 0, len(cfg.APIKeys))
	for _, k := range cfg.APIKeys {
		k = trim(k)
		if k != "" {
			out = append(out, k)
		}
	}
	cfg.APIKeys = out
	return cfg, nil
}

func trim(s string) string {
	// no unicode tricks; api keys are ascii tokens
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t' || s[0] == '\n' || s[0] == '\r') {
		s = s[1:]
	}
	for len(s) > 0 {
		last := s[len(s)-1]
		if last == ' ' || last == '\t' || last == '\n' || last == '\r' {
			s = s[:len(s)-1]
			continue
		}
		break
	}
	return s
}
