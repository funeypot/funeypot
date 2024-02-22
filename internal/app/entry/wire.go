// Copyright 2024 The Funeypot Authors. All rights reserved.
// SPDX-License-Identifier: MIT

//go:build wireinject

package entry

import (
	"context"

	"github.com/funeypot/funeypot/internal/app/config"

	"github.com/google/wire"
)

//go:generate go run -mod=mod github.com/google/wire/cmd/wire

func NewEntrypoint(ctx context.Context, cfg *config.Config) (*Entrypoint, error) {
	panic(wire.Build(providerSet))
}
