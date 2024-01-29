package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/wolfogre/funeypot/internal/app/inject"
	"github.com/wolfogre/funeypot/internal/pkg/logs"
)

var (
	configFile = "config.yaml"
)

func init() {
	flag.StringVar(&configFile, "c", configFile, "config file")
}

func main() {
	flag.Parse()
	defer logs.Default().Sync()

	logger := logs.Default()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	ctx = logs.With(ctx, logger)

	entrypoint, err := inject.NewEntrypoint(ctx, configFile)
	if err != nil {
		logs.From(ctx).Fatalf("new entrypoint: %v", err)
		return
	}

	entrypoint.SshServer.Startup(ctx, cancel)
	entrypoint.HttpServer.Startup(ctx, cancel)
	entrypoint.FtpServer.Startup(ctx, cancel)

	<-ctx.Done()
	logger.Infof("shutdown")
	_ = entrypoint.SshServer.Shutdown(ctx)
	_ = entrypoint.HttpServer.Shutdown(ctx)
	_ = entrypoint.FtpServer.Shutdown(ctx)
}
