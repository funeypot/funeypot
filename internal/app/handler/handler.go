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
	"gorm.io/gorm"
)

type Handler struct {
	delay        time.Duration
	abuseIpdbKey string
	db           *gorm.DB
	queue        chan *Request
}

func New(ctx context.Context, delay time.Duration, abuseIpdbKey string, db *gorm.DB) *Handler {
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
	ip = net.ParseIP(ip).String()

	logger := logs.From(ctx).With(
		"ip", ip,
		"remote_addr", request.RemoteAddr,
		"session_id", shortenSessionId(request.SessionId),
		"user", request.User,
		"password", request.Password,
		"client_version", request.ClientVersion,
	)

	attempt := &model.SshAttempt{}
	if err := h.db.Last(&attempt, "ip = ?", ip).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Errorf("get attempt: %v", err)
		return
	} else if time.Since(attempt.StoppedAt) > 24*time.Hour {
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
	logger = logger.With(
		"count", attempt.Count,
		"duration", attempt.Duration().String(),
	)

	logger.Infof("login")

	if attempt.Id == 0 {
		if err := h.db.Create(attempt).Error; err != nil {
			logger.Errorf("create attempt: %v", err)
		}
	} else {
		if err := h.db.Select("user", "password", "client_version", "stopped_at", "count").Updates(attempt).Error; err != nil {
			logger.Errorf("update attempt: %v", err)
		}
	}

	if h.abuseIpdbKey != "" && attempt.Count >= 5 {
		report := &model.AbuseipdbReport{}
		if err := h.db.Last(&report, "ip = ?", ip).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Errorf("get report: %v", err)
			return
		}
		if time.Since(report.ReportedAt) < 20*time.Minute {
			logger.Infof("already reported")
			return
		}
		score, err := h.reportRecord(ctx, attempt)
		if err != nil {
			logger.Errorf("report attempt: %v", err)
			return
		}
		logger.Infof("reported, score: %d", score)
		if !report.ReportedAt.IsZero() && report.Score != score {
			logger.Infof("score changed, %d -> %d", report.Score, score)
		}
		newReport := &model.AbuseipdbReport{
			Ip:         ip,
			ReportedAt: time.Now(),
			Score:      score,
		}
		if err := h.db.Create(newReport).Error; err != nil {
			logger.Errorf("create report: %v", err)
		}
	}

}

func (h *Handler) reportRecord(ctx context.Context, attempt *model.SshAttempt) (int, error) {
	data := url.Values{}
	data.Set("ip", attempt.Ip)
	data.Add("categories", "18,22")
	data.Add("timestamp", attempt.StoppedAt.Format(time.RFC3339))

	comment := fmt.Sprintf(
		"Funeypot detected %d attempts in %s. Last by user %q, password %q, client %q.",
		attempt.Count,
		attempt.Duration().Truncate(time.Second).String(),
		attempt.User,
		attempt.MaskedPassword(),
		attempt.ShortClientVersion(),
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

func (h *Handler) HandleHttp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logs.From(ctx)

	remoteAddr := r.Header.Get("X-Forwarded-For")
	logger = logger.With(
		"remote_addr", remoteAddr,
		"method", r.Method,
		"uri", r.RequestURI,
		"user_agent", r.UserAgent(),
	)

	w.Header().Set("WWW-Authenticate", `Basic realm="security" charset="UTF-8"`)
	w.Header().Add("WWW-Authenticate", "ApiKey")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	username, password, ok := r.BasicAuth()
	if ok {
		_, _ = fmt.Fprintf(w, `{"error":{"root_cause":[{"type":"security_exception","reason":"unable to authenticate user [%s] for REST request [%s]","header":{"WWW-Authenticate":["Basic realm=\"security\" charset=\"UTF-8\"","ApiKey"]}}],"type":"security_exception","reason":"unable to authenticate user [%s] for REST request [%s]","header":{"WWW-Authenticate":["Basic realm=\"security\" charset=\"UTF-8\"","ApiKey"]}},"status":401}`,
			username,
			r.RequestURI,
			username,
			r.RequestURI,
		)
		logger.With("username", username, "password", password).Infof("basic login")
	} else if auth := r.Header.Get("Authorization"); auth != "" {
		logger.Infof("other auth: %s", auth)
	} else {
		_, _ = fmt.Fprintf(w, `{"error":{"root_cause":[{"type":"security_exception","reason":"missing authentication credentials for REST request [%s]","header":{"WWW-Authenticate":["Basic realm=\"security\" charset=\"UTF-8\"","ApiKey"]}}],"type":"security_exception","reason":"missing authentication credentials for REST request [%s]","header":{"WWW-Authenticate":["Basic realm=\"security\" charset=\"UTF-8\"","ApiKey"]}},"status":401}`,
			r.RequestURI,
			r.RequestURI,
		)
		logger.Infof("no auth")
	}

	return
}
