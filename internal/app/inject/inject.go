package inject

import (
	"github.com/wolfogre/funeypot/internal/app/config"
	"github.com/wolfogre/funeypot/internal/app/dashboard"
	"github.com/wolfogre/funeypot/internal/app/model"
	"github.com/wolfogre/funeypot/internal/app/server"
	"github.com/wolfogre/funeypot/internal/pkg/abuseipdb"

	"github.com/google/wire"
)

type Entrypoint struct {
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

var providerSet = wire.NewSet(
	newEntrypoint,
	config.Load,
	wire.FieldsOf(new(*config.Config),
		"Database",
		"Abuseipdb",
		"Dashboard",
		"Ssh",
		"Http",
		"Ftp",
	),
	model.NewDatabase,
	abuseipdb.NewClient,
	dashboard.NewServer,
	server.NewHandler,
	server.NewSshServer,
	server.NewHttpServer,
	server.NewFtpServer,
)
