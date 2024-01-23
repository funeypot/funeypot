package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type Config struct {
	Ssh       Ssh       `yaml:"ssh"`
	Database  Database  `yaml:"database"`
	AbuseIpdb AbuseIpdb `yaml:"abuse_ipdb"`
}

type Ssh struct {
	Address string        `yaml:"address"`
	Delay   time.Duration `yaml:"delay"`
}

func (s Ssh) Validate() error {
	if s.Address == "" {
		return fmt.Errorf("ssh address is required")
	}
	if s.Delay < 0 {
		return fmt.Errorf("ssh delay cannot be negative")
	}
	return nil
}

type Database struct {
	Driver string `yaml:"driver"`
	Dsn    string `yaml:"dsn"`
}

func (d Database) Validate() error {
	if d.Driver == "" {
		return fmt.Errorf("database driver is required")
	}

	switch d.Driver {
	case "sqlite":
	default:
		return fmt.Errorf("unsupported database driver: %s", d.Driver)
	}

	if d.Dsn == "" {
		return fmt.Errorf("database dsn is required")
	}
	return nil
}

type AbuseIpdb struct {
	Enabled bool   `yaml:"enabled"`
	Key     string `json:"key"`
}

func (a AbuseIpdb) Validate() error {
	if a.Enabled {
		if a.Key == "" {
			return fmt.Errorf("abuse ipdb key is required when enabled")
		}
	}
	return nil
}

func Load(file string) (*Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("open file %q: %w", file, err)
	}
	defer f.Close()

	encoder := yaml.NewDecoder(f)
	ret := &Config{}
	if err := encoder.Decode(ret); err != nil {
		return nil, fmt.Errorf("decode yaml: %w", err)
	}
	return ret, nil
}

func (c *Config) Validate() error {
	if err := c.Ssh.Validate(); err != nil {
		return fmt.Errorf("ssh: %w", err)
	}
	if err := c.Database.Validate(); err != nil {
		return fmt.Errorf("database: %w", err)
	}
	if err := c.AbuseIpdb.Validate(); err != nil {
		return fmt.Errorf("abuse ipdb: %w", err)
	}
	return nil
}
