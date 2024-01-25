package server

import (
	"context"
	"fmt"
	"time"

	"github.com/wolfogre/funeypot/internal/app/model"
	"github.com/wolfogre/funeypot/internal/pkg/abuseipdb"
	"github.com/wolfogre/funeypot/internal/pkg/ipapi"
	"github.com/wolfogre/funeypot/internal/pkg/logs"
)

type Server interface {
	Startup(ctx context.Context, cancel context.CancelFunc)
	Shutdown(ctx context.Context) error
}

type Request struct {
	Kind          model.BruteAttemptKind
	Time          time.Time
	Ip            string
	User          string
	Password      string
	SessionId     string
	ClientVersion string
}

func (r Request) ShortSessionId() string {
	if len(r.SessionId) > 8 {
		return r.SessionId[:8]
	}
	return r.SessionId
}

type Handler struct {
	db              *model.Database
	abuseipdbClient *abuseipdb.Client

	queue chan *Request
}

func NewHandler(ctx context.Context, db *model.Database, abuseipdbClient *abuseipdb.Client) *Handler {
	ret := &Handler{
		db:              db,
		abuseipdbClient: abuseipdbClient,
		queue:           make(chan *Request, 1000),
	}
	go ret.handleQueue(ctx)
	return ret
}

func (h *Handler) Handle(ctx context.Context, request *Request) {
	logger := logs.From(ctx)
	select {
	case h.queue <- request:
	default:
		logger.Warnf("ssh queue full, drop requests, please increase queue size")
	}
}

func (h *Handler) handleQueue(ctx context.Context) {
	logger := logs.From(ctx)
	for {
		if l := len(h.queue); l > 0 {
			logger.Debugf("queue lag: %d", l)
		}
		select {
		case request := <-h.queue:
			subCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			subCtx = logs.With(subCtx, logger.With(
				"kind", request.Kind.String(),
				"ip", request.Ip,
				"session_id", request.ShortSessionId(),
			))
			h.handleRequest(subCtx, request)
			cancel()
		case <-ctx.Done():
			logger.Infof("handle queue done")
			return
		}
	}
}

func (h *Handler) handleRequest(ctx context.Context, request *Request) {
	logger := logs.From(ctx)

	attempt, err := h.db.IncrBruteAttempt(
		ctx,
		request.Ip,
		request.Kind,
		request.Time,
		request.User, request.Password, request.ClientVersion,
		request.Time.Add(-24*time.Hour),
	)
	if err != nil {
		logger.Errorf("incr attempt: %v", err)
		return
	}

	loginLogger := logger.With(
		"count", attempt.Count,
		"duration", attempt.Duration().String(),
		"ip", request.Ip,
		"user", request.User,
		"password", request.Password,
		"client_version", request.ClientVersion,
	)

	geo, err := h.getIpGeo(ctx, request.Ip)
	if err != nil {
		loginLogger.Errorf("get ip geo: %v", err)
	} else {
		loginLogger = loginLogger.With(
			"location", geo.Location,
			"isp", geo.Isp,
		)
	}

	loginLogger.Infof("login")

	h.reportAttempt(ctx, attempt)
}

func (h *Handler) getIpGeo(ctx context.Context, ip string) (*model.IpGeo, error) {
	logger := logs.From(ctx)

	geo, ok, err := h.db.TaskIpGeo(ctx, ip)
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
	if err := h.db.Save(ctx, geo); err != nil {
		logger.Errorf("save ip geo: %v", err)
		// go on
	}
	return geo, nil
}

func (h *Handler) reportAttempt(ctx context.Context, attempt *model.BruteAttempt) {
	logger := logs.From(ctx)

	if h.abuseipdbClient == nil {
		return
	}
	if attempt.Count < 5 {
		return
	}
	report, ok, err := h.db.LastAbuseipdbReport(ctx, attempt.Ip)
	if err != nil {
		logger.Errorf("get last report: %v", err)
		return
	}
	if ok && time.Since(report.ReportedAt) < 20*time.Minute {
		return
	}

	comment := fmt.Sprintf(
		"Funeypot detected %d %s attempts in %s. Last by user %q, password %q, client %q.",
		attempt.Count,
		attempt.Kind.String(),
		attempt.Duration().Truncate(time.Second).String(),
		attempt.User,
		attempt.MaskedPassword(),
		attempt.ShortClientVersion(),
	)

	var score int
	switch attempt.Kind {
	case model.BruteAttemptKindSsh:
		score, err = h.abuseipdbClient.ReportSsh(ctx, attempt.Ip, attempt.StartedAt, comment)
	case model.BruteAttemptKindHttp:
		score, err = h.abuseipdbClient.ReportHttp(ctx, attempt.Ip, attempt.StartedAt, comment)
	}
	if err != nil {
		logger.Errorf("report attempt: %v", err)
		return
	}

	logger.Infof("reported, score: %d", score)
	if report != nil && report.Score != score {
		logger.Infof("score changed, %d -> %d", report.Score, score)
	}
	newReport := &model.AbuseipdbReport{
		Ip:         attempt.Ip,
		ReportedAt: time.Now(),
		Score:      score,
	}
	if err := h.db.Create(ctx, newReport); err != nil {
		logger.Errorf("create report: %v", err)
	}
}
