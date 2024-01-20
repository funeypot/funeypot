package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/wolfogre/funeypot/internal/app/model"
	"github.com/wolfogre/funeypot/internal/pkg/logs"

	"github.com/gliderlabs/ssh"
	"github.com/gochore/boltutil"
)

type Handler struct {
	delay        time.Duration
	abuseIpdbKey string
	db           *boltutil.DB
	queue        chan *Request
}

func New(ctx context.Context, delay time.Duration, abuseIpdbKey string, db *boltutil.DB) *Handler {
	ret := &Handler{
		delay:        delay,
		abuseIpdbKey: abuseIpdbKey,
		db:           db,
		queue:        make(chan *Request, 1000),
	}
	go ret.handleQueue(ctx)
	return ret
}

type Request struct {
	Time          time.Time
	User          string
	Password      string
	SessionId     string
	ClientVersion string
	RemoteAddr    string
}

func (h *Handler) Handle(ctx ssh.Context, password string) bool {
	logger := logs.From(ctx)

	wait := time.After(h.delay)

	request := &Request{
		Time:          time.Now(),
		User:          ctx.User(),
		Password:      password,
		SessionId:     ctx.SessionID(),
		ClientVersion: ctx.ClientVersion(),
		RemoteAddr:    ctx.RemoteAddr().String(),
	}
	select {
	case h.queue <- request:
	default:
		logger.Warnf("queue full, drop requests, please increase queue size")
	}

	select {
	case <-ctx.Done():
	case <-wait:
	}
	return false
}

func (h *Handler) handleQueue(ctx context.Context) {
	for {
		select {
		case request := <-h.queue:
			h.handleRequest(ctx, request)
		case <-ctx.Done():
			logs.From(ctx).Infof("handle queue done")
			return
		}
	}
}

func (h *Handler) handleRequest(ctx context.Context, request *Request) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	ip, _, _ := net.SplitHostPort(request.RemoteAddr)
	logger := logs.From(ctx).With(
		"ip", ip,
		"remote_addr", request.RemoteAddr,
		"session_id", shortenSessionId(request.SessionId),
		"user", request.User,
		"password", request.Password,
		"client_version", request.ClientVersion,
	)

	record := &model.Record{
		Ip:        ip,
		StartedAt: request.Time,
	}
	if err := h.db.Get(record); err != nil && !errors.Is(err, boltutil.ErrNotExist) {
		logger.Errorf("get record: %v", err)
		return
	}

	record.User = request.User
	record.Password = request.Password
	record.ClientVersion = request.ClientVersion
	record.StoppedAt = request.Time
	record.Count++
	logger = logger.With(
		"count", record.Count,
		"duration", record.Duration().String(),
	)

	logger.Infof("login")

	if h.abuseIpdbKey != "" && record.Count >= 5 && time.Since(record.ReportedAt) > 20*time.Minute {
		score, err := h.reportRecord(ctx, record)
		if err != nil {
			logger.Errorf("report record: %v", err)
		} else {
			logger.Infof("reported, score: %d", score)
			if !record.ReportedAt.IsZero() && record.Score != score {
				logger.Infof("score changed, %d -> %d", record.Score, score)
			}
			record.ReportedAt = time.Now()
			record.Score = score
		}
	}

	if err := h.db.Put(record); err != nil {
		logger.Errorf("put record: %v", err)
	}
}

func (h *Handler) reportRecord(ctx context.Context, record *model.Record) (int, error) {
	data := url.Values{}
	data.Set("ip", record.Ip)
	data.Add("categories", "18,22")
	data.Add("timestamp", time.Now().Format(time.RFC3339))

	comment := fmt.Sprintf(
		"Caught by funeypot, tried to crack SSH password %d times within %s. Last attempt by user %s with password '%s' via client %s.",
		record.Count,
		record.Duration().Truncate(time.Second).String(),
		record.User,
		record.MaskedPassword(),
		record.ClientVersion,
	)
	data.Add("comment", comment)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.abuseipdb.com/api/v2/report", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return 0, fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Key", h.abuseIpdbKey)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	result := &abuseIpdbResponse{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return 0, fmt.Errorf("decode response: %w", err)
	}

	if len(result.Errors) > 0 {
		return 0, fmt.Errorf("report abuse ip db: %v", result.Errors)
	}

	return result.Data.AbuseConfidenceScore, nil
}

func shortenSessionId(sessionId string) string {
	if len(sessionId) > 8 {
		sessionId = sessionId[:8]
	}
	return sessionId
}

type abuseIpdbResponse struct {
	Data struct {
		IpAddress            string `json:"ipAddress"`
		AbuseConfidenceScore int    `json:"abuseConfidenceScore"`
	} `json:"data"`
	Errors []struct {
		Detail string `json:"detail"`
		Status int    `json:"status"`
		Source struct {
			Parameter string `json:"parameter"`
		} `json:"source"`
	} `json:"errors"`
}
