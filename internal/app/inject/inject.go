// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package inject

import (
	"context"

	"github.com/wolfogre/funeypot/internal/app/config"
	"github.com/wolfogre/funeypot/internal/app/dashboard"
	"github.com/wolfogre/funeypot/internal/app/model"
	"github.com/wolfogre/funeypot/internal/app/server"
	"github.com/wolfogre/funeypot/internal/pkg/logs"

	"github.com/google/wire"
)

type Entrypoint struct {
	Config     *config.Config
	SshServer  *server.SshServer
	HttpServer *server.HttpServer
	FtpServer  *server.FtpServer
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

var providerSet = wire.NewSet(
	newEntrypoint,
	wire.FieldsOf(new(*config.Config),
		"Database",
		"Abuseipdb",
		"Dashboard",
		"Ssh",
		"Http",
		"Ftp",
	),
	model.NewDatabase,
	newAbuseipdbClient,
	dashboard.NewServer,
	server.NewHandler,
	server.NewSshServer,
	server.NewHttpServer,
	server.NewFtpServer,
)

// to suppress "unused" error
var _ = providerSet
