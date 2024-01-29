//go:build wireinject

package inject

import (
	"context"

	"github.com/google/wire"
)

//go:generate go run github.com/google/wire/cmd/wire gen

func NewEntrypoint(ctx context.Context, configFile string) (*Entrypoint, error) {
	panic(wire.Build(providerSet))
}
