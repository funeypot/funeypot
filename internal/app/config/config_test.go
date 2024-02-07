// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package config

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name         string
		modifyConfig func(cfg *Config)
		wantErr      assert.ErrorAssertionFunc
	}{
		{
			name:         "default",
			modifyConfig: nil,
			wantErr:      assert.NoError,
		},
		{
			name: "empty log level",
			modifyConfig: func(cfg *Config) {
				cfg.Log.Level = ""
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid log level",
			modifyConfig: func(cfg *Config) {
				cfg.Log.Level = "test"
			},
			wantErr: assert.Error,
		},
		{
			name: "empty ssh address",
			modifyConfig: func(cfg *Config) {
				cfg.Ssh.Address = ""
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid ssh delay",
			modifyConfig: func(cfg *Config) {
				cfg.Ssh.Delay = -1
			},
			wantErr: assert.Error,
		},
		{
			name: "empty http address",
			modifyConfig: func(cfg *Config) {
				cfg.Http.Enabled = true
				cfg.Http.Address = ""
			},
			wantErr: assert.Error,
		},
		{
			name: "valid http",
			modifyConfig: func(cfg *Config) {
				cfg.Http.Enabled = true
				cfg.Http.Address = ":8080"
			},
			wantErr: assert.NoError,
		},
		{
			name: "empty ftp address",
			modifyConfig: func(cfg *Config) {
				cfg.Ftp.Enabled = true
				cfg.Ftp.Address = ""
			},
			wantErr: assert.Error,
		},
		{
			name: "valid ftp",
			modifyConfig: func(cfg *Config) {
				cfg.Ftp.Enabled = true
				cfg.Ftp.Address = ":2121"
			},
			wantErr: assert.NoError,
		},
		{
			name: "empty database driver",
			modifyConfig: func(cfg *Config) {
				cfg.Database.Driver = ""
			},
			wantErr: assert.Error,
		},
		{
			name: "empty database dsn",
			modifyConfig: func(cfg *Config) {
				cfg.Database.Dsn = ""
			},
			wantErr: assert.Error,
		},
		{
			name: "empty database dsn",
			modifyConfig: func(cfg *Config) {
				cfg.Database.Dsn = ""
			},
			wantErr: assert.Error,
		},
		{
			name: "empty dashboard username",
			modifyConfig: func(cfg *Config) {
				cfg.Dashboard.Enabled = true
				cfg.Dashboard.Username = ""
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid dashboard username",
			modifyConfig: func(cfg *Config) {
				cfg.Dashboard.Enabled = true
				cfg.Dashboard.Username = "username"
				cfg.Dashboard.Password = "pwd"
			},
			wantErr: assert.Error,
		},
		{
			name: "dashboard without http",
			modifyConfig: func(cfg *Config) {
				cfg.Dashboard.Enabled = true
				cfg.Dashboard.Username = "username"
				cfg.Dashboard.Password = "password"
			},
			wantErr: assert.Error,
		},
		{
			name: "valid dashboard",
			modifyConfig: func(cfg *Config) {
				cfg.Http.Enabled = true
				cfg.Dashboard.Enabled = true
				cfg.Dashboard.Username = "username"
				cfg.Dashboard.Password = "password"
			},
			wantErr: assert.NoError,
		},
		{
			name: "empty abuseipdb key",
			modifyConfig: func(cfg *Config) {
				cfg.Abuseipdb.Enabled = true
				cfg.Abuseipdb.Key = ""
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid abuseipdb interval",
			modifyConfig: func(cfg *Config) {
				cfg.Abuseipdb.Enabled = true
				cfg.Abuseipdb.Key = "key"
				cfg.Abuseipdb.Interval = time.Minute
			},
			wantErr: assert.Error,
		},
		{
			name: "valid abuseipdb",
			modifyConfig: func(cfg *Config) {
				cfg.Abuseipdb.Enabled = true
				cfg.Abuseipdb.Key = "key"
				cfg.Abuseipdb.Interval = 15 * time.Minute
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Load(filepath.Join(t.TempDir(), "funeypot.yaml"), true)
			require.NoError(t, err)
			if tt.modifyConfig != nil {
				tt.modifyConfig(cfg)
			}
			tt.wantErr(t, cfg.Validate())
		})
	}
}

func TestLoad(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		_, err := Load(filepath.Join(t.TempDir(), "funeypot.yaml"), false)
		assert.Error(t, err)
	})
}
