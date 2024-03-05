// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package model

import (
	"context"
	"fmt"

	"github.com/funeypot/funeypot/internal/app/config"
	"github.com/funeypot/funeypot/internal/pkg/logs"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	db *gorm.DB
}

func NewDatabase(ctx context.Context, cfg config.Database) (*Database, error) {
	var (
		db  *gorm.DB
		err error
	)

	switch cfg.Driver {
	case "sqlite", "sqlite3":
		db, err = gorm.Open(sqlite.Open(cfg.Dsn), &gorm.Config{
			Logger: logs.GormLogger{},
		})
	case "postgres", "postgresql":
		db, err = gorm.Open(postgres.Open(cfg.Dsn), &gorm.Config{
			Logger: logs.GormLogger{},
		})
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	ret := &Database{
		db: db,
	}

	if err := ret.withContext(ctx).AutoMigrate(models...); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	return ret, nil
}

var models []any

func registerModel(model any) {
	for _, m := range models {
		if fmt.Sprintf("%T", m) == fmt.Sprintf("%T", model) {
			panic(fmt.Sprintf("model %T already registered", model))
		}
	}
	models = append(models, model)
}

func (db *Database) Create(ctx context.Context, value any) error {
	return db.withContext(ctx).
		Create(value).
		Error
}

func (db *Database) Save(ctx context.Context, value any) error {
	return db.withContext(ctx).
		Save(value).
		Error
}

func (db *Database) withContext(ctx context.Context) *gorm.DB {
	return db.db.WithContext(unwrapContext(ctx))
}

func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

// unwrapContext unwrap gin.Context to request context.
// Or it could cause data racing since gin will reuse gin.Context and sqlite will use it to do something in the background.
// See:
//
//	WARNING: DATA RACE
//	Write at 0x00c0000cc220 by goroutine 115:
//	  github.com/gin-gonic/gin.(*Engine).ServeHTTP()
//	      /home/runner/go/pkg/mod/github.com/gin-gonic/gin@v1.9.1/gin.go:573 +0xe7
//	  github.com/funeypot/funeypot/internal/app/dashboard.(*Server).ServeHTTP()
//	      /home/runner/work/funeypot/funeypot/internal/app/dashboard/dashboard.go:80 +0xd0d
//	  github.com/funeypot/funeypot/internal/app/server.(*HttpServer).ServeHTTP()
//	      /home/runner/work/funeypot/funeypot/internal/app/server/http_server.go:57 +0xc54
//	  net/http.serverHandler.ServeHTTP()
//	      /opt/hostedtoolcache/go/1.22.0/x64/src/net/http/server.go:3137 +0x2a1
//	  net/http.(*conn).serve()
//	      /opt/hostedtoolcache/go/1.22.0/x64/src/net/http/server.go:2039 +0x13c4
//	  net/http.(*Server).Serve.gowrap3()
//	      /opt/hostedtoolcache/go/1.22.0/x64/src/net/http/server.go:3285 +0x4f
//	Previous read at 0x00c0000cc220 by goroutine 111:
//	  github.com/gin-gonic/gin.(*Context).hasRequestContext()
//	      /home/runner/go/pkg/mod/github.com/gin-gonic/gin@v1.9.1/context.go:1186 +0xa7
//	  github.com/gin-gonic/gin.(*Context).Done()
//	      /home/runner/go/pkg/mod/github.com/gin-gonic/gin@v1.9.1/context.go:1200 +0x134
//	  github.com/glebarez/go-sqlite.interruptOnDone.func1()
//	      /home/runner/go/pkg/mod/github.com/glebarez/go-sqlite@v1.21.2/sqlite.go:760 +0x63
func unwrapContext(ctx context.Context) context.Context {
	if c, ok := ctx.(*gin.Context); ok {
		return c.Request.Context()
	}
	return ctx
}
