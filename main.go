package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/wolfogre/funeypot/internal/app/cache"
	"github.com/wolfogre/funeypot/internal/pkg/logs"

	"github.com/gliderlabs/ssh"
)

var (
	addr         = ":2222"
	delay        = 2 * time.Second
	abuseIpDbKey = ""
	cacheDir     = "/tmp/funeypot"
)

func init() {
	flag.StringVar(&addr, "addr", addr, "address to listen")
	flag.DurationVar(&delay, "delay", delay, "delay to login")
	flag.StringVar(&abuseIpDbKey, "abuseipdb-key", abuseIpDbKey, "abuseipdb key")
	flag.StringVar(&cacheDir, "cache-dir", cacheDir, "cache dir")
}

var (
	Cache *cache.Cache
)

func main() {
	flag.Parse()
	defer logs.Default().Sync()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()
	ctx = logs.With(ctx, logs.Default())

	c, err := cache.New(ctx, cacheDir)
	if err != nil {
		logs.From(ctx).Fatalf("new cache: %v", err)
		return
	}
	Cache = c

	StartReport(ctx)

	logs.From(ctx).With("addr", addr, "delay", delay.String()).Infof("start listening")

	sever := &ssh.Server{
		Version: "OpenSSH_8.0",
		Addr:    addr,
		Handler: func(session ssh.Session) {
			_, _ = fmt.Fprintln(session, "Fuck you")
		},
		PasswordHandler: func(ctx ssh.Context, password string) bool {
			after := time.After(delay)

			sessionId := ctx.SessionID()
			if len(sessionId) > 8 {
				sessionId = sessionId[:8]
			}

			remoteIp, _, _ := net.SplitHostPort(ctx.RemoteAddr().String())

			logger := logs.From(ctx).With(
				"session_id", sessionId,
				"user", ctx.User(),
				"password", password,
				"client_version", ctx.ClientVersion(),
				"remote_ip", remoteIp,
			)

			record, err := Cache.IncrRecord(ctx, remoteIp)
			if err != nil {
				logs.From(ctx).Errorf("get record: %v", err)
			} else {
				logger = logger.With(
					"count", record.Count,
					"duration", record.Duration().String(),
					"geo", record.Geo,
				)
			}

			logger.Infof("login")
			select {
			case <-ctx.Done():
				return false
			case <-after:
				return false
			}
		},
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
