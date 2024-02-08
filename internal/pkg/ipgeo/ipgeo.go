// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package ipgeo

import (
	"context"
	"net"
)

type Info struct {
	Ip        net.IP
	Location  string
	Latitude  float64
	Longitude float64
}

type Querier interface {
	Query(ctx context.Context, ip string) (*Info, error)
}
