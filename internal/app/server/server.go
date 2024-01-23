package server

import (
	"context"
)

type Server interface {
	Startup(ctx context.Context, cancel context.CancelFunc)
	Shutdown(ctx context.Context) error
}
