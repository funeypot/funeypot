package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/wolfogre/funeypot/internal/app/config"
	"github.com/wolfogre/funeypot/internal/app/model"
	"github.com/wolfogre/funeypot/internal/pkg/abuseipdb"
	"github.com/wolfogre/funeypot/internal/pkg/ipapi"
	"github.com/wolfogre/funeypot/internal/pkg/logs"

	"github.com/gliderlabs/ssh"
)

type SshServer struct {
	server *ssh.Server
	delay  time.Duration

	db              *model.Database
	abuseipdbClient *abuseipdb.Client

	queue chan *SshRequest
}

func NewSshServer(cfg config.Ssh, db *model.Database, abuseipdbClient *abuseipdb.Client) *SshServer {
	ret := &SshServer{
		delay:           cfg.Delay,
		db:              db,
		abuseipdbClient: abuseipdbClient,
		queue:           make(chan *SshRequest, 1000),
	}
	ret.server = &ssh.Server{
		Version: "OpenSSH_8.0",
		Addr:    cfg.Address,
		Handler: func(session ssh.Session) {
			_ = session.Exit(0)
		},
		PasswordHandler: ret.handlePassword,
	}

	return ret
}

func (s *SshServer) Startup(ctx context.Context, cancel context.CancelFunc) {
	go func() {
		logger := logs.From(ctx)
		logger.Infof("start ssh server, listen on %s", s.server.Addr)
		if err := s.server.ListenAndServe(); !errors.Is(err, ssh.ErrServerClosed) {
			logger.Errorf("listen and serve: %v", err)
		}
		cancel()
	}()
	go func() {
		s.handleQueue(ctx)
		logs.From(ctx).Infof("handle queue done")
		cancel()
	}()
}

func (s *SshServer) Shutdown(ctx context.Context) error {
	logs.From(ctx).Infof("shutdown ssh server")
	return s.server.Shutdown(ctx)
}

func (s *SshServer) handlePassword(ctx ssh.Context, password string) bool {
	logger := logs.From(ctx)

	wait := time.After(s.delay)

	request := &SshRequest{
		Time:          time.Now(),
		User:          ctx.User(),
		Password:      password,
		SessionId:     ctx.SessionID(),
		ClientVersion: ctx.ClientVersion(),
		RemoteAddr:    ctx.RemoteAddr().String(),
	}
	select {
	case s.queue <- request:
	default:
		logger.Warnf("ssh queue full, drop requests, please increase queue size")
	}

	select {
	case <-ctx.Done():
	case <-wait:
	}
	return false
}

func (s *SshServer) handleQueue(ctx context.Context) {
	logger := logs.From(ctx)
	for {
		if l := len(s.queue); l > 0 {
			logger.Debugf("queue lag: %d", l)
		}
		select {
		case request := <-s.queue:
			s.handleRequest(ctx, request)
		case <-ctx.Done():
			logger.Infof("handle queue done")
			return
		}
	}
}

func (s *SshServer) handleRequest(ctx context.Context, request *SshRequest) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	ip, _, _ := net.SplitHostPort(request.RemoteAddr)
	ip = net.ParseIP(ip).String()

	logger := logs.From(ctx).With(
		"ip", ip,
		"session_id", request.ShortSessionId(),
	)
	ctx = logs.With(ctx, logger)

	attempt, ok, err := s.db.LastSshAttempt(ctx, ip)
	if err != nil {
		logger.Errorf("get last attempt: %v", err)
		return
	}
	if !ok || time.Since(attempt.StoppedAt) > 24*time.Hour {
		attempt = &model.SshAttempt{
			Ip:        ip,
			StartedAt: request.Time,
		}
	}

	attempt.User = request.User
	attempt.Password = request.Password
	attempt.ClientVersion = request.ClientVersion
	attempt.StoppedAt = request.Time
	attempt.Count++ // TODO: use atomic

	loginLogger := logger.With(
		"count", attempt.Count,
		"duration", attempt.Duration().String(),
		"remote_addr", request.RemoteAddr,
		"user", request.User,
		"password", request.Password,
		"client_version", request.ClientVersion,
	)

	geo, err := s.getIpGeo(ctx, ip)
	if err != nil {
		loginLogger.Errorf("get ip geo: %v", err)
	} else {
		loginLogger = loginLogger.With(
			"location", geo.Location,
			"isp", geo.Isp,
		)
	}

	loginLogger.Infof("login")

	if attempt.Id == 0 {
		if err := s.db.Create(ctx, attempt); err != nil {
			logger.Errorf("create attempt: %v", err)
		}
	} else {
		if err := s.db.Update(ctx, attempt, "user", "password", "client_version", "stopped_at", "count"); err != nil {
			logger.Errorf("update attempt: %v", err)
		}
	}

	s.reportAttempt(ctx, attempt)
}

func (s *SshServer) reportAttempt(ctx context.Context, attempt *model.SshAttempt) {
	logger := logs.From(ctx)

	if s.abuseipdbClient == nil {
		return
	}
	if attempt.Count < 5 {
		return
	}
	report, ok, err := s.db.LastAbuseipdbReport(ctx, attempt.Ip)
	if err != nil {
		logger.Errorf("get last report: %v", err)
		return
	}
	if ok && time.Since(report.ReportedAt) < 20*time.Minute {
		return
	}

	comment := fmt.Sprintf(
		"Funeypot detected %d attempts in %s. Last by user %q, password %q, client %q.",
		attempt.Count,
		attempt.Duration().Truncate(time.Second).String(),
		attempt.User,
		attempt.MaskedPassword(),
		attempt.ShortClientVersion(),
	)

	score, err := s.abuseipdbClient.ReportSsh(ctx, attempt.Ip, attempt.StartedAt, comment)
	if err != nil {
		logger.Errorf("report attempt: %v", err)
		return
	}
	logger.Infof("reported, score: %d", score)
	if !report.ReportedAt.IsZero() && report.Score != score {
		logger.Infof("score changed, %d -> %d", report.Score, score)
	}
	newReport := &model.AbuseipdbReport{
		Ip:         attempt.Ip,
		ReportedAt: time.Now(),
		Score:      score,
	}
	if err := s.db.Create(ctx, newReport); err != nil {
		logger.Errorf("create report: %v", err)
	}
}

func (s *SshServer) getIpGeo(ctx context.Context, ip string) (*model.IpGeo, error) {
	logger := logs.From(ctx)

	geo, ok, err := s.db.TaskIpGeo(ctx, ip)
	if err != nil {
		return nil, fmt.Errorf("get ip geo: %w", err)
	}

	if ok && time.Since(geo.CreatedAt) < 24*time.Hour {
		return geo, nil
	}

	result, err := ipapi.Query(ctx, ip)
	if err != nil {
		return nil, err
	}

	geo = (&model.IpGeo{
		Ip: ip,
	}).FillIpapiResponse(result)
	if err := s.db.Save(ctx, geo); err != nil {
		logger.Errorf("save ip geo: %v", err)
		// go on
	}
	return geo, nil
}

var _ Server = (*SshServer)(nil)

type SshRequest struct {
	Time          time.Time
	User          string
	Password      string
	SessionId     string
	ClientVersion string
	RemoteAddr    string
}

func (r SshRequest) ShortSessionId() string {
	if len(r.SessionId) > 8 {
		return r.SessionId[:8]
	}
	return r.SessionId
}
