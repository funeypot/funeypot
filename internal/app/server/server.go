// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"fmt"
	"time"

	"github.com/funeypot/funeypot/internal/app/model"
	"github.com/funeypot/funeypot/internal/pkg/abuseipdb"
	"github.com/funeypot/funeypot/internal/pkg/ipgeo"
	"github.com/funeypot/funeypot/internal/pkg/logs"
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
	ipgeoQuerier    ipgeo.Querier

	queue chan *Request
}

func NewHandler(ctx context.Context, db *model.Database, ipgeoQuerier ipgeo.Querier, abuseipdbClient *abuseipdb.Client) *Handler {
	ret := &Handler{
		db:              db,
		ipgeoQuerier:    ipgeoQuerier,
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
			if len(h.queue) > 0 {
				logger.Warnf("ignore %d unhandled requests", len(h.queue))
			}
			return
		}
	}
}

func (h *Handler) handleRequest(ctx context.Context, request *Request) {
	logger := logs.From(ctx)

	logger.Debugf("handle request: %v", request)

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

	geo, err := h.ipgeoQuerier.Query(ctx, request.Ip)
	if err != nil {
		loginLogger.Errorf("get ip geo: %v", err)
	} else {
		loginLogger = loginLogger.With(
			"location", geo.Location,
		)
	}

	loginLogger.Infof("login")

	h.reportAttempt(ctx, attempt)
}

func (h *Handler) reportAttempt(ctx context.Context, attempt *model.BruteAttempt) {
	logger := logs.From(ctx)

	if !h.abuseipdbClient.Enabled() {
		return
	}
	if attempt.Count < 5 {
		return
	}
	if until, ok := h.abuseipdbClient.Cooldown(); ok {
		logger.Debugf("abuseipdb cooldown, until: %v", until.Format(time.RFC3339))
		return
	}
	report, ok, err := h.db.LastAbuseipdbReport(ctx, attempt.Ip)
	if err != nil {
		logger.Errorf("get last report: %v", err)
		return
	}
	if ok && time.Since(report.ReportedAt) < h.abuseipdbClient.Interval() {
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
		score, err = h.abuseipdbClient.ReportSsh(ctx, attempt.Ip, attempt.StoppedAt, comment)
	case model.BruteAttemptKindHttp:
		score, err = h.abuseipdbClient.ReportHttp(ctx, attempt.Ip, attempt.StoppedAt, comment)
	case model.BruteAttemptKindFtp:
		score, err = h.abuseipdbClient.ReportFtp(ctx, attempt.Ip, attempt.StoppedAt, comment)
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
