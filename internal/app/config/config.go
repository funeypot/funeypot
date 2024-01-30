package config

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Log       Log       `yaml:"log"`
	Ssh       Ssh       `yaml:"ssh"`
	Http      Http      `yaml:"http"`
	Ftp       Ftp       `yaml:"ftp"`
	Database  Database  `yaml:"database"`
	Dashboard Dashboard `yaml:"dashboard"`
	Abuseipdb Abuseipdb `yaml:"abuseipdb"`
}

type Log struct {
	Level string `yaml:"level"`
}

func (l Log) Validate() error {
	if l.Level == "" {
		return fmt.Errorf("level is required")
	}
	if _, err := zapcore.ParseLevel(l.Level); err != nil {
		return fmt.Errorf("invalid level %q: %w", l.Level, err)
	}
	return nil
}

type Ssh struct {
	Address string        `yaml:"address"`
	Delay   time.Duration `yaml:"delay"`
	KeySeed string        `yaml:"key_seed"`
}

func (s Ssh) Validate() error {
	if s.Address == "" {
		return fmt.Errorf("address is required")
	}
	if s.Delay < 0 {
		return fmt.Errorf("delay cannot be negative")
	}
	return nil
}

type Http struct {
	Enabled bool   `yaml:"enabled"`
	Address string `yaml:"address"`
}

func (h Http) Validate() error {
	if !h.Enabled {
		return nil
	}
	if h.Address == "" {
		return fmt.Errorf("address is required")
	}
	return nil
}

type Ftp struct {
	Enabled bool   `yaml:"enabled"`
	Address string `yaml:"address"`
}

func (f Ftp) Validate() error {
	if !f.Enabled {
		return nil
	}
	if f.Address == "" {
		return fmt.Errorf("address is required")
	}
	return nil
}

type Database struct {
	Driver string `yaml:"driver"`
	Dsn    string `yaml:"dsn"`
}

func (d Database) Validate() error {
	if d.Driver == "" {
		return fmt.Errorf("driver is required")
	}

	if d.Dsn == "" {
		return fmt.Errorf("dsn is required")
	}
	return nil
}

type Dashboard struct {
	Enabled  bool   `yaml:"enabled"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func (d Dashboard) Validate() error {
	if !d.Enabled {
		return nil
	}
	if d.Username == "" {
		return fmt.Errorf("username is required")
	}
	if len(d.Password) < 8 {
		return fmt.Errorf("password is required and must be at least 8 characters")
	}
	return nil
}

type Abuseipdb struct {
	Enabled  bool          `yaml:"enabled"`
	Key      string        `yaml:"key"`
	Interval time.Duration `yaml:"interval"`
}

func (a Abuseipdb) Validate() error {
	if !a.Enabled {
		return nil
	}
	if a.Key == "" {
		return fmt.Errorf("key is required when enabled")
	}
	if a.Interval < 15*time.Minute {
		return fmt.Errorf("interval must be at least 15 minutes")
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

	if err := ret.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return ret, nil
}

func (c *Config) Validate() error {
	if err := c.Log.Validate(); err != nil {
		return fmt.Errorf("log: %w", err)
	}
	if err := c.Ssh.Validate(); err != nil {
		return fmt.Errorf("ssh: %w", err)
	}
	if err := c.Http.Validate(); err != nil {
		return fmt.Errorf("http: %w", err)
	}
	if err := c.Database.Validate(); err != nil {
		return fmt.Errorf("database: %w", err)
	}
	if err := c.Dashboard.Validate(); err != nil {
		return fmt.Errorf("dashboard: %w", err)
	}
	if err := c.Abuseipdb.Validate(); err != nil {
		return fmt.Errorf("abuse ipdb: %w", err)
	}

	if !c.Http.Enabled && c.Dashboard.Enabled {
		return fmt.Errorf("http.enabled must be true when dashboard.enabled is true")
	}

	return nil
}
