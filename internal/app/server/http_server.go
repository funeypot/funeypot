package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/wolfogre/funeypot/internal/app/config"
	"github.com/wolfogre/funeypot/internal/app/dashboard"
	"github.com/wolfogre/funeypot/internal/pkg/logs"

	"github.com/gliderlabs/ssh"
)

type HttpServer struct {
	server *http.Server

	dashboardServer *dashboard.Server
}

var _ Server = (*HttpServer)(nil)

func NewHttpServer(cfg config.Http, dashboardServer *dashboard.Server) *HttpServer {
	ret := &HttpServer{}
	ret.server = &http.Server{
		Addr:    cfg.Address,
		Handler: dashboardServer,
	}

	return ret
}

func (s *HttpServer) Startup(ctx context.Context, cancel context.CancelFunc) {
	go func() {
		logger := logs.From(ctx)
		logger.Infof("start http server, listen on %s", s.server.Addr)
		if err := s.server.ListenAndServe(); !errors.Is(err, ssh.ErrServerClosed) {
			logger.Errorf("listen and serve: %v", err)
		}
		cancel()
	}()
}

func (s *HttpServer) Shutdown(ctx context.Context) error {
	logs.From(ctx).Infof("shutdown http server")
	return s.server.Shutdown(ctx)
}
