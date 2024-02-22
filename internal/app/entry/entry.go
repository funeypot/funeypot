// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package entry

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/funeypot/funeypot/internal/app/config"
	"github.com/funeypot/funeypot/internal/pkg/logs"
)

func Run(ctx context.Context, version string, args []string) error {
	defer logs.Close()

	set := flag.NewFlagSet("funeypot", flag.ContinueOnError)
	configFile := set.String("c", "config.yaml", "config file")
	configDisableGenerate := set.Bool("disable-generate", false, "don't generate config file if not exists")
	if err := set.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		} else {
			return err
		}
	}

	logger := logs.Default()
	logger.Infof("funeypot %s starting", version)

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load(*configFile, !*configDisableGenerate)
	if err != nil {
		logger.Errorf("load config: %v", err)
		return err
	}
	logger = logs.SetLevel(cfg.Log.Level)
	ctx = logs.With(ctx, logger)

	entrypoint, err := NewEntrypoint(ctx, cfg)
	if err != nil {
		logger.Errorf("new entrypoint: %v", err)
		return err
	}

	entrypoint.Startup(ctx, cancel)

	<-ctx.Done()
	logger.Infof("shutdown")
	entrypoint.Shutdown(ctx)

	return nil
}
