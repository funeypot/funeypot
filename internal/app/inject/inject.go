package inject

import (
	"github.com/wolfogre/funeypot/internal/app/config"
	"github.com/wolfogre/funeypot/internal/app/dashboard"
	"github.com/wolfogre/funeypot/internal/app/model"
	"github.com/wolfogre/funeypot/internal/app/server"

	"github.com/google/wire"
)

type Entrypoint struct {
	Config     *config.Config
	SshServer  *server.SshServer
	HttpServer *server.HttpServer
	FtpServer  *server.FtpServer
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
