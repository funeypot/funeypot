package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/wolfogre/funeypot/internal/app/config"
	"github.com/wolfogre/funeypot/internal/app/dashboard"
	"github.com/wolfogre/funeypot/internal/app/model"
	"github.com/wolfogre/funeypot/internal/pkg/logs"

	"github.com/google/uuid"
)

type HttpServer struct {
	server *http.Server

	handler         *Handler
	dashboardServer *dashboard.Server
}

var _ Server = (*HttpServer)(nil)

func NewHttpServer(cfg config.Http, handler *Handler, dashboardServer *dashboard.Server) *HttpServer {
	ret := &HttpServer{
		dashboardServer: dashboardServer,
		handler:         handler,
	}
	ret.server = &http.Server{
		Addr:    cfg.Address,
		Handler: ret,
	}

	return ret
}

func (s *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if ok && s.dashboardServer.Verify(username, password) {
		s.dashboardServer.Handle(w, r)
		return
	}

	if ok {
		request := &Request{
			Kind:          model.BruteAttemptKindHttp,
			Time:          time.Now(),
			User:          username,
			Password:      password,
			SessionId:     uuid.New().String(),
			ClientVersion: r.UserAgent(),
			RemoteAddr:    r.RemoteAddr, // TODO: get real remote addr from proxy
		}
		s.handler.Handle(r.Context(), request)
	}

	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
	http.Error(w, "unauthorized", http.StatusUnauthorized)
	return
}

func (s *HttpServer) Startup(ctx context.Context, cancel context.CancelFunc) {
	go func() {
		logger := logs.From(ctx)
		logger.Infof("start http server, listen on %s", s.server.Addr)
		if err := s.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("listen and serve: %v", err)
		}
		cancel()
	}()
}

func (s *HttpServer) Shutdown(ctx context.Context) error {
	logs.From(ctx).Infof("shutdown http server")
	return s.server.Shutdown(ctx)
}
