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
	"github.com/wolfogre/funeypot/internal/pkg/ipapi"
	"github.com/wolfogre/funeypot/internal/pkg/logs"
)

type Server struct {
	username string
	password string

	db *model.Database

	static http.Handler
}

func NewServer(cfg config.Dashboard, db *model.Database) (*Server, error) {
	s, err := fs.Sub(static, "static")
	if err != nil {
		return nil, err
	}
	return &Server{
		username: cfg.Username,
		password: cfg.Password,
		db:       db,
		static:   http.FileServer(http.FS(s)),
	}, nil
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
	Ip        string  `json:"ip"`
	Count     int64   `json:"count"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
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

	pointM := map[string]struct{}{}
	var points []*responsePoint
	next := after
	if err := s.db.ScanSshAttempt(ctx, after, func(attempt *model.SshAttempt, geo *model.IpGeo) bool {
		if _, ok := pointM[attempt.Ip]; ok {
			return true
		}
		pointM[attempt.Ip] = struct{}{}
		if geo == nil {
			return true
		}
		points = append(points, &responsePoint{
			Ip:        attempt.Ip,
			Count:     attempt.Count,
			Latitude:  geo.Latitude,
			Longitude: geo.Longitude,
		})
		if attempt.UpdatedAt.After(next) {
			next = attempt.UpdatedAt
		}
		return true
	}); err != nil {
		logger.Errorf("scan ssh attempt: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

	resp, err := ipapi.Query(ctx, "")
	if err != nil {
		logger.Errorf("query ipapi: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(&responseGetSelf{
		Latitude:  resp.Lat,
		Longitude: resp.Lon,
	}); err != nil {
		logger.Errorf("encode response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
