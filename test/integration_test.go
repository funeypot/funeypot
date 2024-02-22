// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package test

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/funeypot/funeypot/internal/app/config"
	"github.com/funeypot/funeypot/internal/app/entry"
	"github.com/funeypot/funeypot/internal/pkg/logs"
)

func PrepareServers(t *testing.T, modifyConfig func(cfg *config.Config)) func() {
	ctx, cancel := context.WithCancel(context.Background())

	cfg, err := config.Load(filepath.Join(t.TempDir(), "funeypot.yaml"), true)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	// adjust config for testing
	cfg.Ssh.Address = ":2222"
	cfg.Http.Address = ":8080"
	cfg.Ftp.Address = ":2121"
	cfg.Log.Level = "error"
	cfg.Database.Dsn = filepath.Join(t.TempDir(), "funeypot.db")

	if modifyConfig != nil {
		modifyConfig(cfg)
	}

	logs.SetLevel(cfg.Log.Level)

	entrypoint, err := entry.NewEntrypoint(ctx, cfg)
	if err != nil {
		cancel()
		t.Fatalf("new entrypoint: %v", err)
	}

	entrypoint.Startup(ctx, cancel)

	waitServers(t, cfg)

	return func() {
		cancel()
		entrypoint.Shutdown(ctx)
	}
}

func waitServers(t *testing.T, cfg *config.Config) {
	wg := &sync.WaitGroup{}
	deadline := time.Now().Add(5 * time.Second)

	var (
		sshErr  error
		httpErr error
		ftpErr  error
	)

	{
		wg.Add(1)
		go func() {
			defer wg.Done()
			sshErr = waitTcp(deadline, cfg.Ssh.Address)
		}()
	}

	if cfg.Http.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			httpErr = waitTcp(deadline, cfg.Http.Address)
		}()
	}

	if cfg.Ftp.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ftpErr = waitTcp(deadline, cfg.Ftp.Address)
		}()
	}

	wg.Wait()

	if sshErr != nil {
		t.Fatalf("ssh server not ready: %v", sshErr)
	}
	if httpErr != nil {
		t.Fatalf("http server not ready: %v", httpErr)
	}
	if ftpErr != nil {
		t.Fatalf("ftp server not ready: %v", ftpErr)
	}
}

func waitTcp(deadline time.Time, addr string) error {
	interval := 100 * time.Millisecond

	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid address: %v", err)
	}
	addr = fmt.Sprintf("127.0.0.1:%s", port)

	retErr := fmt.Errorf("server not ready: %s", addr)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, interval)
		if err == nil {
			retErr = nil
			_ = conn.Close()
			break
		}
		retErr = err
		time.Sleep(interval)
	}
	return retErr
}

func WaitAssert(timeout time.Duration, f func() bool) {
	deadline := time.Now().Add(timeout)
	interval := time.Millisecond
	for time.Now().Before(deadline) {
		if f() {
			return
		}
		time.Sleep(interval)
	}
}
