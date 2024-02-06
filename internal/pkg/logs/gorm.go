// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package logs

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type GormLogger struct{}

func (l GormLogger) getLogger(ctx context.Context) Logger {
	return From(ctx)
}

func (l GormLogger) LogMode(_ logger.LogLevel) logger.Interface {
	return l
}

func (l GormLogger) Info(ctx context.Context, s string, i ...interface{}) {
	l.getLogger(ctx).Infof(s, i...)
}

func (l GormLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	l.getLogger(ctx).Warnf(s, i...)
}

func (l GormLogger) Error(ctx context.Context, s string, i ...interface{}) {
	l.getLogger(ctx).Errorf(s, i...)
}

func (l GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	lgr := l.getLogger(ctx)

	duration := time.Since(begin)
	lgr = lgr.With("duration", duration.String())
	sql, rows := fc()
	lgr = lgr.With("rows", rows)

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		lgr.Errorf("%v: %s", err, sql)
	case duration > time.Second:
		lgr.Warnf("slow: %s", sql)
	default:
		lgr.Debug(sql)
	}
}

var _ logger.Interface = (*GormLogger)(nil)
