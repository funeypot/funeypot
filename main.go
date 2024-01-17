package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/gliderlabs/ssh"
)

var (
	addr  = ":2222"
	delay = 2 * time.Second
)

func init() {
	flag.StringVar(&addr, "addr", addr, "address to listen")
	flag.DurationVar(&delay, "delay", delay, "delay to login")
}

func main() {
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})))

	slog.Info("start listening",
		"addr", addr,
		"delay", delay.String(),
	)

	sever := &ssh.Server{
		Version: "OpenSSH_8.0",
		Addr:    addr,
		Handler: func(session ssh.Session) {
			_, _ = fmt.Fprintln(session, "Fuck you")
		},
		PasswordHandler: func(ctx ssh.Context, password string) bool {
			sessionId := ctx.SessionID()
			if len(sessionId) > 8 {
				sessionId = sessionId[:8]
			}

			remoteIp, _, _ := net.SplitHostPort(ctx.RemoteAddr().String())

			record := GetRecord(remoteIp)

			slog.Info("new login",
				"session_id", sessionId,
				"user", ctx.User(),
				"password", password,
				"client_version", ctx.ClientVersion(),
				"remote_ip", remoteIp,
				"count", record.Count,
				"duration", record.Duration().String(),
			)
			if password == "test" {
				return true
			}
			select {
			case <-ctx.Done():
				return false
			case <-time.After(delay):
				return false
			}
		},
	}
	if err := sever.ListenAndServe(); err != nil {
		slog.Error("listen and serve failed",
			"error", err,
		)
	}
}
