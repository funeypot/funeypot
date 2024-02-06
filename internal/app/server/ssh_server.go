// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/wolfogre/funeypot/internal/app/config"
	"github.com/wolfogre/funeypot/internal/app/model"
	"github.com/wolfogre/funeypot/internal/pkg/fakever"
	"github.com/wolfogre/funeypot/internal/pkg/logs"
	"github.com/wolfogre/funeypot/internal/pkg/sshkey"

	"github.com/gliderlabs/ssh"
)

type SshServer struct {
	server *ssh.Server
	delay  time.Duration

	handler *Handler
}

var _ Server = (*SshServer)(nil)

func NewSshServer(cfg config.Ssh, handler *Handler) (*SshServer, error) {
	ret := &SshServer{
		delay:   cfg.Delay,
		handler: handler,
	}

	signer, err := sshkey.GenerateSigner(cfg.KeySeed)
	if err != nil {
		return nil, fmt.Errorf("generate signer: %w", err)
	}

	ret.server = &ssh.Server{
		HostSigners: []ssh.Signer{signer},
		Version:     fakever.SshVersion,
		Addr:        cfg.Address,
		Handler: func(session ssh.Session) {
			_ = session.Exit(0)
		},
		PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
			logs.From(ctx).Infof("public key: %v", key.Type())
			return false
		},
		PasswordHandler: ret.handlePassword,
		ConnectionFailedCallback: func(conn net.Conn, reason error) {
			logs.Default().Infof("connection %s failed: %v", conn.RemoteAddr(), reason)
		},
	}

	return ret, nil
}

func (s *SshServer) Startup(ctx context.Context, cancel context.CancelFunc) {
	go func() {
		logger := logs.From(ctx)
		logger.Infof("start ssh server, listen on %s", s.server.Addr)
		if err := s.server.ListenAndServe(); !errors.Is(err, ssh.ErrServerClosed) {
			logger.Errorf("listen and serve: %v", err)
		}
		cancel()
	}()
}

func (s *SshServer) Shutdown(ctx context.Context) error {
	logs.From(ctx).Infof("shutdown ssh server")
	return s.server.Shutdown(ctx)
}

func (s *SshServer) handlePassword(ctx ssh.Context, password string) bool {
	logger := logs.From(ctx)

	remoteAddr := ctx.RemoteAddr().String()
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil || net.ParseIP(ip) == nil {
		logger.Warnf("invalid remote addr %q: %v", remoteAddr, err)
	} else {
		s.handler.Handle(ctx, &Request{
			Kind:          model.BruteAttemptKindSsh,
			Ip:            ip,
			Time:          time.Now(),
			User:          ctx.User(),
			Password:      password,
			SessionId:     ctx.SessionID(),
			ClientVersion: ctx.ClientVersion(),
		})
	}

	wait := time.After(s.delay)

	select {
	case <-ctx.Done():
	case <-wait:
	}
	return false
}
