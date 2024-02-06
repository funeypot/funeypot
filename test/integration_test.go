// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/wolfogre/funeypot/internal/app/config"
	"github.com/wolfogre/funeypot/internal/app/inject"
	"github.com/wolfogre/funeypot/internal/pkg/logs"
)

func PrepareServers(t *testing.T, modifyConfig func(cfg *config.Config)) func() {
	ctx, cancel := context.WithCancel(context.Background())

	cfg, err := config.Load(filepath.Join(t.TempDir(), "funeypot.yaml"), true)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	cfg.Ssh.Address = ":2222"
	cfg.Log.Level = "error"
	cfg.Database.Dsn = filepath.Join(t.TempDir(), "funeypot.db")

	if modifyConfig != nil {
		modifyConfig(cfg)
	}

	logs.SetLevel(cfg.Log.Level)

	entrypoint, err := inject.NewEntrypoint(ctx, cfg)
	if err != nil {
		cancel()
		t.Fatalf("new entrypoint: %v", err)
	}

	entrypoint.Startup(ctx, cancel)
	time.Sleep(time.Second)

	return func() {
		cancel()
		entrypoint.Shutdown(ctx)
	}
}
