// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package dashboard

import (
	"crypto/subtle"
	"embed"
	"io/fs"
	"net/http"
	"strconv"
	"time"

	"github.com/funeypot/funeypot/internal/app/config"
	"github.com/funeypot/funeypot/internal/app/model"
	"github.com/funeypot/funeypot/internal/pkg/ipgeo"
	"github.com/funeypot/funeypot/internal/pkg/logs"
	"github.com/funeypot/funeypot/internal/pkg/selfip"

	"github.com/gin-gonic/gin"
)

type Server struct {
	username string
	password string

	db           *model.Database
	ipgeoQuerier ipgeo.Querier

	engine *gin.Engine
}

func NewServer(cfg config.Dashboard, db *model.Database, ipgeoQuerier ipgeo.Querier) (*Server, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	server := &Server{
		username:     cfg.Username,
		password:     cfg.Password,
		db:           db,
		ipgeoQuerier: ipgeoQuerier,
	}

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.ContextWithFallback = true

	apiGroup := engine.Group("/api/v1")
	apiGroup.GET("/points", server.handleGetPoints)
	apiGroup.GET("/self", server.handleGetSelf)

	staticFs, err := fs.Sub(static, "static")
	if err != nil {
		return nil, err
	}
	httpFs := http.FS(staticFs)
	engine.StaticFileFS("/", "/", httpFs)
	engine.StaticFS("/static", httpFs)

	server.engine = engine

	return server, nil
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

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.engine.ServeHTTP(w, r)
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

func (s *Server) handleGetPoints(c *gin.Context) {
	logger := logs.From(c)

	afterI, _ := strconv.ParseInt(c.Query("after"), 10, 64)
	after := time.Unix(afterI, 0)
	if afterI == 0 {
		after = time.Now().AddDate(0, 0, -30) // TODO: make default range configurable
	}

	pointM := map[string]struct{}{}
	var points []*responsePoint
	next := after
	if err := s.db.ScanBruteAttempt(c, after, func(attempt *model.BruteAttempt, geo *model.IpGeo) bool {
		if _, ok := pointM[attempt.Ip]; ok {
			return true
		}
		pointM[attempt.Ip] = struct{}{}
		if geo == nil {
			// ignore attempts from unknown ip location
			return true
		}
		point := &responsePoint{
			Ip:          attempt.Ip,
			Count:       attempt.Count,
			Latitude:    geo.Latitude,
			Longitude:   geo.Longitude,
			ActivatedAt: attempt.StoppedAt,
		}
		points = append(points, point)
		if attempt.UpdatedAt.After(next) {
			next = attempt.UpdatedAt
		}
		return true
	}); err != nil {
		logger.Errorf("scan ssh attempt: %v", err)
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, &responseGetPoints{
		Points: points,
		Next:   next.Unix(),
	})
}

type responseGetSelf struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (s *Server) handleGetSelf(c *gin.Context) {
	logger := logs.From(c)

	selfIp, err := selfip.Get(c)
	if err != nil {
		logger.Errorf("get self ip: %v", err)
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	geo, err := s.ipgeoQuerier.Query(c, selfIp)
	if err != nil {
		logger.Errorf("query geo: %v", err)
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, &responseGetSelf{
		Latitude:  geo.Latitude,
		Longitude: geo.Longitude,
	})
}
