// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package inject

import (
	"context"

	"github.com/funeypot/funeypot/internal/app/config"
	"github.com/funeypot/funeypot/internal/app/server"
	"github.com/funeypot/funeypot/internal/pkg/logs"
)

type Entrypoint struct {
	Config     *config.Config
	SshServer  *server.SshServer
	HttpServer *server.HttpServer
	FtpServer  *server.FtpServer
}

func newEntrypoint(
	sshServer *server.SshServer,
	httpServer *server.HttpServer,
	ftpServer *server.FtpServer,
) *Entrypoint {
	return &Entrypoint{
		SshServer:  sshServer,
		HttpServer: httpServer,
		FtpServer:  ftpServer,
	}
}

func (e *Entrypoint) Startup(ctx context.Context, cancel context.CancelFunc) {
	e.SshServer.Startup(ctx, cancel)
	e.HttpServer.Startup(ctx, cancel)
	e.FtpServer.Startup(ctx, cancel)
}

func (e *Entrypoint) Shutdown(ctx context.Context) {
	logger := logs.From(ctx)
	if err := e.SshServer.Shutdown(ctx); err != nil {
		logger.Warnf("shutdown ssh server: %v", err)
	}
	if err := e.HttpServer.Shutdown(ctx); err != nil {
		logger.Warnf("shutdown http server: %v", err)
	}
	if err := e.FtpServer.Shutdown(ctx); err != nil {
		logger.Warnf("shutdown ftp server: %v", err)
	}
}
