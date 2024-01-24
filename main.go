package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/wolfogre/funeypot/internal/app/config"
	"github.com/wolfogre/funeypot/internal/app/dashboard"
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

	// TODO: use wire

	db, err := model.NewDatabase(cfg.Database)
	if err != nil {
		logs.From(ctx).Fatalf("new database: %v", err)
		return
	}

	abuseipdbClient := abuseipdb.NewClient(cfg.Abuseipdb.Key)

	sshServer, err := server.NewSshServer(cfg.Ssh, db, abuseipdbClient)
	if err != nil {
		logs.From(ctx).Fatalf("new ssh server: %v", err)
		return
	}
	sshServer.Startup(ctx, cancel)

	dashboardServer, err := dashboard.NewServer(cfg.Dashboard, db)
	if err != nil {
		logs.From(ctx).Fatalf("new dashboard server: %v", err)
		return
	}

	httpServer := server.NewHttpServer(cfg.Http, dashboardServer)
	httpServer.Startup(ctx, cancel)

	<-ctx.Done()
	logger.Infof("shutdown")
	_ = sshServer.Shutdown(ctx)
	_ = httpServer.Shutdown(ctx)
}
