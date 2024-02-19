// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package model

import (
	"context"
	"fmt"

	"github.com/funeypot/funeypot/internal/app/config"
	"github.com/funeypot/funeypot/internal/pkg/logs"

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

	if err := db.WithContext(ctx).AutoMigrate(models...); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	return &Database{
		db: db,
	}, nil
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
	return db.db.WithContext(ctx).
		Create(value).
		Error
}

func (db *Database) Save(ctx context.Context, value any) error {
	return db.db.WithContext(ctx).
		Save(value).
		Error
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
