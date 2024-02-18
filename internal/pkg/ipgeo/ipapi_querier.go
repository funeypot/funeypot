// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package ipgeo

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
)

type IpapiComQuerier struct{}

var _ Querier = (*IpapiComQuerier)(nil)

func NewIpapiComQuerier() Querier {
	return &IpapiComQuerier{}
}

func (i *IpapiComQuerier) Query(ctx context.Context, ip string) (*Info, error) {
	result := &ipapiComQuerierResponse{}

	if netIp := net.ParseIP(ip); !netIp.IsGlobalUnicast() || netIp.IsPrivate() {
		result.Query = ip
		result.Country = "Reserved IP"
		return result.Info(), nil
	}

	resp, err := resty.NewWithClient(&http.Client{
		Transport: http.DefaultTransport,
	}).R().
		SetContext(ctx).
		SetResult(result).
		Get(fmt.Sprintf("http://ip-api.com/json/%s?fields=status,message,continent,continentCode,country,countryCode,region,regionName,city,district,zip,lat,lon,timezone,offset,currency,isp,org,as,asname,reverse,mobile,proxy,hosting,query", ip))
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("%d %s", resp.StatusCode(), resp.Status())
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("%s: %s", result.Status, result.Message)
	}

	return result.Info(), nil
}

type ipapiComQuerierResponse struct {
	Query         string  `json:"query"`
	Status        string  `json:"status"`
	Message       string  `json:"message"`
	Continent     string  `json:"continent"`
	ContinentCode string  `json:"continentCode"`
	Country       string  `json:"country"`
	CountryCode   string  `json:"countryCode"`
	Region        string  `json:"region"`
	RegionName    string  `json:"regionName"`
	City          string  `json:"city"`
	District      string  `json:"district"`
	Zip           string  `json:"zip"`
	Lat           float64 `json:"lat"`
	Lon           float64 `json:"lon"`
	Timezone      string  `json:"timezone"`
	Offset        int     `json:"offset"`
	Currency      string  `json:"currency"`
	Isp           string  `json:"isp"`
	Org           string  `json:"org"`
	As            string  `json:"as"`
	Asname        string  `json:"asname"`
	Reverse       string  `json:"reverse"`
	Mobile        bool    `json:"mobile"`
	Proxy         bool    `json:"proxy"`
	Hosting       bool    `json:"hosting"`
}

func (r *ipapiComQuerierResponse) Info() *Info {
	location := make([]string, 0, 4)
	if r.Country != "" {
		location = append(location, r.Country)
	}
	if r.RegionName != "" {
		location = append(location, r.RegionName)
	}
	if r.City != "" {
		location = append(location, r.City)
	}
	if r.District != "" {
		location = append(location, r.District)
	}

	return &Info{
		Ip:        net.ParseIP(r.Query),
		Location:  strings.Join(location, ", "),
		Latitude:  r.Lat,
		Longitude: r.Lon,
	}
}
