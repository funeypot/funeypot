package main

import (
	"context"
	"errors"
	"flag"
	"github.com/gliderlabs/ssh"
	"github.com/wolfogre/funeypot/internal/app/config"
	"github.com/wolfogre/funeypot/internal/app/handler"
	"github.com/wolfogre/funeypot/internal/app/model"
	"github.com/wolfogre/funeypot/internal/pkg/logs"
	"net/http"
	"os"
	"os/signal"
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

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
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

	h := handler.New(ctx, cfg.Ssh.Delay, cfg.Abuseipdb.Key, db)

	logger.With("addr", cfg.Ssh.Address, "delay", cfg.Ssh.Delay.String()).Infof("start listening")

	sever := &ssh.Server{
		Version: "OpenSSH_8.0",
		Addr:    cfg.Ssh.Address,
		Handler: func(session ssh.Session) {
			_ = session.Exit(0)
		},
		PasswordHandler: h.Handle,
	}
	go func() {
		if err := sever.ListenAndServe(); !errors.Is(err, ssh.ErrServerClosed) {
			logs.From(ctx).Errorf("listen and serve: %v", err)
		}
		cancel()
	}()

	httpServer := &http.Server{
		Addr:    ":9200",
		Handler: http.HandlerFunc(h.HandleHttp),
	}
	logger.Infof("start http listening")
	go func() {
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logs.From(ctx).Errorf("listen and serve: %v", err)
		}
		cancel()
	}()

	<-ctx.Done()
	logs.From(ctx).Infof("shutdown")
	_ = sever.Shutdown(ctx)
	_ = httpServer.Shutdown(ctx)
}
