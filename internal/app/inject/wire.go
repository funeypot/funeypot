// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

//go:build wireinject

package inject

import (
	"context"

	"github.com/wolfogre/funeypot/internal/app/config"

	"github.com/google/wire"
)

//go:generate go run github.com/google/wire/cmd/wire gen

func NewEntrypoint(ctx context.Context, cfg *config.Config) (*Entrypoint, error) {
	panic(wire.Build(providerSet))
}
