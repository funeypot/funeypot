// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package inject

import (
	"github.com/wolfogre/funeypot/internal/app/config"
	"github.com/wolfogre/funeypot/internal/app/dashboard"
	"github.com/wolfogre/funeypot/internal/app/model"
	"github.com/wolfogre/funeypot/internal/app/server"
	"github.com/wolfogre/funeypot/internal/pkg/abuseipdb"
	"github.com/wolfogre/funeypot/internal/pkg/ipgeo"

	"github.com/google/wire"
)

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
	newCachedIpGeoQuerier,
)

// to suppress "unused" error
var _ = providerSet

func newAbuseipdbClient(cfg config.Abuseipdb) *abuseipdb.Client {
	if !cfg.Enabled {
		return nil
	}
	return abuseipdb.NewClient(cfg.Key, cfg.Interval)
}

func newCachedIpGeoQuerier(db *model.Database) ipgeo.Querier {
	return model.NewCachedIpGeoQuerier(ipgeo.NewIpapiComQuerier(), db)
}
