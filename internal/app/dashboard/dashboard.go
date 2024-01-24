package dashboard

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/wolfogre/funeypot/internal/app/config"
	"github.com/wolfogre/funeypot/internal/app/model"
	"github.com/wolfogre/funeypot/internal/pkg/logs"
)

type Server struct {
	username string
	password string

	db *model.Database
}

func NewServer(cfg config.Dashboard, db *model.Database) *Server {
	return &Server{
		username: cfg.Username,
		password: cfg.Password,
		db:       db,
	}
}

type apiResponse struct {
	Points []*apiPoint `json:"points"`
	Next   int64       `json:"next"`
}

type apiPoint struct {
	Ip        string  `json:"ip"`
	Count     int64   `json:"count"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodGet || r.URL.Path != "/api/points" {
		http.NotFound(w, r)
		return
	}

	ctx := r.Context()
	logger := logs.From(ctx)

	afterQ := r.URL.Query().Get("after")
	afterI, _ := strconv.ParseInt(afterQ, 10, 64)
	after := time.Unix(afterI, 0)

	pointM := map[string]struct{}{}
	var points []*apiPoint
	next := after
	if err := s.db.ScanSshAttempt(ctx, after, func(attempt *model.SshAttempt, geo *model.IpGeo) bool {
		if _, ok := pointM[attempt.Ip]; ok {
			return true
		}
		pointM[attempt.Ip] = struct{}{}
		if geo == nil {
			return true
		}
		points = append(points, &apiPoint{
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

	response := &apiResponse{
		Points: points,
		Next:   next.Unix(),
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errorf("encode response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
