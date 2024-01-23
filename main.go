package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/wolfogre/funeypot/internal/app/config"
	"github.com/wolfogre/funeypot/internal/app/model"
	"github.com/wolfogre/funeypot/internal/app/server"
	"github.com/wolfogre/funeypot/internal/pkg/abuseipdb"
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

	cfg, err := config.Load(configFile)
	if err != nil {
		logs.From(ctx).Fatalf("load config: %v", err)
		return
	}

	db, err := model.NewDatabase(cfg.Database)
	if err != nil {
		logs.From(ctx).Fatalf("new database: %v", err)
		return
	}

	abuseipdbClient := abuseipdb.NewClient(cfg.Abuseipdb.Key)

	sshServer := server.NewSshServer(cfg.Ssh, db, abuseipdbClient)
	sshServer.Startup(ctx, cancel)

	<-ctx.Done()
	logger.Infof("shutdown")
	_ = sshServer.Shutdown(ctx)
}
