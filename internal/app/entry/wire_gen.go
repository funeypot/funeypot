// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package entry

import (
	"context"
	"github.com/funeypot/funeypot/internal/app/config"
	"github.com/funeypot/funeypot/internal/app/dashboard"
	"github.com/funeypot/funeypot/internal/app/model"
	"github.com/funeypot/funeypot/internal/app/server"
)

// Injectors from wire.go:

func NewEntrypoint(ctx context.Context, cfg *config.Config) (*Entrypoint, error) {
	ssh := cfg.Ssh
	database := cfg.Database
	modelDatabase, err := model.NewDatabase(ctx, database)
	if err != nil {
		return nil, err
	}
	querier := newCachedIpGeoQuerier(modelDatabase)
	abuseipdb := cfg.Abuseipdb
	client := newAbuseipdbClient(abuseipdb)
	handler := server.NewHandler(ctx, modelDatabase, querier, client)
	sshServer, err := server.NewSshServer(ssh, handler)
	if err != nil {
		return nil, err
	}
	http := cfg.Http
	configDashboard := cfg.Dashboard
	dashboardServer, err := dashboard.NewServer(configDashboard, modelDatabase, querier)
	if err != nil {
		return nil, err
	}
	httpServer := server.NewHttpServer(http, handler, dashboardServer)
	ftp := cfg.Ftp
	ftpServer := server.NewFtpServer(ftp, handler)
	entrypoint := newEntrypoint(sshServer, httpServer, ftpServer)
	return entrypoint, nil
}
