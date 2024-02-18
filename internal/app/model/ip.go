// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package model

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/wolfogre/funeypot/internal/pkg/ipgeo"

	"gorm.io/gorm"
)

func init() {
	registerModel(new(IpGeo))
}

type IpGeo struct {
	Ip        string `gorm:"primaryKey; size:39"`
	Location  string `gorm:"size:255"`
	Latitude  float64
	Longitude float64

	CreatedAt time.Time `gorm:"<-:create"`
	UpdatedAt time.Time
}

func (m *IpGeo) FillInfo(r *ipgeo.Info) *IpGeo {
	m.Ip = r.Ip.String()
	m.Location = r.Location
	m.Latitude = r.Latitude
	m.Longitude = r.Longitude

	return m
}

func (m *IpGeo) Info() *ipgeo.Info {
	return &ipgeo.Info{
		Ip:        net.ParseIP(m.Ip),
		Location:  m.Location,
		Latitude:  m.Latitude,
		Longitude: m.Longitude,
	}
}

func (m *IpGeo) BeforeSave(_ *gorm.DB) error {
	m.Location = truncateString(m.Location, 255)
	return nil
}

func (db *Database) TakeIpGeo(ctx context.Context, ip string) (*IpGeo, bool, error) {
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

func NewCachedIpGeoQuerier(querier ipgeo.Querier, db *Database) ipgeo.Querier {
	return ipgeo.NewCachedQuerier(querier,
		func(ctx context.Context, ip string) (*ipgeo.Info, bool, error) {
			geo, ok, err := db.TakeIpGeo(ctx, ip)
			if err != nil {
				return nil, false, fmt.Errorf("get ip geo: %w", err)
			}
			if ok && time.Since(geo.CreatedAt) < 24*time.Hour {
				return geo.Info(), true, nil
			}
			return nil, false, nil
		},
		func(ctx context.Context, ip string, info *ipgeo.Info) error {
			return db.Save(ctx, (&IpGeo{
				Ip: ip,
			}).FillInfo(info))
		})
}
