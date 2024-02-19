// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/funeypot/funeypot/internal/app/config"
	"github.com/funeypot/funeypot/internal/app/inject"
	"github.com/funeypot/funeypot/internal/pkg/logs"
)

var (
	Version               = "dev"
	configFile            = "config.yaml"
	configDisableGenerate = false
)

func init() {
	flag.StringVar(&configFile, "c", configFile, "config file")
	flag.BoolVar(&configDisableGenerate, "disable-generate", configDisableGenerate, "don't generate config file if not exists")
}

func main() {
	defer logs.Close()

	flag.Parse()

	logger := logs.Default()
	logger.Infof("funeypot %s starting", Version)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load(configFile, !configDisableGenerate)
	if err != nil {
		logger.Fatalf("load config: %v", err)
		return
	}
	logger = logs.SetLevel(cfg.Log.Level)
	ctx = logs.With(ctx, logger)

	entrypoint, err := inject.NewEntrypoint(ctx, cfg)
	if err != nil {
		logger.Fatalf("new entrypoint: %v", err)
		return
	}

	entrypoint.Startup(ctx, cancel)

	<-ctx.Done()
	logger.Infof("shutdown")
	entrypoint.Shutdown(ctx)
}
