// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package ipgeo

import (
	"context"

	"github.com/wolfogre/funeypot/internal/pkg/logs"
)

type CachedQuerier struct {
	inner  Querier
	getter func(ctx context.Context, ip string) (*Info, bool, error)
	setter func(ctx context.Context, ip string, info *Info) error
}

var _ Querier = (*CachedQuerier)(nil)

func NewCachedQuerier(inner Querier, getter func(ctx context.Context, ip string) (*Info, bool, error), setter func(ctx context.Context, ip string, info *Info) error) *CachedQuerier {
	return &CachedQuerier{
		inner:  inner,
		getter: getter,
		setter: setter,
	}
}

func (c *CachedQuerier) Query(ctx context.Context, ip string) (*Info, error) {
	info, ok, err := c.getter(ctx, ip)
	if err != nil {
		return nil, err
	}
	if ok {
		return info, nil
	}

	info, err = c.inner.Query(ctx, ip)
	if err != nil {
		return nil, err
	}

	if err := c.setter(ctx, ip, info); err != nil {
		logs.From(ctx).Errorf("set info for ip %s: %v", ip, err)
		// go on, don't block the main flow
	}

	return info, nil
}
