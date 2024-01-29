package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
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

func (s *HttpServer) Enabled() bool {
	return s != nil
}

func (s *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := logs.From(r.Context())

	username, password, ok := r.BasicAuth()
	if ok && s.dashboardServer.Enabled() && s.dashboardServer.Verify(username, password) {
		s.dashboardServer.Handle(w, r)
		return
	}

	if ok {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil || net.ParseIP(ip) == nil {
			logger.Warnf("invalid remote addr %q: %v", r.RemoteAddr, err)
		} else {
			if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
				netIp := net.ParseIP(ip)
				if !netIp.IsGlobalUnicast() || netIp.IsPrivate() {
					splits := strings.Split(forwardedFor, ",")
					forwardedIp := strings.TrimSpace(splits[len(splits)-1])
					if net.ParseIP(forwardedIp) == nil {
						logger.Warnf("invalid X-Forwarded-For %q", forwardedFor)
					} else {
						ip = forwardedIp
					}
				}
			}

			s.handler.Handle(r.Context(), &Request{
				Kind:          model.BruteAttemptKindHttp,
				Time:          time.Now(),
				Ip:            ip,
				User:          username,
				Password:      password,
				SessionId:     uuid.New().String(),
				ClientVersion: r.UserAgent(),
			})
		}
	}

	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
	http.Error(w, "unauthorized", http.StatusUnauthorized)
	return
}

func (s *HttpServer) Startup(ctx context.Context, cancel context.CancelFunc) {
	logger := logs.From(ctx)

	if !s.Enabled() {
		logger.Infof("skip starting ftp server since it is not enabled")
		return
	}
	go func() {
		logger.Infof("start http server, listen on %s", s.server.Addr)
		if err := s.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("listen and serve: %v", err)
		}
		cancel()
	}()
}

func (s *HttpServer) Shutdown(ctx context.Context) error {
	if !s.Enabled() {
		return nil
	}
	logs.From(ctx).Infof("shutdown http server")
	return s.server.Shutdown(ctx)
}
