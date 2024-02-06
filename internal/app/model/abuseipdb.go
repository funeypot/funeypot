// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package model

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

func init() {
	registerModel(new(AbuseipdbReport))
}

type AbuseipdbReport struct {
	Id         int64
	Ip         string `gorm:"size:39"`
	ReportedAt time.Time
	Score      int

	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time
}

func (db *Database) LastAbuseipdbReport(ctx context.Context, ip string) (*AbuseipdbReport, bool, error) {
	report := &AbuseipdbReport{}
	result := db.db.
		WithContext(ctx).
		Last(&report, "ip = ?", ip)
	if err := result.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return report, true, nil
}
