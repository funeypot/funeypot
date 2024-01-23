package model

import (
	"time"

	"github.com/wolfogre/funeypot/internal/pkg/ipapi"
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
