// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package inject

import (
	"github.com/wolfogre/funeypot/internal/app/config"
	"github.com/wolfogre/funeypot/internal/app/server"
	"github.com/wolfogre/funeypot/internal/pkg/abuseipdb"
)

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

func newAbuseipdbClient(cfg config.Abuseipdb) *abuseipdb.Client {
	if !cfg.Enabled {
		return nil
	}
	return abuseipdb.NewClient(cfg.Key, cfg.Interval)
}
