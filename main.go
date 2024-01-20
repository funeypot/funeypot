package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/wolfogre/funeypot/internal/app/handler"
	"github.com/wolfogre/funeypot/internal/app/model"
	"github.com/wolfogre/funeypot/internal/pkg/logs"

	"github.com/gliderlabs/ssh"
)

var (
	addr         = ":2222"
	delay        = 2 * time.Second
	abuseIpdbKey = ""
	dataDir      = "/tmp/funeypot"
)

func init() {
	flag.StringVar(&addr, "addr", addr, "address to listen")
	flag.DurationVar(&delay, "delay", delay, "delay to login")
	flag.StringVar(&abuseIpdbKey, "abuseipdb-key", abuseIpdbKey, "abuseipdb key")
	flag.StringVar(&dataDir, "data-dir", dataDir, "data dir")
}

func main() {
	flag.Parse()
	defer logs.Default().Sync()

	logger := logs.Default()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()
	ctx = logs.With(ctx, logger)

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		logger.Fatalf("mkdir: %v", err)
		return
	}

	db, err := model.NewDatabase(ctx, filepath.Join(dataDir, "funeypot.db"))
	if err != nil {
		logs.From(ctx).Fatalf("new database: %v", err)
		return
	}

	h := handler.New(ctx, delay, abuseIpdbKey, db)

	logger.With("addr", addr, "delay", delay.String()).Infof("start listening")

	sever := &ssh.Server{
		Version: "OpenSSH_8.0",
		Addr:    addr,
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

	<-ctx.Done()
	logs.From(ctx).Infof("shutdown")
	_ = sever.Shutdown(ctx)
}
