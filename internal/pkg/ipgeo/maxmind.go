// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package ipgeo

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/oschwald/geoip2-golang"
)

type MaxmindQuerier struct {
	reader *geoip2.Reader
}

var _ Querier = (*MaxmindQuerier)(nil)

func NewMaxmindQuerier(file string) (*MaxmindQuerier, error) {
	file, err := releaseEmbed(file)
	if err != nil {
		return nil, fmt.Errorf("release embed: %w", err)
	}

	reader, err := geoip2.Open(file)
	if err != nil {
		return nil, fmt.Errorf("open maxmind db %q: %w", file, err)
	}
	return &MaxmindQuerier{reader: reader}, nil
}

func (q *MaxmindQuerier) Query(ctx context.Context, ip string) (*Info, error) {
	netIp := net.ParseIP(ip)
	if netIp == nil {
		return nil, fmt.Errorf("invalid ip address: %s", ip)
	}

	record, err := q.reader.City(netIp)
	if err != nil {
		return nil, fmt.Errorf("query maxmind db: %w", err)
	}

	var locations []string
	if v := record.Country.Names["en"]; v != "" {
		locations = append(locations, v)
	}
	for _, subdivision := range record.Subdivisions {
		if v := subdivision.Names["en"]; v != "" {
			locations = append(locations, v)
		}
	}
	if v := record.City.Names["en"]; v != "" {
		locations = append(locations, v)
	}

	return &Info{
		Ip:        netIp,
		Location:  strings.Join(locations, ", "),
		Latitude:  record.Location.Latitude,
		Longitude: record.Location.Longitude,
	}, nil
}

func releaseEmbed(file string) (string, error) {
	embed := len(geoLite2City) > 0

	tempDir := filepath.Join(os.TempDir(), "funeypot")
	tempFile := filepath.Join(tempDir, "GeoLite2-City.mmdb")

	if len(geoLite2City) > 1 {
		// release the data to tempFile
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			return "", fmt.Errorf("create tempDir %q: %w", tempDir, err)
		}
		if err := os.WriteFile(tempFile, geoLite2City, 0644); err != nil {
			return "", fmt.Errorf("write tempFile %q: %w", tempFile, err)
		}
		// free the memory, keep 1 byte to indicate the release
		geoLite2City = []byte{0}
	}

	if file == "embed" {
		if embed {
			return tempFile, nil
		}
		return "", fmt.Errorf("you are running a version without the embedded ip geo database, " +
			"please use a different version or provide the path to the database file")
	}

	return file, nil
}
