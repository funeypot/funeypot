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
	return attempt, db.withContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
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

func (db *Database) ScanBruteAttempt(ctx context.Context, updatedAfter time.Time, f func(attempt *BruteAttempt, geo *IpGeo) bool) error {
	sess := db.withContext(ctx)
	rows, err := sess.
		Model(&BruteAttempt{}).
		Select("brute_attempts.*, ip_geos.*").
		Joins("LEFT JOIN ip_geos ON brute_attempts.ip = ip_geos.ip").
		Where("brute_attempts.updated_at > ?", updatedAfter).
		Order("brute_attempts.updated_at").
		Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		result := &struct {
			*BruteAttempt
			*IpGeo
		}{}
		if err := sess.ScanRows(rows, result); err != nil {
			return err
		}
		if !f(result.BruteAttempt, result.IpGeo) {
			break
		}
	}
	return nil
}
