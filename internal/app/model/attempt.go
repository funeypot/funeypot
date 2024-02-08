// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package model

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func init() {
	registerModel(new(BruteAttempt))
}

//go:generate go run golang.org/x/tools/cmd/stringer -type BruteAttemptKind -linecomment -output attempt_kind.go
type BruteAttemptKind int

const (
	_                    BruteAttemptKind = iota // none
	BruteAttemptKindSsh                          // ssh
	BruteAttemptKindHttp                         // http
	BruteAttemptKindFtp                          // ftp
)

type BruteAttempt struct {
	Id            int64
	Ip            string           `gorm:"size:39,index:ip_kind"` // max ipv6 length
	Kind          BruteAttemptKind `gorm:"index:ip_kind"`
	User          string           `gorm:"size:255"`
	Password      string           `gorm:"size:255"`
	ClientVersion string           `gorm:"size:255"`
	StartedAt     time.Time
	StoppedAt     time.Time
	Count         int64

	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time `gorm:"index"`
}

func (r *BruteAttempt) Duration() time.Duration {
	return r.StoppedAt.Sub(r.StartedAt)
}

func (r *BruteAttempt) MaskedPassword() string {
	prefix := len(r.Password) / 3
	if prefix > 4 {
		prefix = 4
	}
	suffix := prefix

	return r.Password[:prefix] + strings.Repeat("*", len(r.Password)-prefix-suffix) + r.Password[len(r.Password)-suffix:]
}

func (r *BruteAttempt) ShortClientVersion() string {
	return strings.TrimPrefix(r.ClientVersion, "SSH-2.0-")
}

func (r *BruteAttempt) BeforeSave(_ *gorm.DB) error {
	r.User = truncateString(r.User, 255)
	r.Password = truncateString(r.Password, 255)
	r.ClientVersion = truncateString(r.ClientVersion, 255)
	return nil
}

func (db *Database) IncrBruteAttempt(
	ctx context.Context,
	ip string,
	kind BruteAttemptKind,
	timestamp time.Time,
	user, password, clientVersion string,
	after time.Time,
) (*BruteAttempt, error) {
	attempt := &BruteAttempt{}
	return attempt, db.db.Transaction(func(tx *gorm.DB) error {
		tx = tx.Clauses(clause.Locking{Strength: "UPDATE"})
		result := tx.
			WithContext(ctx).
			Where("ip = ? AND kind = ?", ip, kind).
			Last(attempt)
		if err := result.Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if errors.Is(result.Error, gorm.ErrRecordNotFound) || !attempt.StoppedAt.After(after) {
			attempt = &BruteAttempt{
				Ip:            ip,
				Kind:          kind,
				User:          user,
				Password:      password,
				ClientVersion: clientVersion,
				StartedAt:     timestamp,
				StoppedAt:     timestamp,
				Count:         1,
			}
			return tx.Create(attempt).Error
		}

		attempt.User = user
		attempt.Password = password
		attempt.ClientVersion = clientVersion
		attempt.StoppedAt = timestamp
		attempt.Count++
		return tx.Select("user", "password", "client_version", "stopped_at", "count").
			Updates(attempt).Error
	})
}

func (db *Database) FindBruteAttempt(ctx context.Context, updatedAfter time.Time) ([]*BruteAttempt, error) {
	sess := db.db.
		WithContext(ctx).
		Order("updated_at")
	if !updatedAfter.IsZero() {
		sess = sess.Where("updated_at > ?", updatedAfter)
	}

	var ret []*BruteAttempt
	return ret, sess.Find(&ret).Error
}
