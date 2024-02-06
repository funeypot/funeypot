// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"time"

	"github.com/wolfogre/funeypot/internal/app/config"
	"github.com/wolfogre/funeypot/internal/app/model"
	"github.com/wolfogre/funeypot/internal/pkg/logs"

	"github.com/fclairamb/ftpserverlib"
	"github.com/google/uuid"
)

type FtpServer struct {
	server *ftpserver.FtpServer
	addr   string

	handler *Handler
}

var _ Server = (*FtpServer)(nil)

func NewFtpServer(cfg config.Ftp, handler *Handler) *FtpServer {
	if !cfg.Enabled {
		return nil
	}

	ret := &FtpServer{
		addr:    cfg.Address,
		handler: handler,
	}

	ret.server = ftpserver.NewFtpServer(ret)

	return ret
}

func (s *FtpServer) Enabled() bool {
	return s != nil
}

func (s *FtpServer) Startup(ctx context.Context, cancel context.CancelFunc) {
	logger := logs.From(ctx)

	if !s.Enabled() {
		logger.Infof("skip starting ftp server since it is not enabled")
		return
	}
	go func() {
		logger.Infof("start ftp server, listen on %s", s.addr)
		if err := s.server.ListenAndServe(); err != nil {
			logger.Errorf("listen and serve: %v", err)
		}
		cancel()
	}()
}

func (s *FtpServer) Shutdown(ctx context.Context) error {
	if !s.Enabled() {
		return nil
	}

	logs.From(ctx).Infof("shutdown ftp server")
	return s.server.Stop()
}

var _ ftpserver.MainDriver = (*FtpServer)(nil)

func (s *FtpServer) GetSettings() (*ftpserver.Settings, error) {
	return &ftpserver.Settings{
		ListenAddr: s.addr,
	}, nil
}

func (s *FtpServer) ClientConnected(cc ftpserver.ClientContext) (string, error) {
	logs.Default().Debugf("ftp client connected: %s", cc.RemoteAddr().String())
	return "", nil
}

func (s *FtpServer) ClientDisconnected(cc ftpserver.ClientContext) {
	logs.Default().Debugf("ftp client disconnected: %s", cc.RemoteAddr().String())
}

func (s *FtpServer) AuthUser(cc ftpserver.ClientContext, user, pass string) (ftpserver.ClientDriver, error) {
	ctx := context.Background()

	logger := logs.From(ctx)

	remoteAddr := cc.RemoteAddr().String()
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil || net.ParseIP(ip) == nil {
		logger.Warnf("invalid remote addr %q: %v", remoteAddr, err)
	} else {
		s.handler.Handle(ctx, &Request{
			Kind:          model.BruteAttemptKindFtp,
			Ip:            ip,
			Time:          time.Now(),
			User:          user,
			Password:      pass,
			SessionId:     uuid.New().String(),
			ClientVersion: cc.GetClientVersion(),
		})
	}

	return nil, errors.New("invalid user or password")
}

func (s *FtpServer) GetTLSConfig() (*tls.Config, error) {
	return nil, errors.New("TLS is not configured")
}
