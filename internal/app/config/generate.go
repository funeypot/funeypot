package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

var (
	//go:embed config.yaml
	defaultConfigYaml []byte
)

func Generate(file string) error {
	dir := filepath.Dir(file)
	if dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create directory %q: %w", dir, err)
		}
	}
	if err := os.WriteFile(file, defaultConfigYaml, 0o644); err != nil {
		return fmt.Errorf("write file %q: %w", file, err)
	}
	return nil
}
