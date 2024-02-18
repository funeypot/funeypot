// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package dashboard

import (
	"crypto/subtle"
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"strconv"
	"time"

	"github.com/wolfogre/funeypot/internal/app/config"
	"github.com/wolfogre/funeypot/internal/app/model"
	"github.com/wolfogre/funeypot/internal/pkg/ipgeo"
	"github.com/wolfogre/funeypot/internal/pkg/logs"
	"github.com/wolfogre/funeypot/internal/pkg/selfip"
)

type Server struct {
	username string
	password string

	db           *model.Database
	ipgeoQuerier ipgeo.Querier

	static http.Handler
}

func NewServer(cfg config.Dashboard, db *model.Database, ipgeoQuerier ipgeo.Querier) (*Server, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	s, err := fs.Sub(static, "static")
	if err != nil {
		return nil, err
	}
	return &Server{
		username:     cfg.Username,
		password:     cfg.Password,
		db:           db,
		ipgeoQuerier: ipgeoQuerier,
		static:       http.FileServer(http.FS(s)),
	}, nil
}

func (s *Server) Enabled() bool {
	return s != nil
}

func (s *Server) Verify(username, password string) bool {
	return subtle.ConstantTimeCompare([]byte(s.username), []byte(username)) == 1 &&
		subtle.ConstantTimeCompare([]byte(s.password), []byte(password)) == 1
}

//go:embed static
var static embed.FS

func (s *Server) Handle(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/v1/points":
		switch r.Method {
		case http.MethodGet:
			s.handleGetPoints(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	case "/api/v1/self":
		switch r.Method {
		case http.MethodGet:
			s.handleGetSelf(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		s.static.ServeHTTP(w, r)
	}
}

type responsePoint struct {
	Ip          string    `json:"ip"`
	Count       int64     `json:"count"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	ActivatedAt time.Time `json:"activated_at"`
}

type responseGetPoints struct {
	Points []*responsePoint `json:"points"`
	Next   int64            `json:"next"`
}

func (s *Server) handleGetPoints(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logs.From(ctx)

	afterQ := r.URL.Query().Get("after")
	afterI, _ := strconv.ParseInt(afterQ, 10, 64)
	after := time.Unix(afterI, 0)
	if afterI == 0 {
		after = time.Now().AddDate(0, 0, -30) // TODO: make default range configurable
	}

	next := after
	attempts, err := s.db.FindBruteAttempt(ctx, after, 100) // TODO: make limit configurable
	if err != nil {
		logger.Errorf("find ssh attempt: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	points := make([]*responsePoint, 0, len(attempts))
	pointM := make(map[string]struct{}, len(attempts))

	for _, attempt := range attempts {
		if _, ok := pointM[attempt.Ip]; ok {
			continue
		}
		pointM[attempt.Ip] = struct{}{}
		geo, err := s.ipgeoQuerier.Query(ctx, attempt.Ip)
		if err != nil {
			logger.Errorf("query geo: %v", err)
			continue
		}
		points = append(points, &responsePoint{
			Ip:          attempt.Ip,
			Count:       attempt.Count,
			Latitude:    geo.Latitude,
			Longitude:   geo.Longitude,
			ActivatedAt: attempt.StoppedAt,
		})
		if attempt.UpdatedAt.After(next) {
			next = attempt.UpdatedAt
		}
	}

	if err := json.NewEncoder(w).Encode(&responseGetPoints{
		Points: points,
		Next:   next.Unix(),
	}); err != nil {
		logger.Errorf("encode response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type responseGetSelf struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (s *Server) handleGetSelf(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logs.From(ctx)

	selfIp, err := selfip.Get(ctx)
	if err != nil {
		logger.Errorf("get self ip: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	geo, err := s.ipgeoQuerier.Query(ctx, selfIp)
	if err != nil {
		logger.Errorf("query geo: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(&responseGetSelf{
		Latitude:  geo.Latitude,
		Longitude: geo.Longitude,
	}); err != nil {
		logger.Errorf("encode response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
