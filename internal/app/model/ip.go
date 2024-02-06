// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package model

import (
	"context"
	"errors"
	"time"

	"github.com/wolfogre/funeypot/internal/pkg/ipapi"

	"gorm.io/gorm"
)

func init() {
	registerModel(new(IpGeo))
}

type IpGeo struct {
	Ip          string `gorm:"primaryKey; size:39"`
	CountryCode string `gorm:"size:2"`
	Location    string `gorm:"size:255"`
	Latitude    float64
	Longitude   float64
	Isp         string `gorm:"size:255"`

	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time
}

func (m *IpGeo) FillIpapiResponse(r *ipapi.Response) *IpGeo {
	m.CountryCode = r.CountryCode
	m.Location = r.Location()
	m.Latitude = r.Lat
	m.Longitude = r.Lon
	m.Isp = r.Isp

	return m
}

func (m *IpGeo) BeforeSave(_ *gorm.DB) error {
	m.Location = truncateString(m.Location, 255)
	m.Isp = truncateString(m.Isp, 255)
	return nil
}

func (db *Database) TaskIpGeo(ctx context.Context, ip string) (*IpGeo, bool, error) {
	geo := &IpGeo{}
	result := db.db.
		WithContext(ctx).
		Take(&geo, "ip = ?", ip)
	if err := result.Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return geo, true, nil
}
